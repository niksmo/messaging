package blocker

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/pkg/logger"
)

const (
	group  goka.Group  = "blocker-group"
	Stream goka.Stream = "blocked_users"
)

var Table goka.Table = goka.GroupTable(group)

func Run(ctx context.Context, logger logger.Logger, brokers []string) error {
	const op = "blocker.Run"

	g := makeGroupGraph(logger)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return p.Run(ctx)
}

func makeGroupGraph(logger logger.Logger) *goka.GroupGraph {
	blockValueCodec := NewBlockValueCodec(logger)
	return goka.DefineGroup(
		group,
		goka.Input(Stream, blockValueCodec, processCallback(logger)),
		goka.Persist(blockValueCodec),
	)
}

func processCallback(logger logger.Logger) goka.ProcessCallback {
	const op = "blocker.processCallback"
	log := logger.WithOp(op)

	return func(ctx goka.Context, msg any) {
		v, ok := msg.(BlockValue)
		if !ok {
			log.Error().Type("msgType", msg).Msg("invalid msg type")
			return
		}
		ctx.SetValue(v)
	}
}

type BlockValue bool

type BlockValueCodec struct {
	log logger.Logger
}

func NewBlockValueCodec(log logger.Logger) BlockValueCodec {
	return BlockValueCodec{log}
}

func (v BlockValueCodec) Encode(value any) ([]byte, error) {
	const op = "BlockValueCodec.Encode"
	log := v.log.WithOp(op)
	vt, ok := value.(BlockValue)
	if !ok {
		log.Error().Msg("invalid value type")
		return nil, fmt.Errorf("%s: %w",
			op, errors.New("invalid value type"))
	}

	s := strconv.FormatBool(bool(vt))

	return []byte(s), nil
}

func (v BlockValueCodec) Decode(data []byte) (any, error) {
	const op = "BlockValueCodec.Decode"
	log := v.log.WithOp(op)

	bv, err := strconv.ParseBool(string(data))
	if err != nil {
		log.Error().Err(err).Msg("data could not be converted to a boolean value")
		return nil, fmt.Errorf("%s: %w", op, err)

	}
	return BlockValue(bv), nil
}
