package avframe

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"

	frameerror "github.com/pingostack/pingos/pkg/avframe/frame_error"
)

var AllPayloadTypes = []PayloadType{PayloadTypeAudio, PayloadTypeVideo, PayloadTypeData, PayloadTypeMetadata}

type Feedback struct {
	Type  string
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
	Metadata() Metadata
}

type pipelineNode struct {
	p            Processor
	PayloadTypes []PayloadType
}

func (p *pipelineNode) contains(payloadType PayloadType) bool {
	for _, pt := range p.PayloadTypes {
		if pt == payloadType {
			return true
		}
	}
	return false
}

type Pipeline struct {
	p         Processor
	nexts     []*pipelineNode
	prevs     []*pipelineNode
	lock      sync.RWMutex
	onceClose sync.Once
	closed    atomic.Int32
}

func NewPipeline(p Processor) *Pipeline {
	pl := &Pipeline{p: p}

	return pl
}

type PayloadTypesOption func(pl *Pipeline) []PayloadType

func WithPayloadType(payloadTypes ...PayloadType) PayloadTypesOption {
	return func(_ *Pipeline) []PayloadType {
		return payloadTypes
	}
}

func WithoutPayloadType(exceptPayloadTypes ...PayloadType) PayloadTypesOption {
	return func(_ *Pipeline) []PayloadType {
		result := []PayloadType{}
		for _, pt := range AllPayloadTypes {
			if !slices.Contains(exceptPayloadTypes, pt) {
				result = append(result, pt)
			}
		}
		return result
	}
}

func WithAllPayloadTypes() PayloadTypesOption {
	return func(_ *Pipeline) []PayloadType {
		return AllPayloadTypes
	}
}
func (pl *Pipeline) hasPrev(p Processor) bool {
	if pl.closed.Load() == 1 {
		return false
	}

	for _, prev := range pl.prevs {
		if prev.p == p {
			return true
		}
	}

	return false
}

func (pl *Pipeline) hasNext(p Processor) bool {
	if pl.closed.Load() == 1 {
		return false
	}

	for _, next := range pl.nexts {
		if next.p == p {
			return true
		}
	}

	return false
}

func (pl *Pipeline) AddPrev(prev *Pipeline, opt PayloadTypesOption) error {
	if pl.closed.Load() == 1 {
		return errors.New("pipeline closed")
	}

	if prev == nil {
		return errors.New("prev is nil")
	}

	pl.lock.Lock()
	defer pl.lock.Unlock()

	if pl.hasPrev(prev) {
		return errors.New("prev already exists")
	}

	if pl.hasNext(prev) {
		return errors.New("prev already exists")
	}

	pl.prevs = append(pl.prevs, &pipelineNode{p: prev, PayloadTypes: opt(pl)})
	return nil
}

func (pl *Pipeline) AddNext(sub Processor, opt PayloadTypesOption) error {
	if pl.closed.Load() == 1 {
		return errors.New("pipeline closed")
	}

	if sub == nil {
		return errors.New("sub is nil")
	}

	payloadTypes := opt(pl)

	pl.lock.Lock()
	defer pl.lock.Unlock()

	if pl.hasNext(sub) {
		return errors.New("sub already exists")
	}

	if pl.hasPrev(sub) {
		return errors.New("sub already exists")
	}

	pl.nexts = append(pl.nexts, &pipelineNode{p: sub, PayloadTypes: payloadTypes})

	nextPl, ok := sub.(*Pipeline)
	if ok {
		if err := nextPl.AddPrev(pl, opt); err != nil {
			return err
		}
	}

	return nil
}

func (pl *Pipeline) GetPrevs() []Processor {
	if pl.closed.Load() == 1 {
		return nil
	}

	pl.lock.RLock()
	defer pl.lock.RUnlock()

	prevs := make([]Processor, len(pl.prevs))
	for i, p := range pl.prevs {
		prevs[i] = p.p
	}

	return prevs
}

func (pl *Pipeline) RemoveNext(next Processor) {
	if pl.closed.Load() == 1 {
		return
	}

	pl.lock.Lock()
	defer pl.lock.Unlock()
	for i, n := range pl.nexts {
		if n.p == next {
			nextPl, ok := next.(*Pipeline)
			if ok {
				nextPl.DeletePrev(pl)
			}
			pl.nexts = append(pl.nexts[:i], pl.nexts[i+1:]...)
			break
		}
	}
}

