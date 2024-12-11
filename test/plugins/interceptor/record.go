package interceptor

import (
	"context"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

func init() {
	plugin.RegisterInterceptorPlugin(avframe.FormatRtpRtcp, func(ctx context.Context, metadata avframe.Metadata) (plugin.Interceptor, error) {
		return &Record{
			ctx:        ctx,
			inMetadata: metadata,
		}, nil
	})
}

type Record struct {
	ctx        context.Context
	inMetadata avframe.Metadata
	frame      *avframe.Frame
}

func (r *Record) Priority() int {
	return 0
}

func (r *Record) Close() error {
	logger.Info("record interceptor close")
	return nil
}

func (r *Record) Feedback(fb *avframe.Feedback) error {
	logger.Info("record interceptor feedback", fb)
	return nil
}

func (r *Record) Format() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func (r *Record) Read() (*avframe.Frame, error) {
	logger.Info("record interceptor read frame", r.frame)
	return r.frame, nil
}

func (r *Record) Write(frame *avframe.Frame) error {
	frame.TTL++
	logger.Info("record interceptor write frame", frame)
	r.frame = frame
	return nil
}

func (r *Record) Metadata() avframe.Metadata {
	return avframe.Metadata{
		FmtType:        r.inMetadata.FmtType,
		AudioCodecType: r.inMetadata.AudioCodecType,
		VideoCodecType: r.inMetadata.VideoCodecType,
	}
}

func (r *Record) UpdateSourceMetadata(metadata avframe.Metadata) {
	r.inMetadata = metadata
}
