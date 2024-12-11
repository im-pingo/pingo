package avframe

type NoopProcessor struct {
	metadata Metadata
	f        *Frame
}

func NewNoopProcessor(metadata Metadata) *NoopProcessor {
	return &NoopProcessor{metadata: metadata}
}

func (n *NoopProcessor) Feedback(feedback *Feedback) error {
	return nil
}

func (n *NoopProcessor) Format() FmtType {
	return n.metadata.FmtType
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

func (n *NoopProcessor) Metadata() Metadata {
	return n.metadata
}

func (n *NoopProcessor) UpdateSourceMetadata(metadata Metadata) {
	n.metadata = metadata
}
