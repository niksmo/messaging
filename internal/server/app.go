package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/niksmo/messaging/internal/messaging"
	"github.com/niksmo/messaging/pkg/logger"
)

type options struct {
	addr     string
	brokers  []string
	outTopic string
	inTopic  string
}

type App struct {
	log logger.Logger
	s   *http.Server
	e   *messaging.Emitter
	v   *messaging.View
}

type Option func(*options) error

func WithAddr(addr string) Option {
	return func(o *options) error {
		o.addr = addr
		return nil
	}
}

func WithBrokers(brokers []string) Option {
	return func(o *options) error {
		o.brokers = brokers
		return nil
	}
}

func WithOutTopic(topic string) Option {
	return func(o *options) error {
		o.outTopic = topic
		return nil
	}
}

func WithInTopic(topic string) Option {
	return func(o *options) error {
		o.inTopic = topic
		return nil
	}
}

func New(l logger.Logger, opts ...Option) (*App, error) {
	var options options
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	s := &http.Server{Addr: options.addr}

	app := &App{log: l, s: s}

	err := app.initMsgEmitter(options.brokers, options.outTopic)
	if err != nil {
		return nil, err
	}

	err = app.initMsgView(options.brokers, options.inTopic)
	if err != nil {
		return nil, err
	}

	app.setupHandler()

	return app, nil
}

func (a *App) Run(ctx context.Context) {
	const op = "server.Run"
	log := a.log.WithOp(op)

	log.Info().Str("addr", a.s.Addr).Msg("start listening")

	_, cancel := context.WithCancel(ctx)
	defer cancel()

	go a.runServer(ctx, func(err error) {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("server start closing")
			return
		}
		if err != nil {
			log.Error().Err(err).Msg(
				"internal server error")
			cancel()
		}
	})

	go a.runView(ctx, func(err error) {
		log.Error().Err(err).Msg("failed to run view")
		cancel()
	})
}

func (a *App) Close(timeoutCtx context.Context) {
	const op = "server.Close"
	log := a.log.WithOp(op)

	if err := a.s.Shutdown(timeoutCtx); err != nil {
		log.Error().Err(err).Msg("failed shutdown server gracefully")
		return
	}
	log.Info().Msg("server closed gracefully")
}

func (a *App) initMsgEmitter(brokers []string, topic string) error {
	e, err := messaging.NewEmitter(a.log, brokers, topic)
	if err != nil {
		return fmt.Errorf("failed to construct message emitter: %w", err)
	}
	a.e = e
	return nil
}

func (a *App) initMsgView(brokers []string, topic string) error {
	v, err := messaging.NewView(a.log, brokers, topic)
	if err != nil {
		return fmt.Errorf("failed to construct message view: %w", err)
	}
	a.v = v
	return nil
}

func (a *App) setupHandler() {
	mux := http.NewServeMux()
	NewHandler(a.log, mux, a.e, a.v)
	a.s.Handler = mux
}

func (a *App) runServer(ctx context.Context, errCb func(error)) {
	err := a.s.ListenAndServe()
	if err != nil {
		errCb(err)
	}
}

func (a *App) runView(ctx context.Context, errCb func(error)) {
	err := a.v.Run(ctx)
	if err != nil {
		errCb(err)
	}
}
