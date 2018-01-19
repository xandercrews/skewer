package dests

import (
	"context"
	"fmt"
	"os"

	"github.com/stephane-martin/skewer/conf"
	"github.com/stephane-martin/skewer/model"
	"github.com/stephane-martin/skewer/model/encoders"
)

type StderrDestination struct {
	*baseDestination
}

func NewStderrDestination(ctx context.Context, e *Env) (Destination, error) {
	d := &StderrDestination{
		baseDestination: newBaseDestination(conf.Stderr, "stderr", e),
	}
	err := d.setFormat(e.config.StderrDest.Format)
	if err != nil {
		return nil, fmt.Errorf("Error getting encoder: %s", err)
	}

	return d, nil
}

func (d *StderrDestination) Send(message *model.FullMessage, partitionKey string, partitionNumber int32, topic string) (err error) {
	var buf []byte
	buf, err = encoders.ChainEncode(d.encoder, message, "\n")
	if err != nil {
		d.permerr(message.Uid)
		model.Free(message.Fields)
		return err
	}
	_, err = os.Stderr.Write(buf)
	if err != nil {
		d.nack(message.Uid)
		d.dofatal()
		model.Free(message.Fields)
		return err
	}
	d.ack(message.Uid)
	model.Free(message.Fields)
	return nil
}

func (d *StderrDestination) Close() error {
	return nil
}
