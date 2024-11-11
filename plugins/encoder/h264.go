package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterEncoder(&plugin.EncoderPlugin{
		CodecType: mode.CodecTypeH264,
		CreateEncoder: func(ctx context.Context) (plugin.Encoder, error) {
			return NewH264Encoder(ctx)
		},
	})
}

type H264Encoder struct {
	ctx context.Context
}

func NewH264Encoder(ctx context.Context) (plugin.Encoder, error) {
	return &H264Encoder{ctx: ctx}, nil
}

func (e *H264Encoder) Close() error {
	return nil
}

func (e *H264Encoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeH264
}

func (e *H264Encoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	frame.Codec = mode.CodecTypeH264
	fmt.Printf("h264 encoder process frame: %+v\n", frame)
	return frame, nil
}

func (e *H264Encoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("h264 encoder feedback: %+v\n", fb)
	return nil
}
