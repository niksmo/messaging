package collector

import (
	"context"
	"fmt"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/internal/messaging"
	"github.com/niksmo/messaging/pkg/logger"
)

const (
	Group       goka.Group  = "collector-group"
	inputStream goka.Stream = "filtered_messages"
)

func Run(ctx context.Context, logger logger.Logger, brokers []string) error {
	const op = "collector.Run"

	g := makeGroupGraph(logger)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return p.Run(ctx)
}

func makeGroupGraph(logger logger.Logger) *goka.GroupGraph {
	msgCodec := messaging.NewMessageCodec(logger)
	return goka.DefineGroup(
		Group,
		goka.Input(inputStream, msgCodec, inputCallback(logger)),
		goka.Loop(msgCodec, loopCallback(logger)),
		goka.Persist(messaging.NewMessageListCodec(logger)),
	)
}

func inputCallback(logger logger.Logger) goka.ProcessCallback {
	const op = "collector.inputCallback"
	log := logger.WithOp(op)

	return func(ctx goka.Context, msg any) {
		msgt, ok := msg.(messaging.Message)
		if !ok {
			log.Error().Type("msgType", msg).Msg("invalid msg type")
			return
		}
		ctx.Loopback(msgt.To, msg)
	}
}

func loopCallback(logger logger.Logger) goka.ProcessCallback {
	const op = "collector.loopCallback"
	log := logger.WithOp(op)

	return func(ctx goka.Context, msg any) {
		msgt, ok := msg.(messaging.Message)
		if !ok {
			log.Error().Type("msgType", msg).Msg("invalid msg type")
			return
		}

		var ml []messaging.Message
		if v := ctx.Value(); v != nil {
			vml, ok := v.([]messaging.Message)
			if !ok {
				log.Error().Type("msgListType", msg).Msg("invalid msg list type")
				return
			}
			ml = vml
		}

		ml = append(ml, msgt)
		ctx.SetValue(ml)
	}
}
