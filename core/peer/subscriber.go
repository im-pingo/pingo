package peer

import "github.com/pingostack/pingos/pkg/avframe"

type Subscriber interface {
	avframe.WriteCloser
	// ID returns the unique identifier of the subscriber
	ID() string
	Format() avframe.FmtType
	AudioCodecSupported() []avframe.CodecType
	VideoCodecSupported() []avframe.CodecType
	SetProcessor(processor avframe.Processor)
}
