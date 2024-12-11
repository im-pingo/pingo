package encoder

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	frameerror "github.com/pingostack/pingos/pkg/avframe/frame_error"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterEncoderPlugin(avframe.CodecTypeOPUS, func(ctx context.Context, metadata avframe.Metadata) (plugin.Encoder, error) {
		return NewOPUSEncoder(ctx, metadata)
	})
}

type OPUSEncoder struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func NewOPUSEncoder(ctx context.Context, metadata avframe.Metadata) (plugin.Encoder, error) {
	return &OPUSEncoder{ctx: ctx, inMetadata: metadata}, nil
}

func (e *OPUSEncoder) Close() error {
	logger.Info("opus encoder close")
	return nil
}

func (e *OPUSEncoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeOPUS
}

func (e *OPUSEncoder) Read() (*avframe.Frame, error) {
	e.frame.Fmt = avframe.FormatRaw
	e.frame.PayloadType = avframe.PayloadTypeAudio
	e.frame.WriteAudioHeader(&avframe.AudioHeader{
		Codec: avframe.CodecTypeOPUS,
		Rate:  44100,
		Bits:  16,
	})
	return e.frame, nil
}

func (e *OPUSEncoder) Write(frame *avframe.Frame) error {
	if !frame.IsAudio() {
		return frameerror.ErrBreak
	}

	frame.TTL++
	logger.Info("opus encoder write frame", frame)
	e.frame = frame
	return nil
}

func (e *OPUSEncoder) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (e *OPUSEncoder) Feedback(fb *avframe.Feedback) error {
	logger.Infof("opus encoder feedback: %+v\n", fb)
	return nil
}

func (e *OPUSEncoder) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatRaw,
		AudioCodecType: avframe.CodecTypeOPUS,
	}
}

func (e *OPUSEncoder) UpdateSourceMetadata(metadata avframe.Metadata) {
	e.inMetadata = metadata
}
