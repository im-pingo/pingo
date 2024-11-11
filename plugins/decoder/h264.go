package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterDecoder(&plugin.DecoderPlugin{
		CodecType: mode.CodecTypeH264,
		CreateDecoder: func(ctx context.Context) (plugin.Decoder, error) {
			return NewH264Decoder(ctx)
		},
	})
}

type H264Decoder struct {
	ctx context.Context
}

func NewH264Decoder(ctx context.Context) (plugin.Decoder, error) {
	return &H264Decoder{ctx: ctx}, nil
}

func (d *H264Decoder) Close() error {
	return nil
}

func (d *H264Decoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeH264
}

func (d *H264Decoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("h264 decoder process frame: %+v\n", frame)
	return frame, nil
}

func (d *H264Decoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("h264 decoder feedback: %+v\n", fb)
	return nil
}
