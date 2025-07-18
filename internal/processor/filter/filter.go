package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/internal/messaging"
	"github.com/niksmo/messaging/internal/processor/blocker"
	"github.com/niksmo/messaging/internal/processor/censor"
	"github.com/niksmo/messaging/pkg/logger"
)

const (
	group        goka.Group  = "filter-group"
	InputStream  goka.Stream = messaging.Stream
	OutputStream goka.Stream = "filtered_messages"
)

var (
	BlockerTable = blocker.Table
	CensorTable  = censor.Table
)

func Run(ctx context.Context, logger logger.Logger, brokers []string) error {
	const op = "filter.Run"

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
		group,
		goka.Input(InputStream, msgCodec, processCallback(logger)),
		goka.Output(OutputStream, msgCodec),
		goka.Join(BlockerTable, blocker.NewBlockValueCodec(logger)),
		goka.Lookup(CensorTable, censor.NewCensorValueCodec(logger)),
	)
}

func processCallback(logger logger.Logger) goka.ProcessCallback {
	const op = "filter.processCallback"
	log := logger.WithOp(op)

	return func(ctx goka.Context, msg any) {
		log := log.With().Str("user", ctx.Key()).Logger()

		m, ok := msg.(messaging.Message)
		if !ok {
			log.Error().Type("msgType", msg).Msg("invalid msg type")
			return
		}

		log.Info().Msg("receive message")

		if senderBlocked(ctx) {
			log.Info().Str("reason", "user is blocked").Msg("skipped")
			return
		}

		if applyCensor(ctx, &m) {
			log.Info().Msg("censored")
		}

		ctx.Emit(OutputStream, m.From, m)

		log.Info().Msg("forward to filtered")
	}
}

func senderBlocked(ctx goka.Context) bool {
	v, ok := ctx.Join(BlockerTable).(blocker.BlockValue)
	return ok && bool(v)
}

func applyCensor(ctx goka.Context, msg *messaging.Message) (apply bool) {
	s := strings.Fields(msg.Content)
	for i, word := range s {
		replacement := ctx.Lookup(CensorTable, word)
		if replacement != nil {
			s[i] = replacement.(string)
			apply = true
		}
	}
	msg.Content = strings.Join(s, " ")
	return
}
