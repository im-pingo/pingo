package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterEncoder(&plugin.EncoderPlugin{
		CodecType: mode.CodecTypeAAC,
		CreateEncoder: func(ctx context.Context) (plugin.Encoder, error) {
			return NewAACEncoder(ctx)
		},
	})
}

type AACEncoder struct {
	ctx context.Context
}

func NewAACEncoder(ctx context.Context) (plugin.Encoder, error) {
	return &AACEncoder{ctx: ctx}, nil
}

func (e *AACEncoder) Close() error {
	return nil
}

func (e *AACEncoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeAAC
}

func (e *AACEncoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("aac encoder process frame: %+v\n", frame)
	return frame, nil
}

func (e *AACEncoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("aac encoder feedback: %+v\n", fb)
	return nil
}
