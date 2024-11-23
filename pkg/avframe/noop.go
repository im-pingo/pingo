package avframe

type NoopProcessor struct {
	format FmtType
	f      *Frame
}

func NewNoopProcessor(format FmtType) *NoopProcessor {
	return &NoopProcessor{format: format}
}

func (n *NoopProcessor) Feedback(feedback *Feedback) error {
	return nil
}

func (n *NoopProcessor) Format() FmtType {
	return n.format
}

func (n *NoopProcessor) Write(f *Frame) error {
	n.f = f
	return nil
}

func (n *NoopProcessor) Read() (*Frame, error) {
	return n.f, nil
}

func (n *NoopProcessor) Close() error {
	return nil
}
