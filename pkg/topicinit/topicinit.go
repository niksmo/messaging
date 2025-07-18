package topicinit

import (
	"fmt"

	"github.com/lovoo/goka"
)

func EnsureTopicExists(
	topic string, brokers []string, npart, rfactor int,
) error {
	const op = "topicinit.EnsureTopicExists"
	tm, err := createTopicManager(brokers)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tm.Close()
	err = tm.EnsureTopicExists(topic, npart, rfactor, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func EnsureTableExists(topic string, brokers []string, npart int) error {
	const op = "topicinit.EnsureTableExists"
	tm, err := createTopicManager(brokers)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tm.Close()
	err = tm.EnsureTableExists(topic, npart)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func createTopicManager(brokers []string) (goka.TopicManager, error) {
	const op = "topicinit.createTopicManager"
	tmc := goka.NewTopicManagerConfig()
	tm, err := goka.NewTopicManager(brokers, goka.DefaultConfig(), tmc)
	if err != nil {
		return tm, fmt.Errorf("%s: %w", op, err)
	}
	return tm, nil
}
