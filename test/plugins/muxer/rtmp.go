package muxer

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterMuxerPlugin(avframe.FormatRtmp, func(ctx context.Context, metadata avframe.Metadata) (plugin.Muxer, error) {
		return NewRTMPMuxer(ctx, metadata)
	})
}

type RTMPMuxer struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func NewRTMPMuxer(ctx context.Context, metadata avframe.Metadata) (plugin.Muxer, error) {
	return &RTMPMuxer{ctx: ctx, inMetadata: metadata}, nil
}

func (m *RTMPMuxer) Close() error {
	logger.Info("rtmp muxer close")
	return nil
}

func (m *RTMPMuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtmp
}

func (m *RTMPMuxer) Feedback(fb *avframe.Feedback) error {
	logger.Infof("rtmp muxer feedback: %+v\n", fb)
	return nil
}

func (m *RTMPMuxer) Read() (*avframe.Frame, error) {
	logger.Info("rtmp muxer read frame", m.frame)
	m.frame.Fmt = avframe.FormatRtmp
	return m.frame, nil
}

func (m *RTMPMuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	logger.Info("rtmp muxer write frame", frame)
	m.frame = frame
	return nil
}

func (m *RTMPMuxer) Format() avframe.FmtType {
	return avframe.FormatRtmp
}

func (m *RTMPMuxer) UpdateSourceMetadata(metadata avframe.Metadata) {
	m.inMetadata = metadata
}

func (m *RTMPMuxer) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        m.inMetadata.FmtType,
		AudioCodecType: m.inMetadata.AudioCodecType,
		VideoCodecType: m.inMetadata.VideoCodecType,
	}
}
