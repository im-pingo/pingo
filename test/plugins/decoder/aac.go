package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterDecoderPlugin(avframe.CodecTypeAAC, func(ctx context.Context) (plugin.Decoder, error) {
		return NewAACDecoder(ctx)
	})
}

type AACDecoder struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewAACDecoder(ctx context.Context) (plugin.Decoder, error) {
	return &AACDecoder{ctx: ctx}, nil
}

func (d *AACDecoder) Close() error {
	return nil
}

func (d *AACDecoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeAAC
}

func (d *AACDecoder) Read() (*avframe.Frame, error) {
	fmt.Println("aac decoder read frame", d.frame)
	return d.frame, nil
}

func (d *AACDecoder) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("aac decoder write frame", frame)
	d.frame = frame
	return nil
}

func (d *AACDecoder) Format() avframe.FmtType {
	return avframe.FormatAudioSample
}

func (d *AACDecoder) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("aac decoder feedback: %+v\n", fb)
	return nil
}
