package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterDecoder(&plugin.DecoderPlugin{
		CodecType: mode.CodecTypeVP8,
		CreateDecoder: func(ctx context.Context) (plugin.Decoder, error) {
			return NewVP8Decoder(ctx)
		},
	})
}

type VP8Decoder struct {
	ctx context.Context
}

func NewVP8Decoder(ctx context.Context) (plugin.Decoder, error) {
	return &VP8Decoder{ctx: ctx}, nil
}

func (d *VP8Decoder) Close() error {
	return nil
}

func (d *VP8Decoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeVP8
}

func (d *VP8Decoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("vp8 decoder process frame: %+v\n", frame)
	return frame, nil
}

func (d *VP8Decoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("vp8 decoder feedback: %+v\n", fb)
	return nil
}
