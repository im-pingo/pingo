package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterEncoderPlugin(avframe.CodecTypeAAC, func(ctx context.Context) (plugin.Encoder, error) {
		return NewAACEncoder(ctx)
	})
}

type AACEncoder struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewAACEncoder(ctx context.Context) (plugin.Encoder, error) {
	return &AACEncoder{ctx: ctx}, nil
}

func (e *AACEncoder) Close() error {
	return nil
}

func (e *AACEncoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeAAC
}

func (e *AACEncoder) Read() (*avframe.Frame, error) {
	fmt.Println("aac encoder read frame")
	return e.frame, nil
}

func (e *AACEncoder) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("aac encoder write frame", frame)
	e.frame = frame
	return nil
}

func (e *AACEncoder) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (e *AACEncoder) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("aac encoder feedback: %+v\n", fb)
	return nil
}
