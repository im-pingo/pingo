package muxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

type RtpRtcpMuxer struct {
	ctx context.Context
}

func (m *RtpRtcpMuxer) Close() error {
	return nil
}

func (m *RtpRtcpMuxer) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("rtprtcp muxer process frame: %+v\n", frame)
	return frame, nil
}

func (m *RtpRtcpMuxer) GetFormatType() mode.FmtType {
	return mode.FormatRtpRtcp
}

func (m *RtpRtcpMuxer) Feedback(fb *mode.Feedback) error {
	fmt.Printf("rtprtcp muxer feedback: %+v\n", fb)
	return nil
}

func init() {
	plugin.RegisterMuxer(&plugin.MuxerPlugin{
		CreateMuxer: func(ctx context.Context) (plugin.Muxer, error) {
			return &RtpRtcpMuxer{ctx: ctx}, nil
		},
		FmtType: mode.FormatRtpRtcp,
	})
}
