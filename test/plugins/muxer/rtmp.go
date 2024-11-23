package muxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterMuxerPlugin(avframe.FormatRtmp, func(ctx context.Context) (plugin.Muxer, error) {
		return NewRTMPMuxer(ctx)
	})
}

type RTMPMuxer struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewRTMPMuxer(ctx context.Context) (plugin.Muxer, error) {
	return &RTMPMuxer{ctx: ctx}, nil
}

func (m *RTMPMuxer) Close() error {
	return nil
}

func (m *RTMPMuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtmp
}

func (m *RTMPMuxer) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("rtmp muxer feedback: %+v\n", fb)
	return nil
}

func (m *RTMPMuxer) Read() (*avframe.Frame, error) {
	fmt.Println("rtmp muxer read frame", m.frame)
	return m.frame, nil
}

func (m *RTMPMuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("rtmp muxer write frame", frame)
	m.frame = frame
	return nil
}

func (m *RTMPMuxer) Format() avframe.FmtType {
	return avframe.FormatRtmp
}