func (pl *Pipeline) Write(frame *Frame) error {
	if pl.closed.Load() == 1 || pl.p == nil {
		return frameerror.ErrEOF
	}

	var f *Frame
	err := pl.p.Write(frame)
	if err != nil {
		if errors.Is(err, frameerror.ErrBreak) {
			return nil
		} else if errors.Is(err, frameerror.ErrDecline) {
			f = frame
		} else {
			return err
		}
	} else {
		f, err = pl.p.Read()
		if err != nil {
			if errors.Is(err, frameerror.ErrAgain) {
				return nil
			} else {
				return err
			}
		}
	}

	pl.lock.RLock()
	defer pl.lock.RUnlock()

	for _, next := range pl.nexts {
		if next.contains(f.PayloadType) {
			next.p.Write(f)
		}
	}

	return nil
}

// Read will read the frame from the pipeline
func (pl *Pipeline) Read() (*Frame, error) {
	if pl.closed.Load() == 1 || pl.p == nil {
		return nil, frameerror.ErrEOF
	}

	return pl.p.Read()
}

func (pl *Pipeline) Close() error {
	var err error
	pl.onceClose.Do(func() {
		pl.closed.Store(1)

		if pl.p != nil {
			err = pl.p.Close()
		}

		pl.lock.Lock()
		defer pl.lock.Unlock()

		pl.prevs = nil
		pl.nexts = nil
	})

	return err
}

// Feedback will feedback the frame to the previous pipeline
func (pl *Pipeline) Feedback(fb *Feedback) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("feedback panic: %v", r)
		}
	}()

	if pl.p != nil {
		if err := pl.p.Feedback(fb); err != nil {
			if errors.Is(err, frameerror.ErrBreak) {
				// do nothing, just break the frame process
				return nil
			} else if errors.Is(err, frameerror.ErrDecline) {
				// do nothing, just decline the frame
			} else {
				// return the error
				return err
			}
		}
	}

	prevs := pl.GetPrevs()
	if len(prevs) > 0 {
		for _, prev := range prevs {
			prev.Feedback(fb)
		}
	}

	return err
}

func (pl *Pipeline) Format() FmtType {
	return pl.p.Format()
}

func (pl *Pipeline) Metadata() Metadata {
	return pl.p.Metadata()
}

func (pl *Pipeline) DeletePrev(prev Processor) {
	pl.lock.Lock()
	defer pl.lock.Unlock()

	for i, p := range pl.prevs {
		if p.p == prev {
			pl.prevs = append(pl.prevs[:i], pl.prevs[i+1:]...)
			break
		}
	}
}

func (pl *Pipeline) CheckCycle() bool {
	return pl.checkPrevCycle() || pl.checkNextCycle()
}

func (pl *Pipeline) checkPrevCycle() bool {
	visited := make(map[Processor]bool)
	stack := make([]Processor, 0)

	var dfs func(Processor) bool
	dfs = func(p Processor) bool {
		if visited[p] {
			return false
		}
		if len(stack) > 0 && stack[len(stack)-1] == p {
			return true
		}
		visited[p] = true
		stack = append(stack, p)
		prevPl, ok := p.(*Pipeline)
		if ok {
			for _, prev := range prevPl.prevs {
				if dfs(prev.p) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		return false
	}

	for _, prev := range pl.prevs {
		if dfs(prev.p) {
			return true
		}
	}

	return false
}

func (pl *Pipeline) checkNextCycle() bool {
	visited := make(map[Processor]bool)
	stack := make([]Processor, 0)

	var dfs func(Processor) bool
	dfs = func(p Processor) bool {
		if visited[p] {
			return false
		}
		if len(stack) > 0 && stack[len(stack)-1] == p {
			return true
		}
		visited[p] = true
		stack = append(stack, p)
		nextPl, ok := p.(*Pipeline)
		if ok {
			for _, next := range nextPl.nexts {
				if dfs(next.p) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		return false
	}

	for _, next := range pl.nexts {
		if dfs(next.p) {
			return true
		}
	}

	return false
}

func (pl *Pipeline) ChainLength() int {
	length := 0
	for _, next := range pl.nexts {
		length++
		if nextPl, ok := next.p.(*Pipeline); ok {
			length += nextPl.ChainLength()
		}
	}
	return length
}
