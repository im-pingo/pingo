package peer

import (
	"github.com/pingostack/pingos/core/mode"
)

type Subscriber interface {
	mode.WriteCloser
	// ID returns the unique identifier of the subscriber
	ID() string
	Format() mode.FmtType
	AudioCodecSupported() []mode.CodecType
	VideoCodecSupported() []mode.CodecType
}
