package dests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/set"
	"github.com/inconshreveable/log15"
	"github.com/olivere/elastic"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/model/encoders"
	"github.com/stephane-martin/skewer/utils"
	"github.com/stephane-martin/skewer/utils/es"
)

func getClient(config conf.ElasticDestConfig, httpClient *http.Client, logger log15.Logger) (c *elastic.Client, err error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	elasticOpts := []elastic.ClientOptionFunc{}
	elasticOpts = append(elasticOpts, elastic.SetURL(config.URLs...))
	elasticOpts = append(elasticOpts, elastic.SetHttpClient(httpClient))
	if logger != nil {
		elasticOpts = append(elasticOpts, elastic.SetErrorLog(&es.ESLogger{Logger: logger}))
	}
	elasticOpts = append(elasticOpts, elastic.SetHealthcheck(config.HealthCheck))
	elasticOpts = append(elasticOpts, elastic.SetSniff(config.Sniffing))
	if config.HealthCheck {
		elasticOpts = append(elasticOpts, elastic.SetHealthcheckTimeout(config.HealthCheckTimeout))
		elasticOpts = append(elasticOpts, elastic.SetHealthcheckTimeoutStartup(config.HealthCheckTimeoutStartup))
		elasticOpts = append(elasticOpts, elastic.SetHealthcheckInterval(config.HealthCheckInterval))
	}

	if config.TLSEnabled {
		elasticOpts = append(elasticOpts, elastic.SetScheme("https"))
	}
	if len(config.Username) > 0 && len(config.Password) > 0 {
		elasticOpts = append(elasticOpts, elastic.SetBasicAuth(config.Username, config.Password))
	}

	c, err = elastic.NewClient(elasticOpts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ElasticDestination struct {
	*baseDestination
	elasticClient     *elastic.Client
	httpClient        *http.Client
	processor         *elastic.BulkProcessor
	indexNameTpl      *template.Template
	messagesType      string
	config            conf.ElasticDestConfig
	knownIndexNames   set.Interface
	createOptionsBody string
}

func NewElasticDestination(ctx context.Context, e *Env) (Destination, error) {
	config := e.config.ElasticDest
	if len(config.URLs) == 0 {
		config.URLs = []string{"http://127.0.0.1:9200"}
	}
	d := &ElasticDestination{
		baseDestination:   newBaseDestination(conf.Elasticsearch, "elasticsearch", e),
		messagesType:      config.MessagesType,
		knownIndexNames:   set.New(set.ThreadSafe),
		createOptionsBody: es.NewOpts(config.NShards, config.NReplicas, config.CheckStartup, config.RefreshInterval).Marshal(),
	}
	var err error
	d.indexNameTpl, err = template.New("index").Parse(config.IndexNameTpl)
	if err != nil {
		return nil, err
	}
	err = d.setFormat(config.Format)
	if err != nil {
		return nil, err
	}

	config.ProxyURL = strings.TrimSpace(config.ProxyURL)

	if strings.HasPrefix(strings.ToLower(config.URLs[0]), "https") {
		config.TLSEnabled = true
	}
	dialer := &net.Dialer{
		Timeout: config.ConnTimeout,
	}
	if config.ConnKeepAlive {
		dialer.KeepAlive = config.ConnKeepAlivePeriod
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		Proxy:                 nil,
		MaxIdleConns:          100,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		DialContext:           dialer.DialContext,
	}

	if config.TLSEnabled {
		tlsconfig, err := utils.NewTLSConfig(
			"",
			config.CAFile,
			config.CAPath,
			config.CertFile,
			config.KeyFile,
			config.Insecure,
			e.confined,
		)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsconfig
	}

	if len(config.ProxyURL) > 0 {
		url, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(url)
	}

	d.httpClient = &http.Client{
		Transport: transport,
		Jar:       nil,
	}

	d.config = config
	d.elasticClient, err = d.getClient()
	if err != nil {
		return nil, err
	}

	names, err := d.elasticClient.IndexNames()
	if err != nil {
		return nil, err
	}
	d.logger.Info("Existing indices in Elasticsearch", "names", strings.Join(names, ","))
	for _, name := range names {
		d.knownIndexNames.Add(name)
	}

	processor := d.elasticClient.BulkProcessor().
		Name("SkewerWorker").
		Workers(http.DefaultMaxIdleConnsPerHost).
		BulkActions(config.BatchSize).
		Stats(true).
		FlushInterval(config.FlushPeriod).
		After(d.after).
		Backoff(elastic.StopBackoff{})

	d.processor, err = processor.Do(context.Background())
	if err != nil {
		return nil, err
	}

	if config.Rebind > 0 {
		go func() {
			select {
			case <-ctx.Done():
				// the store service asked for stop
			case <-time.After(config.Rebind):
				e.logger.Info("HTTP destination rebind period has expired", "rebind", config.Rebind.String())
				d.dofatal()
			}
		}()
	}

	return d, nil
}

func (d *ElasticDestination) getClient() (*elastic.Client, error) {
	return getClient(d.config, d.httpClient, d.logger)
}

func (d *ElasticDestination) after(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if response == nil {
		d.dofatal()
		return
	}
	successes := response.Succeeded()
	failures := response.Failed()
	var item *elastic.BulkResponseItem
	var uid utils.MyULID
	var e error
	for _, item = range successes {
		uid, e = utils.Parse(item.Id)
		if e != nil {
			continue
		}
		d.ack(uid)
	}
	if len(failures) == 0 {
		return
	}

	for _, item = range failures {
		uid, e = utils.Parse(item.Id)
		if e != nil {
			continue
		}
		d.nack(uid)
		if item.Error != nil {
			d.logger.Warn("Elasticsearch index error", "type", item.Error.Type, "reason", item.Error.Reason, "index", item.Error.Index)
		}
	}
	d.dofatal()
}

func (d *ElasticDestination) Close() error {
	return d.processor.Close()
}

func (d *ElasticDestination) Send(msg *model.FullMessage, partitionKey string, partitionNumber int32, topic string) (err error) {
	indexBuf := bytes.NewBuffer(nil)
	err = d.indexNameTpl.Execute(indexBuf, msg.Fields)
	if err != nil {
		d.permerr(msg.Uid)
		return err
	}
	// create index in ES if needed
	indexName := indexBuf.String()
	if d.config.CreateIndices && !d.knownIndexNames.Has(indexName) {
		d.logger.Info("Index does not exist yet in Elasticsearch", "name", indexName)
		client, err := d.getClient()
		if err != nil {
			d.nack(msg.Uid)
			d.dofatal()
			return err
		}
		// refresh index names
		names, err := client.IndexNames()
		if err != nil {
			d.nack(msg.Uid)
			d.dofatal()
			return err
		}
		d.knownIndexNames.Clear()
		for _, name := range names {
			d.knownIndexNames.Add(name)
		}
		if !d.knownIndexNames.Has(indexName) {
			res, err := client.CreateIndex(indexName).BodyString(d.createOptionsBody).Do(context.Background())
			if err != nil {
				d.nack(msg.Uid)
				return err
			}
			if !res.Acknowledged {
				d.nack(msg.Uid)
				return fmt.Errorf("Index creation not acknowledged")
			}
			d.knownIndexNames.Add(indexName)
			d.logger.Info("Created new index in Elasticsearch", "name", indexName)
		}
	}

	// add message to the bulk processor work list
	var buf json.RawMessage
	buf, err = encoders.ChainEncode(d.encoder, msg)
	if err != nil {
		d.permerr(msg.Uid)
		return err
	}
	d.processor.Add(
		elastic.NewBulkIndexRequest().Index(indexName).Type(d.messagesType).Id(msg.Uid.String()).Doc(buf),
	)
	return nil
}