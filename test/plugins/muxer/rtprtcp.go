package muxer

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterMuxerPlugin(avframe.FormatRtpRtcp, func(ctx context.Context, metadata avframe.Metadata) (plugin.Muxer, error) {
		return &RtpRtcpMuxer{ctx: ctx, inMetadata: metadata}, nil
	})
}

type RtpRtcpMuxer struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func (m *RtpRtcpMuxer) Close() error {
	logger.Info("rtprtcp muxer close")
	return nil
}

func (m *RtpRtcpMuxer) Read() (*avframe.Frame, error) {
	logger.Info("rtprtcp muxer read frame")
	m.frame.Fmt = avframe.FormatRtpRtcp
	return m.frame, nil
}

func (m *RtpRtcpMuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	logger.Info("rtprtcp muxer write frame", frame)
	m.frame = frame
	return nil
}

func (m *RtpRtcpMuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func (m *RtpRtcpMuxer) Feedback(fb *avframe.Feedback) error {
	logger.Infof("rtprtcp muxer feedback: %+v", fb)
	return nil
}

func (m *RtpRtcpMuxer) Format() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func (m *RtpRtcpMuxer) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatRtpRtcp,
		AudioCodecType: m.inMetadata.AudioCodecType,
		VideoCodecType: m.inMetadata.VideoCodecType,
	}
}

func (m *RtpRtcpMuxer) UpdateSourceMetadata(metadata avframe.Metadata) {
	m.inMetadata = metadata
}
