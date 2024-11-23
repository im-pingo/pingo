package demuxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

func init() {
	plugin.RegisterDemuxerPlugin(avframe.FormatRtmp, func(ctx context.Context) (plugin.Demuxer, error) {
		return NewRtmpDemuxer(ctx)
	})
}

type RtmpDemuxer struct {
	ctx   context.Context
	frame *avframe.Frame
}

func NewRtmpDemuxer(ctx context.Context) (plugin.Demuxer, error) {
	return &RtmpDemuxer{ctx: ctx}, nil
}

func (d *RtmpDemuxer) Close() error {
	return nil
}

func (d *RtmpDemuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtmp
}

func (d *RtmpDemuxer) Read() (*avframe.Frame, error) {
	fmt.Println("rtmp demuxer read frame")
	return d.frame, nil
}

func (d *RtmpDemuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("rtmp demuxer write frame", frame)
	d.frame = frame
	return nil
}

func (d *RtmpDemuxer) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (d *RtmpDemuxer) Feedback(fb *avframe.Feedback) error {
	fmt.Printf("rtmp demuxer feedback: %+v\n", fb)
	return nil
}
