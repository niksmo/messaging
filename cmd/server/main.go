package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/niksmo/messaging/internal/messaging"
	"github.com/niksmo/messaging/internal/processor/collector"
	"github.com/niksmo/messaging/internal/server"
	"github.com/niksmo/messaging/pkg/logger"
)

type config struct {
	logLevel          string
	addr              string
	brokers           []string
	outTopic          string
	inTopic           string
	partitions        int
	replicationFactor int
	closeTimeout      time.Duration
}

func main() {
	sigCatcher, sigCatcherCancel := signalCatcher()
	defer sigCatcherCancel()

	config := laodConfig()
	logger := logger.New(config.logLevel)

	app := createApp(logger, config)
	app.Run(sigCatcher)

	<-sigCatcher.Done()
	closeAppWithTimeout(app, config.closeTimeout)
}

func signalCatcher() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(
		context.Background(),
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
}

func laodConfig() config {
	return config{
		logLevel: "info",
		addr:     "127.0.0.1:8000",
		brokers: []string{
			"127.0.0.1:19094",
			"127.0.0.1:29094",
			"127.0.0.1:39094",
		},
		outTopic:          messaging.Stream,
		inTopic:           string(collector.Group),
		partitions:        3,
		replicationFactor: 2,
		closeTimeout:      5 * time.Second,
	}
}

func createApp(logger logger.Logger, cfg config) *server.App {
	serverOpts := []server.Option{
		server.WithAddr(cfg.addr),
		server.WithBrokers(cfg.brokers),
		server.WithOutTopic(cfg.outTopic),
		server.WithInTopic(cfg.inTopic),
	}

	app, err := server.New(logger, serverOpts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create server application")
	}
	return app
}

func closeAppWithTimeout(app *server.App, timeout time.Duration) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	app.Close(timeoutCtx)
}
