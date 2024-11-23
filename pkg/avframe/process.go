package avframe

import (
	"fmt"
	"io"
	"sync"
)

type Feedback struct {
	ID    string
	Video bool
	Audio bool
	Data  interface{}
}

type FeedbackListener interface {
	Feedback(feedback *Feedback) error
}

type Processor interface {
	FeedbackListener
	ReadWriteCloser
	Format() FmtType
}

type Transmit struct {
	p                  Processor
	nextAudioTransmits []*Transmit
	nextVideoTransmits []*Transmit
	prevAudioTransmit  *Transmit
	prevVideoTransmit  *Transmit
	audioSubs          []WriteCloser
	videoSubs          []WriteCloser
	lock               sync.RWMutex
}

func NewTransmit(p Processor) *Transmit {
	t := &Transmit{p: p}

	return t
}

func (t *Transmit) AddNextAudioTransmit(next *Transmit) {
	if next == nil {
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.nextAudioTransmits = append(t.nextAudioTransmits, next)
	next.prevAudioTransmit = t
}

func (t *Transmit) AddNextVideoTransmit(next *Transmit) {
	if next == nil {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.nextVideoTransmits = append(t.nextVideoTransmits, next)
	next.prevVideoTransmit = t
}

func (t *Transmit) AddAudioSubscriber(subscriber Processor) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.audioSubs = append(t.audioSubs, subscriber)
}

func (t *Transmit) AddVideoSubscriber(subscriber Processor) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.videoSubs = append(t.videoSubs, subscriber)
}

func (t *Transmit) PrevAudioTransmit() *Transmit {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.prevAudioTransmit
}

func (t *Transmit) PrevVideoTransmit() *Transmit {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.prevVideoTransmit
}

func (t *Transmit) Delete(next *Transmit) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for i, n := range t.nextAudioTransmits {
		if n == next {
			next.prevAudioTransmit = nil
			t.nextAudioTransmits = append(t.nextAudioTransmits[:i], t.nextAudioTransmits[i+1:]...)
			break
		}
	}

	for i, n := range t.nextVideoTransmits {
		if n == next {
			next.prevVideoTransmit = nil
			t.nextVideoTransmits = append(t.nextVideoTransmits[:i], t.nextVideoTransmits[i+1:]...)
			break
		}
	}
}

func (t *Transmit) Write(frame *Frame) error {
	if t.p == nil {
		return io.EOF
	}
	err := t.p.Write(frame)
	if err != nil {
		return err
	}

	f, err := t.p.Read()
	if err != nil {
		return err
	}

	audioSubs := []WriteCloser{}
	videoSubs := []WriteCloser{}
	nextAudioTransmits := []*Transmit{}
	nextVideoTransmits := []*Transmit{}
	t.lock.RLock()
	audioSubs = append(audioSubs, t.audioSubs...)
	videoSubs = append(videoSubs, t.videoSubs...)
	nextAudioTransmits = append(nextAudioTransmits, t.nextAudioTransmits...)
	nextVideoTransmits = append(nextVideoTransmits, t.nextVideoTransmits...)
	t.lock.RUnlock()

	if f.IsAudio() {
		for _, next := range nextAudioTransmits {
			next.Write(f)
		}
	} else {
		for _, next := range nextVideoTransmits {
			next.Write(f)
		}
	}

	if f.IsAudio() {
		for _, sub := range audioSubs {
			sub.Write(f)
		}
	} else {
		for _, sub := range videoSubs {
			sub.Write(f)
		}
	}

	return nil
}

func (t *Transmit) Read() (*Frame, error) {
	if t.p == nil {
		return nil, io.EOF
	}
	return t.p.Read()
}

func (t *Transmit) Close() error {
	if t.p != nil {
		return t.p.Close()
	}
	return nil
}

func (t *Transmit) Feedback(fb *Feedback) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("feedback panic: %v", r)
		}
	}()

	if t.p != nil {
		if err := t.p.Feedback(fb); err != nil {
			return err
		}
	}

	if t.prevAudioTransmit != nil && fb.Audio {
		err = t.prevAudioTransmit.Feedback(fb)
	}

	if t.prevVideoTransmit != nil && fb.Video {
		err = t.prevVideoTransmit.Feedback(fb)
	}

	return err
}

func (t *Transmit) Format() FmtType {
	return t.p.Format()
}
