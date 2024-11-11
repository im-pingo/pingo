package mode

import "sync"

type Writer interface {
	Write(frame *Frame) error
}

type Reader interface {
	Read() (*Frame, error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Closer interface {
	Close() error
}

type WriteCloser interface {
	Writer
	Closer
}

type ReadCloser interface {
	Reader
	Closer
}

type ReadWriterCloser interface {
	ReadWriter
	Closer
}

type Feedback struct {
	ID   string
	Data interface{}
}

type FeedbackListener interface {
	Feedback(feedback *Feedback) error
}

type Processor interface {
	FeedbackListener
	Process(frame *Frame) (*Frame, error)
}

type Transmit struct {
	p     Processor
	nexts []*Transmit
	prev  *Transmit
	lock  sync.RWMutex
}

func NewTransmit(p Processor) *Transmit {
	return &Transmit{p: p}
}

func (t *Transmit) AppendNext(next *Transmit) {
	if next == nil {
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.nexts = append(t.nexts, next)
	next.prev = t
}

func (t *Transmit) Prev() *Transmit {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.prev
}

func (t *Transmit) Delete(next *Transmit) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for i, n := range t.nexts {
		if n == next {
			next.prev = nil
			t.nexts = append(t.nexts[:i], t.nexts[i+1:]...)
			break
		}
	}
}

func (t *Transmit) Process(frame *Frame) (*Frame, error) {
	if t.p != nil {
		result, err := t.p.Process(frame)
		if err != nil {
			return nil, err
		}
		for _, next := range t.nexts {
			_, err := next.Process(result)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	return frame, nil
}

func (t *Transmit) Feedback(fb *Feedback) error {
	if t.p != nil {
		if err := t.p.Feedback(fb); err != nil {
			return err
		}
	}

	if t.prev != nil {
		return t.prev.Feedback(fb)
	}

	return nil
}

func CreateTransmit(processors ...Processor) (head *Transmit, tail *Transmit) {
	var tailTransmit *Transmit
	var headTransmit *Transmit

	appendTransmit := func(transmit *Transmit) {
		if headTransmit == nil {
			headTransmit = transmit
			tailTransmit = transmit
		} else if tailTransmit != nil {
			tailTransmit.AppendNext(transmit)
			tailTransmit = transmit
		}
	}

	for _, p := range processors {
		appendTransmit(NewTransmit(p))
	}

	return headTransmit, tailTransmit
}
