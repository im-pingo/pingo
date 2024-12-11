package stream

import (
	"context"
	"io"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
)

type Format struct {
	fmtType       avframe.FmtType
	ctx           context.Context
	cancel        context.CancelFunc
	data          chan *avframe.Frame
	demuxer       *avframe.Pipeline
	audioDecoder  *avframe.Pipeline
	videoDecoder  *avframe.Pipeline
	audioEncoder  *avframe.Pipeline
	videoEncoder  *avframe.Pipeline
	muxer         plugin.Muxer
	audioProcess  *avframe.Pipeline
	videoProcess  *avframe.Pipeline
	audioFeedback *avframe.Pipeline
	videoFeedback *avframe.Pipeline
}

type FormatSettings struct {
	Demuxer      *avframe.Pipeline
	AudioDecoder *avframe.Pipeline
	VideoDecoder *avframe.Pipeline
	AudioEncoder *avframe.Pipeline
	VideoEncoder *avframe.Pipeline
	Muxer        plugin.Muxer
}

func NewFormat(ctx context.Context, fmtType avframe.FmtType, settings *FormatSettings) (*Format, error) {
	ctx, cancel := context.WithCancel(ctx)
	f := &Format{
		fmtType: fmtType,
		ctx:     ctx,
		cancel:  cancel,
		data:    make(chan *avframe.Frame, 1024),
	}

	if settings != nil {
		f.demuxer = settings.Demuxer
		f.audioDecoder = settings.AudioDecoder
		f.videoDecoder = settings.VideoDecoder
		f.audioEncoder = settings.AudioEncoder
		f.videoEncoder = settings.VideoEncoder
		f.muxer = settings.Muxer
	}

	return f, nil
}

func (f *Format) Write(frame *avframe.Frame) error {
	if f.videoProcess != nil && frame.IsVideo() {
		return f.videoProcess.Write(frame)
	}
	if f.audioProcess != nil && frame.IsAudio() {
		return f.audioProcess.Write(frame)
	}
	return nil
}

func (f *Format) Read() (*avframe.Frame, error) {
	frame, ok := <-f.data
	if !ok {
		return nil, io.EOF
	}
	return frame, nil
}

func (f *Format) Close() error {
	f.cancel()
	return nil
}

func (f *Format) Feedback(fb *avframe.Feedback) error {
	logger.Infof("format feedback: %+v", fb)
	if f.videoFeedback != nil {
		f.videoFeedback.Feedback(fb)
	}
	if f.audioFeedback != nil {
		f.audioFeedback.Feedback(fb)
	}
	return nil
}
