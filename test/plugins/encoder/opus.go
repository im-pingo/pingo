package encoder

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterEncoderPlugin(avframe.CodecTypeOPUS, func(ctx context.Context) (plugin.Encoder, error) {
		return NewOPUSEncoder(ctx)
	})
}

type OPUSEncoder struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewOPUSEncoder(ctx context.Context) (plugin.Encoder, error) {
	return &OPUSEncoder{ctx: ctx}, nil
}

func (e *OPUSEncoder) Close() error {
	return nil
}

func (e *OPUSEncoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeOPUS
}

func (e *OPUSEncoder) Read() (*avframe.Frame, error) {
	return e.frame, nil
}

func (e *OPUSEncoder) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("opus encoder write frame", frame)
	e.frame = frame
	return nil
}

func (e *OPUSEncoder) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (e *OPUSEncoder) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("opus encoder feedback: %+v\n", fb)
	return nil
}
