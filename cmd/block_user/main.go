package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/internal/processor/blocker"
	"github.com/niksmo/messaging/pkg/logger"
)

type config struct {
	logLevel string
	brokers  []string
	topic    string
}

func main() {
	config := loadConfig()
	logger := logger.New(config.logLevel)

	name, blocked := getFlags()

	if err := validateName(name, logger); err != nil {
		logger.Error().Err(err).Send()
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	emitter := createEmitter(logger, config.brokers, config.topic)

	err := emitter.EmitSync(name, blocker.BlockValue(blocked))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to emit block value")
	}
	logger.Info().Str("name", name).Bool("blocked", blocked).Send()
}

func loadConfig() config {
	return config{
		logLevel: "info",
		brokers: []string{
			"127.0.0.1:19094",
			"127.0.0.1:29094",
			"127.0.0.1:39094",
		},
		topic: string(blocker.Stream),
	}
}

func getFlags() (name string, blocked bool) {
	flag.BoolVar(&blocked, "blocked", false, "block value")
	flag.StringVar(&name, "name", "", "user name")
	flag.Parse()
	name = strings.TrimSpace(name)
	return
}

func validateName(name string, log logger.Logger) error {
	if name == "" {
		return errors.New("name is empty")
	}

	return nil
}

func createEmitter(log logger.Logger, brokers []string, topic string) *goka.Emitter {
	codec := blocker.NewBlockValueCodec(log)
	e, err := goka.NewEmitter(brokers, goka.Stream(topic), codec)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to construct emitter")
	}
	return e
}
