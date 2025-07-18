package censor

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/pkg/logger"
)

const (
	group  goka.Group  = "censor-group"
	Stream goka.Stream = "censored_words"
)

var Table goka.Table = goka.GroupTable(group)

func Run(ctx context.Context, logger logger.Logger, brokers []string) error {
	const op = "censor.Run"

	g := makeGroupGraph(logger)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return p.Run(ctx)
}

func makeGroupGraph(logger logger.Logger) *goka.GroupGraph {
	censorCodec := NewCensorValueCodec(logger)
	return goka.DefineGroup(
		group,
		goka.Input(Stream, censorCodec, processCallback(logger)),
		goka.Persist(censorCodec),
	)
}

func processCallback(logger logger.Logger) goka.ProcessCallback {
	const op = "censor.processCallback"
	log := logger.WithOp(op)

	return func(ctx goka.Context, msg any) {
		v, ok := msg.(string)
		if !ok {
			log.Error().Type("msgType", msg).Msg("invalid msg type")
			return
		}
		ctx.SetValue(v)
	}
}

type CensorValueCodec struct {
	log logger.Logger
}

func NewCensorValueCodec(log logger.Logger) CensorValueCodec {
	return CensorValueCodec{log}
}

func (v CensorValueCodec) Encode(value any) ([]byte, error) {
	const op = "CensorValueCodec.Encode"
	log := v.log.WithOp(op)
	s, ok := value.(string)
	if !ok {
		log.Error().Msg("invalid value type")
		return nil, fmt.Errorf("%s: %w",
			op, errors.New("invalid value type"))
	}

	return []byte(s), nil
}

func (v CensorValueCodec) Decode(data []byte) (any, error) {
	const op = "CensorValueCodec.Decode"
	log := v.log.WithOp(op)

	if !utf8.Valid(data) {
		log.Error().Msg("invalid value type")
		return nil, fmt.Errorf("%s: %w", op, errors.New("invalid value type"))
	}

	return string(data), nil
}
