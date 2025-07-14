package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/niksmo/messaging/internal/processor"
	"github.com/niksmo/messaging/pkg/logger"
)

type config struct {
	logLevel string
	brokers  []string
	npart    int
	rFactor  int
}

func main() {
	sigCatcher, sigCatcherCancel := signalCatcher()
	defer sigCatcherCancel()

	config := laodConfig()
	logger := logger.New(config.logLevel)

	processor.Run(sigCatcher, logger,
		processor.WithOptions(config.brokers, config.npart, config.rFactor))
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
		brokers: []string{
			"127.0.0.1:19094",
			"127.0.0.1:29094",
			"127.0.0.1:39094",
		},
		npart:   3,
		rFactor: 2,
	}
}
