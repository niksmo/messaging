package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/internal/processor/censor"
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

	word, change := getFlags()

	validateFlags(logger, word, change)

	emitter := createEmitter(logger, config.brokers, config.topic)

	err := emitter.EmitSync(word, change)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to emit censor word")
	}
	logger.Info().Str("word", word).Str("change", change).Send()
}

func loadConfig() config {
	return config{
		logLevel: "info",
		brokers: []string{
			"127.0.0.1:19094",
			"127.0.0.1:29094",
			"127.0.0.1:39094",
		},
		topic: string(censor.Stream),
	}
}

func getFlags() (word string, change string) {
	flag.StringVar(&word, "word", "", "replaced word")
	flag.StringVar(&change, "change", "", "change on")
	flag.Parse()

	strFlags := []string{word, change}

	for i := range strFlags {
		strFlags[i] = strings.TrimSpace(strFlags[i])
	}

	return strFlags[0], strFlags[1]
}

func validateFlags(log logger.Logger, word, change string) {
	var errs []error
	if err := validateWord(word, log); err != nil {
		errs = append(errs, err)
	}

	if err := validateChange(change, log); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		log.Error().Errs("flagErrs", errs).Send()
		flag.CommandLine.Usage()
		os.Exit(1)
	}
}

func validateWord(word string, log logger.Logger) error {
	if word == "" {
		return errors.New("word is empty")
	}
	return nil
}

func validateChange(change string, log logger.Logger) error {
	if change == "" {
		return errors.New("change is empty")
	}
	return nil
}

func createEmitter(log logger.Logger, brokers []string, topic string) *goka.Emitter {
	codec := censor.NewCensorValueCodec(log)
	e, err := goka.NewEmitter(brokers, goka.Stream(topic), codec)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to construct emitter")
	}
	return e
}
