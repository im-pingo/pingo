package encoder

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	frameerror "github.com/pingostack/pingos/pkg/avframe/frame_error"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterEncoderPlugin(avframe.CodecTypeAAC, func(ctx context.Context, metadata avframe.Metadata) (plugin.Encoder, error) {
		return NewAACEncoder(ctx, metadata)
	})
}

type AACEncoder struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func NewAACEncoder(ctx context.Context, metadata avframe.Metadata) (plugin.Encoder, error) {
	return &AACEncoder{ctx: ctx, inMetadata: metadata}, nil
}

func (e *AACEncoder) Close() error {
	logger.Info("aac encoder close")
	return nil
}

func (e *AACEncoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeAAC
}

func (e *AACEncoder) Read() (*avframe.Frame, error) {
	logger.Info("aac encoder read frame")
	e.frame.Fmt = avframe.FormatRaw
	e.frame.PayloadType = avframe.PayloadTypeAudio
	e.frame.WriteAudioHeader(&avframe.AudioHeader{
		Codec: avframe.CodecTypeAAC,
		Rate:  44100,
		Bits:  16,
	})
	return e.frame, nil
}

func (e *AACEncoder) Write(frame *avframe.Frame) error {
	if !frame.IsAudio() {
		return frameerror.ErrBreak
	}

	frame.TTL++
	logger.Info("aac encoder write frame", frame)
	e.frame = frame
	return nil
}

func (e *AACEncoder) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (e *AACEncoder) Feedback(fb *avframe.Feedback) error {
	logger.Infof("aac encoder feedback: %+v\n", fb)
	return nil
}

func (e *AACEncoder) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatRaw,
		AudioCodecType: avframe.CodecTypeAAC,
	}
}

func (e *AACEncoder) UpdateSourceMetadata(metadata avframe.Metadata) {
	e.inMetadata = metadata
}
