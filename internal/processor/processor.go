package processor

import (
	"context"

	"github.com/niksmo/messaging/internal/processor/blocker"
	"github.com/niksmo/messaging/internal/processor/censor"
	"github.com/niksmo/messaging/internal/processor/collector"
	"github.com/niksmo/messaging/internal/processor/filter"
	"github.com/niksmo/messaging/pkg/logger"
	"github.com/niksmo/messaging/pkg/topicinit"
	"golang.org/x/sync/errgroup"
)

func WithOptions(brokers []string, npart, rfactor int) *options {
	return &options{brokers, npart, rfactor}
}

type options struct {
	brokers []string
	npart   int
	rfactor int
}

type procRunner func(context.Context, logger.Logger, []string) error

func Run(ctx context.Context, log logger.Logger, opt *options) {
	const op = "processor.Run"
	log = log.WithOp(op)

	g, ctx := errgroup.WithContext(ctx)

	initTopics(log, opt.brokers, opt.npart, opt.rfactor)

	runProcessors(ctx, g, log, opt.brokers)

	log.Info().Msg("processors are running")

	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msg("processors are stopped with error")
	}
	log.Info().Msg("processors are stopped")
}

func initTopics(log logger.Logger, brokers []string, npart, rfactor int) {
	log.Info().Msg("initializing topics")
	var errs []error
	topics := []string{
		string(filter.InputStream),
		string(filter.OutputStream),
		string(blocker.Stream),
		string(censor.Stream),
	}

	for _, topic := range topics {
		err := topicinit.EnsureTopicExists(
			topic, brokers, npart, rfactor,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	tables := []string{
		string(filter.BlockerTable),
		string(filter.CensorTable),
	}
	for _, table := range tables {
		err := topicinit.EnsureTableExists(table, brokers, npart)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		log.Error().Errs("errs", errs).Msg("failed to init topics")
	}
	log.Info().Msg("the topics are initialized")
}

func runProcessors(
	ctx context.Context,
	g *errgroup.Group,
	log logger.Logger,
	brokers []string,
) {
	procRunners := []procRunner{
		blocker.Run,
		censor.Run,
		filter.Run,
		collector.Run,
	}
	for _, runner := range procRunners {
		g.Go(func() error {
			return runner(ctx, log, brokers)
		})
	}
}
