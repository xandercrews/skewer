package store

import (
	"context"

	dto "github.com/prometheus/client_model/go"
	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/utils"
)

type Store interface {
	Stash(m *model.FullMessage) (error, error)
	Outputs(dest conf.DestinationType) chan *model.FullMessage
	ACK(uid utils.MyULID, dest conf.DestinationType)
	NACK(uid utils.MyULID, dest conf.DestinationType)
	PermError(uid utils.MyULID, dest conf.DestinationType)
	Errors() chan struct{}
	WaitFinished()
	GetSyslogConfig(configID utils.MyULID) (*conf.FilterSubConfig, error)
	StoreAllSyslogConfigs(c conf.BaseConfig) error
	ReadAllBadgers() (map[string]string, map[string]string, map[string]string)
	Destinations() []conf.DestinationType
	Confined() bool
}

type Forwarder interface {
	Forward(ctx context.Context)
	Fatal() chan struct{}
	WaitFinished()
	Gather() ([]*dto.MetricFamily, error)
}
