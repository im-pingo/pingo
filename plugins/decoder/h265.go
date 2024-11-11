package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

type H265Decoder struct {
}

func init() {
	plugin.RegisterDecoder(&plugin.DecoderPlugin{
		CodecType: mode.CodecTypeH265,
		CreateDecoder: func(ctx context.Context) (plugin.Decoder, error) {
			return &H265Decoder{}, nil
		},
	})
}
func (d *H265Decoder) Close() error {
	return nil
}

func (d *H265Decoder) GetCodecType() mode.CodecType {
	return mode.CodecTypeH265
}

func (d *H265Decoder) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.Codec = mode.CodecTypeYUV
	frame.TTL++
	fmt.Printf("h265 decoder process frame: %+v\n", frame)
	return frame, nil
}

func (d *H265Decoder) Feedback(fb *mode.Feedback) error {
	fmt.Printf("h265 decoder feedback: %+v\n", fb)
	return nil
}
