package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterDecoder(&plugin.DecoderPlugin{
		CodecType: mode.CodecTypeAAC,
		CreateDecoder: func(ctx context.Context) (plugin.Decoder, error) {
			return NewAACDecoder(ctx)
		},
	})
}

type AACDecoder struct {
	ctx context.Context
}

func NewAACDecoder(ctx context.Context) (plugin.Decoder, error) {
	return &AACDecoder{ctx: ctx}, nil
}

func (d *AACDecoder) Close() error {
	return nil
}

func (d *AACDecoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeAAC
}

func (d *AACDecoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("aac decoder process frame: %+v\n", frame)
	return frame, nil
}

func (d *AACDecoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("aac decoder feedback: %+v\n", fb)
	return nil
}
