package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/peer"
	eventemitter "github.com/pingostack/pingos/pkg/eventemmiter"
	"github.com/pkg/errors"
)

type subStream struct {
	ctx               context.Context
	cancel            context.CancelFunc
	format            *Format
	subscribers       map[string]peer.Subscriber
	subLock           sync.RWMutex
	streamEventHandle eventemitter.EventEmitter
	closeOnce         sync.Once
}

func newSubStream(ctx context.Context, ee eventemitter.EventEmitter, format *Format) (*subStream, error) {
	if format == nil {
		return nil, errors.New("format is nil")
	}

	ctx, cancel := context.WithCancel(ctx)
	return &subStream{
		ctx:               ctx,
		cancel:            cancel,
		streamEventHandle: ee,
		format:            format,
		subscribers:       make(map[string]peer.Subscriber),
	}, nil
}

func (ss *subStream) Subscribe(sub peer.Subscriber) {
	ss.subLock.Lock()
	defer ss.subLock.Unlock()
	ss.subscribers[sub.ID()] = sub
}

func (ss *subStream) Unsubscribe(sub peer.Subscriber) {
	ss.subLock.Lock()
	defer ss.subLock.Unlock()
	delete(ss.subscribers, sub.ID())
	if len(ss.subscribers) == 0 {
		ss.streamEventHandle.EmitEvent(ss.ctx, EventSubstreamEmpty, ss)
	}
}

func (ss *subStream) Feedback(fb *mode.Feedback) error {
	fmt.Printf("substream feedback: %+v\n", fb)

	return ss.format.Feedback(fb)
}

func (ss *subStream) Write(f *mode.Frame) error {
	frame, err := ss.format.Process(f)
	if err != nil {
		return errors.Wrap(err, "process frame")
	}

	ss.subLock.RLock()
	defer ss.subLock.RUnlock()
	for _, sub := range ss.subscribers {
		sub.Write(frame)
	}
	return nil
}

func (ss *subStream) GetFormatType() mode.FmtType {
	return ss.format.fmtType
}

func (ss *subStream) Close() {
	ss.closeOnce.Do(func() {
		ss.cancel()
		ss.subLock.Lock()
		defer ss.subLock.Unlock()
		ss.subscribers = nil
	})
}
