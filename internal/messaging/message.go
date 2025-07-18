package messaging

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/niksmo/messaging/pkg/logger"
)

const Stream = "messages"

type Message struct {
	From, To, Content string
}

type MessageCodec struct {
	log logger.Logger
}

func NewMessageCodec(l logger.Logger) MessageCodec {
	return MessageCodec{l}
}

func (c MessageCodec) Encode(value any) ([]byte, error) {
	const op = "MessageCodec.Encode"
	log := c.log.WithOp(op)
	vt, ok := value.(Message)
	if !ok {
		log.Error().Msg("value is not message type")
		return nil, fmt.Errorf("%s: %w", op, errors.New("not message type"))
	}

	b, err := json.Marshal(vt)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal message")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return b, nil
}

func (c MessageCodec) Decode(data []byte) (any, error) {
	const op = "MessageCodec.Decode"
	log := c.log.WithOp(op)

	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal message")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return m, nil
}

type MessageListCodec struct {
	log logger.Logger
}

func NewMessageListCodec(l logger.Logger) MessageListCodec {
	return MessageListCodec{l}
}

func (c MessageListCodec) Encode(value any) ([]byte, error) {
	const op = "MessageListCodec.Encode"
	log := c.log.WithOp(op)

	if _, ok := value.([]Message); !ok {
		log.Error().Msg("value is not message list type")
		return nil, fmt.Errorf(
			"%s: %w", op, errors.New("not message list type"))
	}

	b, err := json.Marshal(value)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal message list")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return b, nil
}

func (c MessageListCodec) Decode(data []byte) (any, error) {
	const op = "MessageListCodec.Decode"
	log := c.log.WithOp(op)

	var m []Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal message list")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return m, nil
}
