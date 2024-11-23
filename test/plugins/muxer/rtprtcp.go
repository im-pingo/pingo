package muxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

type RtpRtcpMuxer struct {
	ctx   context.Context
	frame *avframe.Frame
}

func (m *RtpRtcpMuxer) Close() error {
	return nil
}

func (m *RtpRtcpMuxer) Read() (*avframe.Frame, error) {
	fmt.Println("rtprtcp muxer read frame")
	return m.frame, nil
}

func (m *RtpRtcpMuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("rtprtcp muxer write frame", frame)
	m.frame = frame
	return nil
}

func (m *RtpRtcpMuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func (m *RtpRtcpMuxer) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("rtprtcp muxer feedback: %+v\n", fb)
	return nil
}

func (m *RtpRtcpMuxer) Format() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func init() {
	plugin.RegisterMuxerPlugin(avframe.FormatRtpRtcp, func(ctx context.Context) (plugin.Muxer, error) {
		return &RtpRtcpMuxer{ctx: ctx}, nil
	})
}
