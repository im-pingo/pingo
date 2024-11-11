package demuxer

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

func init() {
	plugin.RegisterDemuxer(&plugin.DemuxerPlugin{
		FmtType: mode.FormatRtmp,
		CreateDemuxer: func(ctx context.Context) (plugin.Demuxer, error) {
			return NewRtmpDemuxer(ctx)
		},
	})
}

type RtmpDemuxer struct {
	ctx context.Context
}

func NewRtmpDemuxer(ctx context.Context) (plugin.Demuxer, error) {
	return &RtmpDemuxer{ctx: ctx}, nil
}

func (d *RtmpDemuxer) Close() error {
	return nil
}

func (d *RtmpDemuxer) Process(frame *mode.Frame) (*mode.Frame, error) {
	frame.TTL++
	fmt.Printf("rtmp demuxer process frame: %+v\n", frame)
	return frame, nil
}

func (d *RtmpDemuxer) Feedback(fb *mode.Feedback) error {
	fmt.Printf("rtmp demuxer feedback: %+v\n", fb)
	return nil
}
