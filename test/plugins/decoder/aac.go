package decoder

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	frameerror "github.com/pingostack/pingos/pkg/avframe/frame_error"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterDecoderPlugin(avframe.CodecTypeAAC, func(ctx context.Context, metadata avframe.Metadata) (plugin.Decoder, error) {
		return NewAACDecoder(ctx, metadata)
	})
}

type AACDecoder struct {
	ctx        context.Context
	frame      *avframe.Frame
	inMetadata avframe.Metadata
}

func NewAACDecoder(ctx context.Context, metadata avframe.Metadata) (plugin.Decoder, error) {
	return &AACDecoder{ctx: ctx, inMetadata: metadata}, nil
}

func (d *AACDecoder) Close() error {
	logger.Info("aac decoder close")
	return nil
}

func (d *AACDecoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeAAC
}

func (d *AACDecoder) Read() (*avframe.Frame, error) {
	logger.Info("aac decoder read frame", d.frame)
	d.frame.Fmt = avframe.FormatAudioSample
	d.frame.PayloadType = avframe.PayloadTypeAudio
	d.frame.WriteAudioHeader(&avframe.AudioHeader{
		Codec: avframe.CodecTypeAudioSample,
		Rate:  0,
		Bits:  0,
	})
	return d.frame, nil
}

func (d *AACDecoder) Write(frame *avframe.Frame) error {
	if !frame.IsAudio() {
		return frameerror.ErrBreak
	}

	frame.TTL++
	logger.Info("aac decoder write frame", frame)
	d.frame = frame
	return nil
}

func (d *AACDecoder) Format() avframe.FmtType {
	return avframe.FormatAudioSample
}

func (d *AACDecoder) Feedback(fb *avframe.Feedback) error {
	logger.Infof("aac decoder feedback: %+v", fb)
	return nil
}

func (d *AACDecoder) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatAudioSample,
		AudioCodecType: avframe.CodecTypeAAC,
	}
}

func (d *AACDecoder) UpdateSourceMetadata(metadata avframe.Metadata) {
	d.inMetadata = metadata
}
