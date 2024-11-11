package eventemitter

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

type EventID int

type eventFunc func(data interface{}) (interface{}, error)

type EventEmitter interface {
	On(eventID EventID, f eventFunc)
	EmitEvent(ctx context.Context, eventID EventID, data interface{}) (err error)
	EmitEventWithResult(ctx context.Context, eventID EventID, data interface{}) (result *EventResult, err error)
}

type EventResult struct {
	Error error
	Data  interface{}
}

type Event struct {
	Signal EventID
	Result chan EventResult
	Data   interface{}
}

func (e Event) String() string {
	return fmt.Sprintf("{eventID: %d, data: %+v}", e.Signal, e.Data)
}

type EventEmitterImpl struct {
	oneventLock sync.RWMutex
	eventCh     chan Event
	listeners   map[EventID][]eventFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	signalCounter atomic.Int32
)

func GenEventID() EventID {
	return EventID(signalCounter.Inc())
}

func NewEventEmitter(ctx context.Context, size int) EventEmitter {
	m := &EventEmitterImpl{
		eventCh:   make(chan Event, size),
		listeners: make(map[EventID][]eventFunc),
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	go m.run()

	return m
}

func (m *EventEmitterImpl) EmitEvent(ctx context.Context, eventID EventID, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("event emitter panic: %v", r)
		}
	}()

	e := Event{
		Signal: eventID,
		Data:   data,
	}

	select {
	case m.eventCh <- e:
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
		return fmt.Errorf("Event queue full")
	}

	return nil
}

func (m *EventEmitterImpl) EmitEventWithResult(ctx context.Context, eventID EventID, data interface{}) (result *EventResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("event emitter panic: %v", r)
		}
	}()

	e := Event{
		Signal: eventID,
		Data:   data,
		Result: make(chan EventResult, 1),
	}

	select {
	case m.eventCh <- e:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	default:
		return nil, fmt.Errorf("event queue full")
	}

	select {
	case r, ok := <-e.Result:
		if !ok {
			return nil, fmt.Errorf("event result channel closed")
		}
		return &r, nil
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *EventEmitterImpl) SyncEmitEvent(eventID EventID, data interface{}) error {
	m.oneventLock.RLock()
	listeners, found := m.listeners[eventID]
	m.oneventLock.RUnlock()

	if !found {
		return nil
	}

	for _, f := range listeners {
		_, err := f(data)
		if err != nil {
			return errors.Wrap(err, "SyncEmitEvent")
		}
	}

	return nil
}

func (m *EventEmitterImpl) run() {
	defer func() {
		close(m.eventCh)
	}()

	for {
		select {
		case <-m.ctx.Done():
			return

		case e, ok := <-m.eventCh:
			if !ok {
				return
			}

			m.oneventLock.RLock()
			listeners, found := m.listeners[e.Signal]
			m.oneventLock.RUnlock()

			if !found {
				continue
			}

			for _, f := range listeners {
				go func(f eventFunc) {
					result, err := f(e.Data)
					e.Result <- EventResult{
						Error: err,
						Data:  result,
					}
				}(f)
			}
		}
	}
}

func (m *EventEmitterImpl) On(eventID EventID, f eventFunc) {
	m.oneventLock.Lock()
	defer m.oneventLock.Unlock()

	listeners, found := m.listeners[eventID]
	if !found {
		listeners = make([]eventFunc, 0)
	}

	listeners = append(listeners, f)

	m.listeners[eventID] = listeners
}

func (m *EventEmitterImpl) Remove(eventID EventID) {
	m.oneventLock.Lock()
	defer m.oneventLock.Unlock()

	delete(m.listeners, eventID)
}

func (m *EventEmitterImpl) Close() {
	m.cancel()
}
