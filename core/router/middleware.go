package router

import (
	"context"
	"sync"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/stream"
	"github.com/pingostack/pingos/pkg/avframe"
)

// stage defines a middleware stage
type Stage int

const (
	StageStart Stage = iota
	StageEnd
)

func (s Stage) String() string {
	return []string{"start", "end"}[s]
}

func StageString(s Stage) string {
	return s.String()
}

func StageFromString(s string) Stage {
	switch s {
	case "start":
		return StageStart
	case "end":
		return StageEnd
	default:
		return StageStart
	}
}

// MiddlewareFunc defines a generic middleware function
type MiddlewareFunc[T any] func(ctx context.Context, t T, stage Stage) error

// MiddlewareChain manages a chain of middleware
type MiddlewareChain[T any] struct {
	middlewares []MiddlewareFunc[T]
	lock        sync.RWMutex
}

// RegisterMiddleware adds a middleware to the chain
func (mc *MiddlewareChain[T]) RegisterMiddleware(mw MiddlewareFunc[T]) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	mc.middlewares = append(mc.middlewares, mw)
}

// Handle executes the middleware chain
func (mc *MiddlewareChain[T]) Handle(ctx context.Context, t T, stage Stage) error {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	for _, mw := range mc.middlewares {
		if err := mw(ctx, t, stage); err != nil {
			return err
		}
	}
	return nil
}

// BuildChain constructs the middleware chain
func (mc *MiddlewareChain[T]) BuildChain(final MiddlewareFunc[T]) MiddlewareFunc[T] {
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		next := final
		mw := mc.middlewares[i]
		final = func(ctx context.Context, t T, stage Stage) error {
			if err := mw(ctx, t, stage); err != nil {
				return err
			}
			return next(ctx, t, stage)
		}
	}
	return final
}

// Specific middleware chains for different types
var publishChain = &MiddlewareChain[peer.Publisher]{}
var subscribeChain = &MiddlewareChain[peer.Subscriber]{}
var streamChain = &MiddlewareChain[*stream.Stream]{}
var metadataChain = &MiddlewareChain[avframe.Metadata]{}
var pullChain = &MiddlewareChain[peer.Publisher]{}
var pushChain = &MiddlewareChain[peer.Subscriber]{}

// RegisterPublishMiddleware adds a publish middleware
func RegisterPublishMiddleware(mw MiddlewareFunc[peer.Publisher]) {
	publishChain.RegisterMiddleware(mw)
}

// HandlePublish handles the publish middleware chain
func HandlePublish(ctx context.Context, publisher peer.Publisher, stage Stage) error {
	return publishChain.Handle(ctx, publisher, stage)
}

// BuildPublishMiddleware builds the publish middleware chain
func BuildPublishMiddleware(ctx context.Context, final MiddlewareFunc[peer.Publisher]) MiddlewareFunc[peer.Publisher] {
	return publishChain.BuildChain(final)
}

// RegisterSubscribeMiddleware adds a subscribe middleware
func RegisterSubscribeMiddleware(mw MiddlewareFunc[peer.Subscriber]) {
	subscribeChain.RegisterMiddleware(mw)
}

// HandleSubscribe handles the subscribe middleware chain
func HandleSubscribe(ctx context.Context, subscriber peer.Subscriber, stage Stage) error {
	return subscribeChain.Handle(ctx, subscriber, stage)
}

// BuildSubscribeMiddleware builds the subscribe middleware chain
func BuildSubscribeMiddleware(ctx context.Context, final MiddlewareFunc[peer.Subscriber]) MiddlewareFunc[peer.Subscriber] {
	return subscribeChain.BuildChain(final)
}

// RegisterStreamMiddleware adds a stream middleware
func RegisterStreamMiddleware(mw MiddlewareFunc[*stream.Stream]) {
	streamChain.RegisterMiddleware(mw)
}

// HandleStream handles the stream middleware chain
func HandleStream(ctx context.Context, stream *stream.Stream, stage Stage) error {
	return streamChain.Handle(ctx, stream, stage)
}

// BuildStreamMiddleware builds the stream middleware chain
func BuildStreamMiddleware(ctx context.Context, final MiddlewareFunc[*stream.Stream]) MiddlewareFunc[*stream.Stream] {
	return streamChain.BuildChain(final)
}

// RegisterMetadataMiddleware adds a metadata middleware
func RegisterMetadataMiddleware(mw MiddlewareFunc[avframe.Metadata]) {
	metadataChain.RegisterMiddleware(mw)
}

// HandleMetadata handles the metadata middleware chain
func HandleMetadata(ctx context.Context, metadata avframe.Metadata, stage Stage) error {
	return metadataChain.Handle(ctx, metadata, stage)
}

// BuildMetadataMiddleware builds the metadata middleware chain
func BuildMetadataMiddleware(ctx context.Context, final MiddlewareFunc[avframe.Metadata]) MiddlewareFunc[avframe.Metadata] {
	return metadataChain.BuildChain(final)
}

// RegisterPullMiddleware adds a pull middleware
func RegisterPullMiddleware(mw MiddlewareFunc[peer.Publisher]) {
	pullChain.RegisterMiddleware(mw)
}

// HandlePull handles the pull middleware chain
func HandlePull(ctx context.Context, publisher peer.Publisher, stage Stage) error {
	return pullChain.Handle(ctx, publisher, stage)
}

// BuildPullMiddleware builds the pull middleware chain
func BuildPullMiddleware(ctx context.Context, final MiddlewareFunc[peer.Publisher]) MiddlewareFunc[peer.Publisher] {
	return pullChain.BuildChain(final)
}

// RegisterPushMiddleware adds a push middleware
func RegisterPushMiddleware(mw MiddlewareFunc[peer.Subscriber]) {
	pushChain.RegisterMiddleware(mw)
}

// HandlePush handles the push middleware chain
func HandlePush(ctx context.Context, subscriber peer.Subscriber, stage Stage) error {
	return pushChain.Handle(ctx, subscriber, stage)
}

// BuildPushMiddleware builds the push middleware chain
func BuildPushMiddleware(ctx context.Context, final MiddlewareFunc[peer.Subscriber]) MiddlewareFunc[peer.Subscriber] {
	return pushChain.BuildChain(final)
}
