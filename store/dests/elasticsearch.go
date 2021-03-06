package dests

import (
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
	"github.com/olivere/elastic"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/encoders"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/utils"
	"github.com/stephane-martin/skewer/utils/eerrors"
	"github.com/stephane-martin/skewer/utils/es"
	"github.com/valyala/bytebufferpool"
	"github.com/zond/gotomic"
)

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
	sentMessagesUids  *gotomic.Hash
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
		sentMessagesUids:  gotomic.NewHash(),
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
	resp, err := d.elasticClient.ClusterHealth().Level("cluster").Do(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.Status != "green" && resp.Status != "yellow" {
		connCounter.WithLabelValues("elasticsearch", "fail").Inc()
		return nil, eerrors.New("Elasticsearch cluster had red status")
	}

	names, err := d.elasticClient.IndexNames()
	if err != nil {
		connCounter.WithLabelValues("elasticsearch", "fail").Inc()
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
		connCounter.WithLabelValues("elasticsearch", "fail").Inc()
		return nil, err
	}

	if config.Rebind > 0 {
		go func() {
			select {
			case <-ctx.Done():
				// the store service asked for stop
			case <-time.After(config.Rebind):
				d.dofatal(eerrors.Errorf("Rebind period has expired (%s)", config.Rebind.String()))
			}
		}()
	}

	return d, nil
}

func (d *ElasticDestination) getClient() (*elastic.Client, error) {
	return es.GetClient(d.config, d.httpClient, d.logger)
}

func (d *ElasticDestination) after(execID int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if response == nil {
		d.dofatal(eerrors.New("BUG: response in ElasticDestination.after is NIL"))
		return
	}
	successes := response.Succeeded()
	failures := response.Failed()
	var item *elastic.BulkResponseItem
	var uid utils.MyULID
	var e error
	for _, item = range successes {
		uid, e = utils.ParseMyULID(item.Id)
		if e != nil {
			continue
		}
		d.sentMessagesUids.Delete(uid)
		d.ACK(uid)
	}
	if len(failures) == 0 {
		return
	}

	for _, item = range failures {
		uid, e = utils.ParseMyULID(item.Id)
		if e != nil {
			continue
		}
		d.sentMessagesUids.Delete(uid)
		d.NACK(uid)
		if item.Error != nil {
			d.logger.Warn("Elasticsearch index error", "type", item.Error.Type, "reason", item.Error.Reason, "index", item.Error.Index)
		}
	}
	d.dofatal(eerrors.New("Elasticsearch bulk delivery error"))
}

func (d *ElasticDestination) Close() error {
	d.sentMessagesUids.Each(func(k gotomic.Hashable, v gotomic.Thing) bool {
		if uid, ok := k.(utils.MyULID); ok {
			d.NACK(uid)
		}
		return false
	})
	return d.processor.Close() // processor.Close() must only be called after we are sure that no more Send() will happen... or panic.
}

func (d *ElasticDestination) sendOne(ctx context.Context, msg *model.FullMessage) (err error) {
	// compute the destination index name
	indexBuf := bytebufferpool.Get()
	err = d.indexNameTpl.Execute(indexBuf, msg.Fields)
	if err != nil {
		bytebufferpool.Put(indexBuf)
		return encoders.EncodingError(err)
	}
	indexName := indexBuf.String()
	bytebufferpool.Put(indexBuf)

	// serialize the message
	var buf string
	buf, err = encoders.ChainEncode(d.encoder, msg)
	if err != nil {
		return encoders.EncodingError(err)
	}

	// create index in ES if needed
	if d.config.CreateIndices && !d.knownIndexNames.Has(indexName) {
		d.logger.Info("Index does not exist yet in Elasticsearch", "name", indexName)
		client, err := d.getClient()
		if err != nil {
			return err
		}
		// refresh index names
		names, err := client.IndexNames()
		if err != nil {
			return err
		}
		d.knownIndexNames.Clear()
		for _, name := range names {
			d.knownIndexNames.Add(name)
		}
		if !d.knownIndexNames.Has(indexName) {
			res, err := client.CreateIndex(indexName).BodyString(d.createOptionsBody).Do(context.Background())
			if err != nil {
				return err
			}
			if !res.Acknowledged {
				return fmt.Errorf("Index creation not acknowledged")
			}
			d.knownIndexNames.Add(indexName)
			d.logger.Info("Created new index in Elasticsearch", "name", indexName)
		}
	}

	// add message to the bulk processor work list
	d.sentMessagesUids.Put(msg.Uid, true)
	d.processor.Add(
		elastic.NewBulkIndexRequest().Index(indexName).Type(d.messagesType).Id(msg.Uid.String()).Doc(json.RawMessage(buf)),
	)

	return nil
}

func (d *ElasticDestination) Send(ctx context.Context, msgs []model.OutputMsg) (err eerrors.ErrorSlice) {
	return d.ForEach(ctx, d.sendOne, false, true, msgs)
}
