package stream

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/plugin"
)

type Format struct {
	fmtType       mode.FmtType
	ctx           context.Context
	cancel        context.CancelFunc
	data          chan *mode.Frame
	demuxer       plugin.Demuxer
	audioDecoder  plugin.Decoder
	videoDecoder  plugin.Decoder
	audioEncoder  plugin.Encoder
	videoEncoder  plugin.Encoder
	muxer         plugin.Muxer
	audioProcess  *mode.Transmit
	videoProcess  *mode.Transmit
	audioFeedback *mode.Transmit
	videoFeedback *mode.Transmit
}

type FormatSettings struct {
	Demuxer      plugin.Demuxer
	AudioDecoder plugin.Decoder
	VideoDecoder plugin.Decoder
	AudioEncoder plugin.Encoder
	VideoEncoder plugin.Encoder
	Muxer        plugin.Muxer
}

func NewFormat(ctx context.Context, fmtType mode.FmtType, settings *FormatSettings) (*Format, error) {
	ctx, cancel := context.WithCancel(ctx)
	f := &Format{
		fmtType: fmtType,
		ctx:     ctx,
		cancel:  cancel,
		data:    make(chan *mode.Frame, 1024),
	}

	if settings != nil {
		f.demuxer = settings.Demuxer
		f.audioDecoder = settings.AudioDecoder
		f.videoDecoder = settings.VideoDecoder
		f.audioEncoder = settings.AudioEncoder
		f.videoEncoder = settings.VideoEncoder
		f.muxer = settings.Muxer
	}

	var videoProcessors, audioProcessors []mode.Processor
	if f.demuxer != nil {
		videoProcessors = append(videoProcessors, f.demuxer)
		audioProcessors = append(audioProcessors, f.demuxer)
	}
	if f.audioDecoder != nil {
		audioProcessors = append(audioProcessors, f.audioDecoder)
	}
	if f.videoDecoder != nil {
		videoProcessors = append(videoProcessors, f.videoDecoder)
	}
	if f.audioEncoder != nil {
		audioProcessors = append(audioProcessors, f.audioEncoder)
	}
	if f.videoEncoder != nil {
		videoProcessors = append(videoProcessors, f.videoEncoder)
	}
	if f.muxer != nil {
		videoProcessors = append(videoProcessors, f.muxer)
		audioProcessors = append(audioProcessors, f.muxer)
	}

	f.videoProcess, f.videoFeedback = mode.CreateTransmit(videoProcessors...)
	f.audioProcess, f.audioFeedback = mode.CreateTransmit(audioProcessors...)

	return f, nil
}

func (f *Format) Process(frame *mode.Frame) (*mode.Frame, error) {
	if f.videoProcess == nil && f.audioProcess == nil {
		return frame, nil
	}

	if f.videoProcess != nil && frame.IsVideo() {
		return f.videoProcess.Process(frame)
	}
	if f.audioProcess != nil && frame.IsAudio() {
		return f.audioProcess.Process(frame)
	}
	return frame, nil
}

func (f *Format) Close() error {
	f.cancel()
	return nil
}

func (f *Format) Feedback(fb *mode.Feedback) error {
	fmt.Printf("format feedback: %+v\n", fb)
	if f.videoFeedback != nil {
		f.videoFeedback.Feedback(fb)
	}
	if f.audioFeedback != nil {
		f.audioFeedback.Feedback(fb)
	}
	return nil
}
