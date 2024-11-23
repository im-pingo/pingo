package stream

import (
	"context"
	"fmt"
	"io"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

type Format struct {
	fmtType       avframe.FmtType
	ctx           context.Context
	cancel        context.CancelFunc
	data          chan *avframe.Frame
	demuxer       *avframe.Transmit
	audioDecoder  *avframe.Transmit
	videoDecoder  *avframe.Transmit
	audioEncoder  *avframe.Transmit
	videoEncoder  *avframe.Transmit
	muxer         plugin.Muxer
	audioProcess  *avframe.Transmit
	videoProcess  *avframe.Transmit
	audioFeedback *avframe.Transmit
	videoFeedback *avframe.Transmit
}

type FormatSettings struct {
	Demuxer      *avframe.Transmit
	AudioDecoder *avframe.Transmit
	VideoDecoder *avframe.Transmit
	AudioEncoder *avframe.Transmit
	VideoEncoder *avframe.Transmit
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
	fmt.Printf("format feedback: %+v\n", fb)
	if f.videoFeedback != nil {
		f.videoFeedback.Feedback(fb)
	}
	if f.audioFeedback != nil {
		f.audioFeedback.Feedback(fb)
	}
	return nil
}
