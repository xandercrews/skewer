package dests

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/oklog/ulid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stephane-martin/skewer/clients"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/utils"
	"github.com/stephane-martin/skewer/utils/queue"
)

type relpDestination struct {
	logger     log15.Logger
	fatal      chan struct{}
	registry   *prometheus.Registry
	client     *clients.RELPClient
	once       sync.Once
	ack        storeCallback
	nack       storeCallback
	permerr    storeCallback
	ackCounter *prometheus.CounterVec
}

func NewRelpDestination(ctx context.Context, bc conf.BaseConfig, ack, nack, permerr storeCallback, logger log15.Logger) (dest Destination, err error) {
	clt := clients.NewRELPClient(logger).
		Host(bc.RelpDest.Host).
		Port(bc.RelpDest.Port).
		Path(bc.RelpDest.UnixSocketPath).
		Format(bc.RelpDest.Format).
		KeepAlive(bc.RelpDest.KeepAlive).
		KeepAlivePeriod(bc.RelpDest.KeepAlivePeriod).
		ConnTimeout(bc.RelpDest.ConnTimeout).
		RelpTimeout(bc.RelpDest.RelpTimeout).
		WindowSize(bc.RelpDest.WindowSize).
		FlushPeriod(bc.RelpDest.FlushPeriod)

	if bc.RelpDest.TLSEnabled {
		config, err := utils.NewTLSConfig(
			bc.RelpDest.Host,
			bc.RelpDest.CAFile,
			bc.RelpDest.CAPath,
			bc.RelpDest.CertFile,
			bc.RelpDest.KeyFile,
			bc.RelpDest.Insecure,
		)
		if err != nil {
			return nil, err
		}
		clt = clt.TLS(config)
	}

	err = clt.Connect()
	if err != nil {
		return nil, err
	}

	d := &relpDestination{
		logger:     logger,
		registry:   prometheus.NewRegistry(),
		ack:        ack,
		nack:       nack,
		permerr:    permerr,
		fatal:      make(chan struct{}),
		client:     clt,
		ackCounter: relpAckCounter,
	}

	rebind := bc.RelpDest.Rebind
	if rebind > 0 {
		go func() {
			select {
			case <-ctx.Done():
				// the store service asked for stop
			case <-time.After(rebind):
				logger.Info("RELP destination rebind period has expired", "rebind", rebind.String())
				d.once.Do(func() { close(d.fatal) })
			}
		}()
	}

	go func() {
		// TODO: metrics
		ackChan := d.client.Ack()
		nackChan := d.client.Nack()
		var err error
		var uid ulid.ULID

		for queue.WaitManyAckQueues(ackChan, nackChan) {
			for {
				uid, _, err = ackChan.Get()
				if err != nil {
					break
				}
				d.ack(uid, conf.Relp)
			}
			for {
				uid, _, err = nackChan.Get()
				if err != nil {
					break
				}
				d.nack(uid, conf.Relp)
				d.logger.Info("RELP server returned a NACK", "uid", uid.String())
				d.once.Do(func() { close(d.fatal) })
			}
		}
	}()

	return d, nil
}

func (d *relpDestination) Send(message model.FullMessage, partitionKey string, partitionNumber int32, topic string) (err error) {
	err = d.client.Send(&message)
	if err != nil {
		// the client send queue has been disposed
		d.nack(message.Uid, conf.Relp)
		d.once.Do(func() { close(d.fatal) })
	}
	return
}

func (d *relpDestination) Close() (err error) {
	return d.client.Close()
}

func (d *relpDestination) Fatal() chan struct{} {
	return d.fatal
}

func (d *relpDestination) Gather() ([]*dto.MetricFamily, error) {
	return d.registry.Gather()
}
