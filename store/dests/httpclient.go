package dests

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	circuit "github.com/rubyist/circuitbreaker"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/encoders"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/utils"
	"github.com/stephane-martin/skewer/utils/eerrors"
	"github.com/stephane-martin/skewer/utils/queue/defered"
	"github.com/valyala/bytebufferpool"
)

type HTTPDestination struct {
	*baseDestination
	clt         *http.Client
	breaker     *circuit.Breaker
	username    string
	password    string
	useragent   string
	url         *template.Template
	method      string
	contentType string
	reqtimeout  time.Duration
	queue       *defered.Ring
	wg          sync.WaitGroup
}

func NewHTTPDestination(ctx context.Context, e *Env) (Destination, error) {
	config := e.config.HTTPDest
	d := &HTTPDestination{
		baseDestination: newBaseDestination(conf.HTTP, "http", e),
		useragent:       config.UserAgent,
		method:          config.Method,
		reqtimeout:      config.RequestTimeout,
	}
	err := d.setFormat(config.Format)
	if err != nil {
		return nil, err
	}

	config.ContentType = strings.TrimSpace(strings.ToLower(config.ContentType))
	d.contentType = config.ContentType
	if config.ContentType == "auto" || config.ContentType == "" {
		if encoders.MimeTypes[d.format] != "" {
			d.contentType = encoders.MimeTypes[d.format]
		} else {
			d.contentType = "text/plain"
		}
	}

	if config.BasicAuth {
		d.username = config.Username
		d.password = config.Password
	}

	config.URL = strings.TrimSpace(config.URL)
	config.ProxyURL = strings.TrimSpace(config.ProxyURL)

	zurl, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	host := zurl.Host

	tmpl, err := template.New("url").Parse(config.URL)
	if err != nil {
		return nil, err
	}
	d.url = tmpl

	if strings.HasPrefix(strings.ToLower(config.URL), "https") {
		config.TLSEnabled = true
	}
	dialer := &net.Dialer{
		Timeout: config.ConnTimeout,
	}
	if config.ConnKeepAlive {
		dialer.KeepAlive = config.ConnKeepAlivePeriod
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		Proxy:                 nil,
		MaxIdleConns:          100,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		DialContext:           dialer.DialContext,
	}

	if config.TLSEnabled {
		tlsconfig, err := utils.NewTLSConfig(
			host,
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

	d.clt = &http.Client{
		Transport: transport,
		Jar:       nil,
	}

	// try to send a HEAD request
	urlbuf := bytes.NewBuffer(nil)
	err = d.url.Execute(urlbuf, &model.SyslogMessage{})
	if err == nil {
		_, err = d.clt.Head(urlbuf.String())
		if eerrors.HasConnRefused(err) || !eerrors.IsTemporary(err) {
			connCounter.WithLabelValues("http", "fail").Inc()
			return nil, err
		}
	}

	d.queue = defered.NewRing(4 * uint64(config.MaxIdleConnsPerHost))
	// TODO: why 5 ?
	d.breaker = circuit.NewConsecutiveBreaker(5)

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

	for i := 0; i < config.MaxIdleConnsPerHost; i++ {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			err := d.dequeue(ctx)
			if err != nil {
				d.dofatal(eerrors.Wrap(err, "Error performing HTTP request"))
			}
		}()
	}

	return d, nil
}

func (d *HTTPDestination) Close() error {
	d.queue.Dispose()
	d.wg.Wait()
	// nack remaining enqueued requests
	for {
		req, err := d.queue.Get()
		if err != nil || req == nil {
			break
		}
		d.NACK(req.UID)
	}
	return nil
}

func (d *HTTPDestination) doHTTP(ctx context.Context, uid utils.MyULID, req *http.Request) (err error) {

	req.Header.Set("Content-Type", d.contentType)
	if len(d.useragent) > 0 {
		req.Header.Set("User-Agent", d.useragent)
	}
	if len(d.username) > 0 && len(d.password) > 0 {
		req.SetBasicAuth(d.username, d.password)
	}
	req = req.WithContext(ctx)

	// perform the HTTP request, retry every second, use a circuit breaker to limit the tries
	var resp *http.Response
	for {
		err = d.breaker.CallContext(
			ctx, func() (e error) {
				resp, e = d.clt.Do(req.WithContext(ctx))
				return e
			},
			d.reqtimeout,
		)
		if err == nil {
			break
		}
		if err == context.Canceled {
			return nil
		}
		if eerrors.HasConnRefused(err) {
			// we stop if there is not even a HTTP server listening
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}

	// not interested in response body
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	httpStatusCounter.WithLabelValues(req.Host, strconv.FormatInt(int64(resp.StatusCode), 10)).Inc()

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return nil
	}
	if 400 <= resp.StatusCode && resp.StatusCode < 500 {
		// client-side error ??!
		return eerrors.Errorf("HTTP error when sending message to server: code '%d', status '%s'", resp.StatusCode, resp.Status)
	}
	if 500 <= resp.StatusCode && resp.StatusCode < 600 {
		// server side error
		return eerrors.Errorf("HTTP error when sending message to server: code '%d', status '%s'", resp.StatusCode, resp.Status)
	}
	return eerrors.Errorf("HTTP error when sending message to server: code '%d', status '%s'", resp.StatusCode, resp.Status)
}

func (d *HTTPDestination) dequeue(ctx context.Context) error {
	for {
		defered, err := d.queue.Get()
		if err != nil || defered == nil {
			return nil
		}
		err = d.doHTTP(ctx, defered.UID, defered.Request)
		if err != nil {
			d.NACK(defered.UID)
			return err
		}
		d.ACK(defered.UID)
	}
}

var ErrCalculateURL = eerrors.New("Error calculating target URL from template")

func (d *HTTPDestination) enqueue(ctx context.Context, msg *model.FullMessage) (err error) {
	urlbuf := bytebufferpool.Get()
	body := bytebufferpool.Get()
	defer func() {
		bytebufferpool.Put(body)
		bytebufferpool.Put(urlbuf)
	}()
	err = d.url.Execute(urlbuf, msg.Fields)
	if err != nil {
		return encoders.EncodingError(ErrCalculateURL)
	}
	err = d.encoder(msg, body)
	if err != nil {
		return err
	}

	// we use String() methods to get a copy of the bytebuffers, so that we can Put them back to the pool afterwards
	req, err := http.NewRequest(d.method, urlbuf.String(), strings.NewReader(body.String()))
	if err != nil {
		return encoders.EncodingError(eerrors.Wrap(err, "Error preparing HTTP request"))
	}

	dreq := &model.DeferedRequest{Request: req, UID: msg.Uid}
	return d.queue.Put(dreq)
}

func (d *HTTPDestination) Send(ctx context.Context, msgs []model.OutputMsg) (err eerrors.ErrorSlice) {
	return d.ForEach(ctx, d.enqueue, false, true, msgs)
}
