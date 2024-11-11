package muxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterMuxer(&plugin.MuxerPlugin{
		FmtType: mode.FormatRtmp,
		CreateMuxer: func(ctx context.Context) (plugin.Muxer, error) {
			return NewRTMPMuxer(ctx)
		},
	})
}

type RTMPMuxer struct {
	ctx context.Context
}

func NewRTMPMuxer(ctx context.Context) (plugin.Muxer, error) {
	return &RTMPMuxer{ctx: ctx}, nil
}

func (m *RTMPMuxer) Close() error {
	return nil
}

func (m *RTMPMuxer) GetFormatType() mode.FmtType {
	return mode.FormatRtmp
}

func (m *RTMPMuxer) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("rtmp muxer process frame: %+v\n", frame)
	return frame, nil
}

func (m *RTMPMuxer) Feedback(fb *mode.Feedback) error {
	fmt.Printf("rtmp muxer feedback: %+v\n", fb)
	return nil
}
