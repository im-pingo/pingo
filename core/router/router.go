package router

import (
	"context"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/stream"
	"github.com/pingostack/pingos/pkg/avframe"
)

type Router struct {
	id     string
	stream *stream.Stream
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRouter(ctx context.Context, id string) *Router {
	ctx, cancel := context.WithCancel(ctx)

	return &Router{
		id:     id,
		stream: stream.NewStream(ctx, id),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (r *Router) ID() string {
	return r.id
}

func (r *Router) Stream() *stream.Stream {
	return r.stream
}

func (r *Router) Publish(publisher peer.Publisher) error {
	return BuildPublishMiddleware(r.ctx, func(ctx context.Context, publisher peer.Publisher, _ Stage) error {
		return r.stream.Publish(publisher)
	})(r.ctx, publisher, StageStart)
}

func (r *Router) Subscribe(subscriber peer.Subscriber) error {
	return BuildSubscribeMiddleware(r.ctx, func(ctx context.Context, subscriber peer.Subscriber, _ Stage) error {
		processor, err := r.stream.Subscribe(subscriber, func(sub peer.Subscriber, processor avframe.Processor, err error) {
			if err != nil {
				return
			}
			sub.SetProcessor(processor)
		})

		if processor != nil {
			subscriber.SetProcessor(processor)
		}

		return err
	})(r.ctx, subscriber, StageStart)
}

func (r *Router) Close() {
	r.cancel()
}
