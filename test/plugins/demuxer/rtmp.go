package demuxer

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	logger.Info("Demuxer initialized")
	plugin.RegisterDemuxerPlugin(avframe.FormatRtmp, func(ctx context.Context, metadata avframe.Metadata) (plugin.Demuxer, error) {
		return NewRtmpDemuxer(ctx, metadata)
	})
}

type RtmpDemuxer struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func NewRtmpDemuxer(ctx context.Context, metadata avframe.Metadata) (plugin.Demuxer, error) {
	return &RtmpDemuxer{ctx: ctx, inMetadata: metadata}, nil
}

func (d *RtmpDemuxer) Close() error {
	logger.Info("rtmp demuxer close")
	return nil
}

func (d *RtmpDemuxer) GetFormatType() avframe.FmtType {
	return avframe.FormatRtmp
}

func (d *RtmpDemuxer) Read() (*avframe.Frame, error) {
	logger.Info("rtmp demuxer read frame")
	return d.frame, nil
}

func (d *RtmpDemuxer) Write(frame *avframe.Frame) error {
	frame.TTL++
	logger.Info("rtmp demuxer write frame", frame)
	d.frame = frame
	return nil
}

func (d *RtmpDemuxer) Format() avframe.FmtType {
	return avframe.FormatRaw
}

func (d *RtmpDemuxer) Feedback(fb *avframe.Feedback) error {
	logger.Infof("rtmp demuxer feedback: %+v\n", fb)
	return nil
}

func (d *RtmpDemuxer) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        avframe.FormatRaw,
		AudioCodecType: d.inMetadata.AudioCodecType,
		VideoCodecType: d.inMetadata.VideoCodecType,
	}
}

func (d *RtmpDemuxer) UpdateSourceMetadata(metadata avframe.Metadata) {
	d.inMetadata = metadata
}
