package messaging

import (
	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/pkg/logger"
)

type Emitter struct {
	ge *goka.Emitter
}

func NewEmitter(
	logger logger.Logger, brokers []string, topic string,
) (*Emitter, error) {
	ge, err := goka.NewEmitter(
		brokers, goka.Stream(topic), NewMessageCodec(logger),
	)
	if err != nil {
		return nil, err
	}
	return &Emitter{ge}, nil
}

func (e *Emitter) Emit(key string, msg Message) error {
	return e.ge.EmitSync(key, msg)
}
