package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterDecoder(&plugin.DecoderPlugin{
		CodecType: mode.CodecTypeOPUS,
		CreateDecoder: func(ctx context.Context) (plugin.Decoder, error) {
			return NewOPUSDecoder(ctx)
		},
	})
}

type OPUSDecoder struct {
	ctx context.Context
}

func NewOPUSDecoder(ctx context.Context) (plugin.Decoder, error) {
	return &OPUSDecoder{ctx: ctx}, nil
}

func (d *OPUSDecoder) Close() error {
	return nil
}

func (d *OPUSDecoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeOPUS
}

func (d *OPUSDecoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("opus decoder process frame: %+v\n", frame)
	return frame, nil
}

func (d *OPUSDecoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("opus decoder feedback: %+v\n", fb)
	return nil
}
