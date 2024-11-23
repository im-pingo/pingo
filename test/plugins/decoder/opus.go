package decoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterDecoderPlugin(avframe.CodecTypeOPUS, func(ctx context.Context) (plugin.Decoder, error) {
		return NewOPUSDecoder(ctx)
	})
}

type OPUSDecoder struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewOPUSDecoder(ctx context.Context) (plugin.Decoder, error) {
	return &OPUSDecoder{ctx: ctx}, nil
}

func (d *OPUSDecoder) Close() error {
	return nil
}

func (d *OPUSDecoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeOPUS
}

func (d *OPUSDecoder) Read() (*avframe.Frame, error) {
	fmt.Println("opus decoder read frame")
	return d.frame, nil
}

func (d *OPUSDecoder) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("opus decoder write frame", frame)
	d.frame = frame
	return nil
}

func (d *OPUSDecoder) Format() avframe.FmtType {
	return avframe.FormatAudioSample
}

func (d *OPUSDecoder) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("opus decoder feedback: %+v\n", fb)
	return nil
}
