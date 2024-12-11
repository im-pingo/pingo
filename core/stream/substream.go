package stream

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

const (
	FeedbackTypeSubscriberActive = "subscriber_active"
)

type FeedbackSubscriberActive struct {
	Subscriber peer.Subscriber
	Active     bool
}

type subscriberInfo struct {
	subscriber peer.Subscriber
	active     bool
}

type subStream struct {
	ctx               context.Context
	cancel            context.CancelFunc
	subscribers       []subscriberInfo
	lock              sync.RWMutex
	closeOnce         sync.Once
	tailProcessor     *avframe.Pipeline
	onSubscriberEmpty func(sub peer.Subscriber)
	onSubStreamClosed func()
	logger            logger.Logger

	deadlineTimer *time.Timer
}

type SubStreamOption func(ss *subStream)

func WithOnSubscriberEmpty(fn func(sub peer.Subscriber)) SubStreamOption {
	return func(ss *subStream) {
		ss.onSubscriberEmpty = fn
	}
}

func WithLogger(logger logger.Logger) SubStreamOption {
	return func(ss *subStream) {
		ss.logger = logger
	}
}

func WithOnSubStreamClosed(fn func()) SubStreamOption {
	return func(ss *subStream) {
		ss.onSubStreamClosed = fn
	}
}

func newSubStream(ctx context.Context, processor *avframe.Pipeline, opts ...SubStreamOption) (ss *subStream, err error) {
	ctx, cancel := context.WithCancel(ctx)
	ss = &subStream{
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make([]subscriberInfo, 0),
	}

	for _, opt := range opts {
		opt(ss)
	}

	if ss.logger == nil {
		ss.logger = logger.WithFields(map[string]interface{}{"substream": processor.Format()})
	}

	interceptors, err := plugin.CreateInterceptorPlugins(ctx, processor.Metadata())
	if err != nil {
		return nil, err
	}

	interceptorPipeline := processor
	for _, interceptor := range interceptors {
		next := avframe.NewPipeline(interceptor)
		interceptorPipeline.AddNext(next, avframe.WithAllPayloadTypes())
		interceptorPipeline = next
	}

	interceptorPipeline.AddNext(ss, avframe.WithAllPayloadTypes())

	ss.tailProcessor = interceptorPipeline

	ss.logger.WithFields(map[string]interface{}{
		"format":   ss.Format(),
		"metadata": ss.Metadata(),
	}).Info("substream created")

	return ss, nil
}

func (ss *subStream) Metadata() avframe.Metadata {
	return ss.tailProcessor.Metadata()
}

func (ss *subStream) Subscribe(sub peer.Subscriber) (err error) {
	ss.logger.WithFields(map[string]interface{}{
		"sub": sub,
	}).Info("subscribe")

	ss.lock.Lock()
	defer ss.lock.Unlock()
	for _, s := range ss.subscribers {
		if s.subscriber == sub {
			return errors.New("subscriber already exists")
		}
	}
	ss.subscribers = append(ss.subscribers, subscriberInfo{
		subscriber: sub,
		active:     false,
	})

	return nil
}

func (ss *subStream) Unsubscribe(sub peer.Subscriber) {
	ss.logger.WithFields(map[string]interface{}{
		"sub": sub,
	}).Info("unsubscribe")

	ss.lock.Lock()
	defer ss.lock.Unlock()
	for i, s := range ss.subscribers {
		if s.subscriber == sub {
			ss.subscribers = append(ss.subscribers[:i], ss.subscribers[i+1:]...)
			break
		}
	}
	if len(ss.subscribers) == 0 {
		if ss.onSubscriberEmpty != nil {
			ss.onSubscriberEmpty(sub)
		}
	}
}

func (ss *subStream) Feedback(fb *avframe.Feedback) error {
	ss.logger.WithFields(map[string]interface{}{
		"fb": fb,
	}).Debug("feedback")

	if fb.Type != FeedbackTypeSubscriberActive {
		return ss.tailProcessor.Feedback(fb)
	}

	subActive := fb.Data.(*FeedbackSubscriberActive)
	if subActive.Subscriber == nil {
		return nil
	}

	ss.lock.RLock()
	defer ss.lock.RUnlock()

	for i, sub := range ss.subscribers {
		if sub.subscriber == subActive.Subscriber {
			ss.subscribers[i].active = subActive.Active
			break
		}
	}

	return nil
}

func (ss *subStream) Write(f *avframe.Frame) error {
	ss.lock.RLock()
	subs := []subscriberInfo{}
	subs = append(subs, ss.subscribers...)
	ss.lock.RUnlock()

	for _, sub := range subs {
		if sub.active {
			sub.subscriber.Write(f)
		}
	}

	return nil
}

func (ss *subStream) Read() (*avframe.Frame, error) {
	return nil, nil
}

func (ss *subStream) Format() avframe.FmtType {
	return ss.tailProcessor.Format()
}

func (ss *subStream) close() {
	ss.closeOnce.Do(func() {
		ss.logger.Info("substream closed")
		ss.subscribers = nil
		ss.cancel()
		if ss.onSubStreamClosed != nil {
			ss.onSubStreamClosed()
		}
	})
}

func (ss *subStream) Close() error {
	ss.logger.Info("closing substream")

	ss.lock.Lock()
	defer ss.lock.Unlock()
	ss.close()

	return nil
}

func (ss *subStream) CloseIfEmpty() bool {
	ss.lock.RLock()
	defer ss.lock.RUnlock()
	if len(ss.subscribers) == 0 {
		ss.logger.Info("substream empty, close it")
		ss.close()
		return true
	}
	return false
}

func (ss *subStream) SetDeadline(d time.Time) error {
	if ss.deadlineTimer != nil {
		ss.deadlineTimer.Stop()
	}

	ss.deadlineTimer = time.AfterFunc(time.Until(d), func() {
		ss.logger.Info("substream deadline expired")

		ss.CloseIfEmpty()
	})

	return nil
}
