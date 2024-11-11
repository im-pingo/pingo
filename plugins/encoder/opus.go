package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterEncoder(&plugin.EncoderPlugin{
		CodecType: mode.CodecTypeOPUS,
		CreateEncoder: func(ctx context.Context) (plugin.Encoder, error) {
			return NewOPUSEncoder(ctx)
		},
	})
}

type OPUSEncoder struct {
	ctx context.Context
}

func NewOPUSEncoder(ctx context.Context) (plugin.Encoder, error) {
	return &OPUSEncoder{ctx: ctx}, nil
}

func (e *OPUSEncoder) Close() error {
	return nil
}

func (e *OPUSEncoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeOPUS
}

func (e *OPUSEncoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("opus encoder process frame: %+v\n", frame)
	return frame, nil
}

func (e *OPUSEncoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("opus encoder feedback: %+v\n", fb)
	return nil
}
