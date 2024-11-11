package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

type Vp8Encoder struct {
	ctx context.Context
}

func (e *Vp8Encoder) Close() error {
	return nil
}

func (e *Vp8Encoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("vp8 encode frame: %+v\n", frame)
	return frame, nil
}

func (e *Vp8Encoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeVP8
}

func (e *Vp8Encoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("vp8 encoder feedback: %+v\n", fb)
	return nil
}

func init() {
	plugin.RegisterEncoder(&plugin.EncoderPlugin{
		CreateEncoder: func(ctx context.Context) (plugin.Encoder, error) {
			return &Vp8Encoder{ctx: ctx}, nil
		},
		CodecType: mode.CodecTypeVP8,
	})
}
