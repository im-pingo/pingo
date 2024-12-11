package decoder

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	frameerror "github.com/pingostack/pingos/pkg/avframe/frame_error"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterDecoderPlugin(avframe.CodecTypeOPUS, func(ctx context.Context, metadata avframe.Metadata) (plugin.Decoder, error) {
		return NewOPUSDecoder(ctx, metadata)
	})
}

type OPUSDecoder struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func NewOPUSDecoder(ctx context.Context, metadata avframe.Metadata) (plugin.Decoder, error) {
	return &OPUSDecoder{ctx: ctx, inMetadata: metadata}, nil
}

func (d *OPUSDecoder) Close() error {
	logger.Info("opus decoder close")
	return nil
}

func (d *OPUSDecoder) GetCodecType() avframe.CodecType {
	return avframe.CodecTypeOPUS
}

func (d *OPUSDecoder) Read() (*avframe.Frame, error) {
	d.frame.Fmt = avframe.FormatAudioSample
	d.frame.PayloadType = avframe.PayloadTypeAudio
	d.frame.WriteAudioHeader(&avframe.AudioHeader{
		Codec: avframe.CodecTypeAudioSample,
		Rate:  0,
		Bits:  0,
	})
	logger.Info("opus decoder read frame")
	return d.frame, nil
}

func (d *OPUSDecoder) Write(frame *avframe.Frame) error {
	if !frame.IsAudio() {
		return frameerror.ErrBreak
	}

	frame.TTL++
	logger.Info("opus decoder write frame", frame)
	d.frame = frame
	return nil
}

func (d *OPUSDecoder) Format() avframe.FmtType {
	return avframe.FormatAudioSample
}

func (d *OPUSDecoder) Feedback(fb *avframe.Feedback) error {
	logger.Infof("opus decoder feedback: %+v", fb)
	return nil
}

func (d *OPUSDecoder) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatAudioSample,
		AudioCodecType: avframe.CodecTypeOPUS,
	}
}

func (d *OPUSDecoder) UpdateSourceMetadata(metadata avframe.Metadata) {
	d.inMetadata = metadata
}

func (d *OPUSDecoder) SetIdle(idle bool) {

}
