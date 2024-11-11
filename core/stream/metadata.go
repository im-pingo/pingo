package stream

import "github.com/pingostack/pingos/core/mode"

type Metadata struct {
	AudioCodecType mode.CodecType
	VideoCodecType mode.CodecType
	FmtType        mode.FmtType
}
