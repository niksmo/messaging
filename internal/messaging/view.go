package messaging

import (
	"context"

	"github.com/lovoo/goka"
	"github.com/niksmo/messaging/pkg/logger"
)

type View struct {
	gv *goka.View
}

func NewView(
	logger logger.Logger, brokers []string, topic string,
) (*View, error) {
	gv, err := goka.NewView(
		brokers, goka.GroupTable(goka.Group(topic)), NewMessageListCodec(logger),
	)

	if err != nil {
		return nil, err
	}
	return &View{gv}, nil
}

func (v *View) Run(ctx context.Context) error {
	return v.gv.Run(ctx)
}

func (v *View) Get(key string) (any, error) {
	return v.gv.Get(key)
}
