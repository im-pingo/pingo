package stream

import (
	"context"
	"sync"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

type subStream struct {
	ctx               context.Context
	cancel            context.CancelFunc
	subscribers       []peer.Subscriber
	subLock           sync.RWMutex
	closeOnce         sync.Once
	processor         *avframe.Transmit
	tailProcessor     *avframe.Transmit
	onSubscriberEmpty func(sub peer.Subscriber)
}

type SubStreamOption func(ss *subStream)

func WithOnSubscriberEmpty(fn func(sub peer.Subscriber)) SubStreamOption {
	return func(ss *subStream) {
		ss.onSubscriberEmpty = fn
	}
}

func newSubStream(ctx context.Context, processor *avframe.Transmit, opts ...SubStreamOption) (ss *subStream, err error) {
	ctx, cancel := context.WithCancel(ctx)
	ss = &subStream{
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make([]peer.Subscriber, 0),
		processor:   processor,
	}

	interceptors, err := plugin.CreateInterceptorPlugins(ctx, processor.Format())
	if err != nil {
		return nil, err
	}

	interceptorTransmit := processor
	for _, interceptor := range interceptors {
		next := avframe.NewTransmit(interceptor)
		interceptorTransmit.AddNextAudioTransmit(next)
		interceptorTransmit.AddNextVideoTransmit(next)
		interceptorTransmit = next
	}

	interceptorTransmit.AddAudioSubscriber(ss)
	interceptorTransmit.AddVideoSubscriber(ss)

	ss.tailProcessor = interceptorTransmit

	for _, opt := range opts {
		opt(ss)
	}

	return ss, nil
}

func (ss *subStream) Subscribe(sub peer.Subscriber) {
	ss.subLock.Lock()
	defer ss.subLock.Unlock()
	for _, s := range ss.subscribers {
		if s == sub {
			return
		}
	}
	ss.subscribers = append(ss.subscribers, sub)
	sub.SetProcessor(ss.tailProcessor)
}

func (ss *subStream) Unsubscribe(sub peer.Subscriber) {
	ss.subLock.Lock()
	defer ss.subLock.Unlock()
	for i, s := range ss.subscribers {
		if s == sub {
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
	return ss.tailProcessor.Feedback(fb)
}

func (ss *subStream) Write(f *avframe.Frame) error {
	ss.subLock.RLock()
	subs := []peer.Subscriber{}
	subs = append(subs, ss.subscribers...)
	ss.subLock.RUnlock()

	for _, sub := range subs {
		sub.Write(f)
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
		ss.subscribers = nil
		ss.cancel()
	})
}

func (ss *subStream) Close() error {
	ss.subLock.Lock()
	defer ss.subLock.Unlock()
	ss.close()

	return nil
}

func (ss *subStream) CloseIfEmpty() bool {
	ss.subLock.RLock()
	defer ss.subLock.RUnlock()
	if len(ss.subscribers) == 0 {
		ss.close()
		return true
	}

	return false
}
