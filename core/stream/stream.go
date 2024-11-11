package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/plugin"
	eventemitter "github.com/pingostack/pingos/pkg/eventemmiter"
	"github.com/pkg/errors"
)

const (
	EventSubstreamEmpty eventemitter.EventID = iota
	EventStreamActive
)

type Stream struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc

	demuxer       plugin.Demuxer
	audioDecoders map[mode.CodecType]plugin.Decoder
	videoDecoders map[mode.CodecType]plugin.Decoder
	audioEncoders map[mode.CodecType]plugin.Encoder
	videoEncoders map[mode.CodecType]plugin.Encoder
	muxers        map[mode.FmtType]plugin.Muxer
	fmtLock       sync.RWMutex

	subStreams    map[mode.FmtType]*subStream
	subStreamLock sync.RWMutex

	waitingSubs     []peer.Subscriber
	waitingSubsLock sync.RWMutex
	actived         bool

	eventHandle eventemitter.EventEmitter
	onceActive  sync.Once

	metadata *Metadata
}

func NewStream(ctx context.Context, id string) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	stream := &Stream{
		id:            id,
		ctx:           ctx,
		cancel:        cancel,
		audioDecoders: make(map[mode.CodecType]plugin.Decoder),
		videoDecoders: make(map[mode.CodecType]plugin.Decoder),
		audioEncoders: make(map[mode.CodecType]plugin.Encoder),
		videoEncoders: make(map[mode.CodecType]plugin.Encoder),
		muxers:        make(map[mode.FmtType]plugin.Muxer),
		subStreams:    make(map[mode.FmtType]*subStream),
		waitingSubs:   make([]peer.Subscriber, 0),
		eventHandle:   eventemitter.NewEventEmitter(ctx, 1024),
		onceActive:    sync.Once{},
	}

	stream.eventHandle.On(EventSubstreamEmpty, func(data interface{}) (interface{}, error) {
		subStream := data.(*subStream)
		subStream.Close()
		stream.deleteSubStream(subStream.GetFormatType())
		return nil, nil
	})

	stream.eventHandle.On(EventStreamActive, func(data interface{}) (interface{}, error) {
		stream.onceActive.Do(func() {
			waitingSubs := make([]peer.Subscriber, 0)
			stream.waitingSubsLock.Lock()
			stream.actived = true
			waitingSubs = append(waitingSubs, stream.waitingSubs...)
			stream.waitingSubs = nil
			stream.waitingSubsLock.Unlock()

			for _, sub := range waitingSubs {
				stream.Subscribe(sub)
			}
		})
		return nil, nil
	})

	return stream
}

func (s *Stream) deleteSubStream(fmtType mode.FmtType) {
	s.subStreamLock.Lock()
	defer s.subStreamLock.Unlock()
	delete(s.subStreams, fmtType)
}

func (s *Stream) GetID() string {
	return s.id
}

func (s *Stream) getOrCreateMuxer(fmtType mode.FmtType) (plugin.Muxer, bool) {
	s.fmtLock.Lock()
	defer s.fmtLock.Unlock()
	if muxer, found := s.muxers[fmtType]; found {
		return muxer, false
	}
	muxer, err := plugin.CreateMuxer(s.ctx, fmtType)
	if err != nil {
		return nil, false
	}
	s.muxers[fmtType] = muxer
	return muxer, true
}

func (s *Stream) getOrCreateDemuxer() (plugin.Demuxer, bool) {
	s.fmtLock.Lock()
	defer s.fmtLock.Unlock()
	if s.demuxer != nil {
		return s.demuxer, true
	}
	demuxer, err := plugin.CreateDemuxer(s.ctx, s.metadata.FmtType)
	if err != nil {
		return nil, false
	}
	s.demuxer = demuxer
	return demuxer, true
}

func (s *Stream) getOrCreateDecoder(codecType mode.CodecType) (plugin.Decoder, bool) {
	s.fmtLock.Lock()
	defer s.fmtLock.Unlock()

	if codecType.IsAudio() {
		if decoder, found := s.audioDecoders[codecType]; found {
			return decoder, true
		}
		decoder, err := plugin.CreateDecoder(s.ctx, codecType)
		if err != nil {
			return nil, false
		}
		s.audioDecoders[codecType] = decoder
		return decoder, true
	}

	if decoder, found := s.videoDecoders[codecType]; found {
		return decoder, true
	}

	decoder, err := plugin.CreateDecoder(s.ctx, codecType)
	if err != nil {
		return nil, false
	}
	s.videoDecoders[codecType] = decoder
	return decoder, true
}

func (s *Stream) getOrCreateEncoder(codecType mode.CodecType) (plugin.Encoder, bool) {
	s.fmtLock.Lock()
	defer s.fmtLock.Unlock()

	if codecType.IsAudio() {
		if encoder, found := s.audioEncoders[codecType]; found {
			return encoder, true
		}

		encoder, err := plugin.CreateEncoder(s.ctx, codecType)
		if err != nil {
			return nil, false
		}
		s.audioEncoders[codecType] = encoder
		return encoder, true
	}

	if encoder, found := s.videoEncoders[codecType]; found {
		return encoder, true
	}

	encoder, err := plugin.CreateEncoder(s.ctx, codecType)
	if err != nil {
		return nil, false
	}
	s.videoEncoders[codecType] = encoder
	return encoder, true
}

func (s *Stream) generateFormatSettings(targetFmtType mode.FmtType, fmtSupported mode.FmtSupported) (*FormatSettings, error) {
	settings := &FormatSettings{}
	if targetFmtType == s.metadata.FmtType {
		return settings, nil
	}

	var muxer plugin.Muxer
	var created bool
	if muxer, created = s.getOrCreateMuxer(targetFmtType); !created {
		settings.Muxer = muxer
		return settings, nil
	}

	if muxer == nil {
		return nil, errors.New("muxer not found")
	}

	// add demuxer
	demuxer, _ := s.getOrCreateDemuxer()
	if demuxer == nil {
		return nil, errors.New("demuxer not found")
	}
	settings.Demuxer = demuxer
	settings.Muxer = muxer

	// get suitable codec type for target fmt type and add decoder
	var targetAudioCodecType mode.CodecType
	var targetVideoCodecType mode.CodecType
	if fmtSupported != nil {
		targetAudioCodecType = fmtSupported.GetSuitableAudioCodecType(targetFmtType, s.metadata.AudioCodecType)
		targetVideoCodecType = fmtSupported.GetSuitableVideoCodecType(targetFmtType, s.metadata.VideoCodecType)
	} else {
		targetAudioCodecType = mode.GetSuitableAudioCodecType(targetFmtType, s.metadata.AudioCodecType)
		targetVideoCodecType = mode.GetSuitableVideoCodecType(targetFmtType, s.metadata.VideoCodecType)
	}

	if targetAudioCodecType == mode.CodecTypeUnknown || targetVideoCodecType == mode.CodecTypeUnknown {
		return nil, fmt.Errorf("no suitable codec type for target fmt type[%s]", targetFmtType)
	}

	if targetAudioCodecType != s.metadata.AudioCodecType {
		settings.AudioDecoder, _ = s.getOrCreateDecoder(s.metadata.AudioCodecType)
		if settings.AudioDecoder == nil {
			return nil, fmt.Errorf("audio decoder[%s] not found", s.metadata.AudioCodecType)
		}
		settings.AudioEncoder, _ = s.getOrCreateEncoder(targetAudioCodecType)
		if settings.AudioEncoder == nil {
			return nil, fmt.Errorf("audio encoder[%s] not found", targetAudioCodecType)
		}
	}

	if targetVideoCodecType != s.metadata.VideoCodecType {
		settings.VideoDecoder, _ = s.getOrCreateDecoder(s.metadata.VideoCodecType)
		if settings.VideoDecoder == nil {
			return nil, fmt.Errorf("video decoder[%s] not found", s.metadata.VideoCodecType)
		}
		settings.VideoEncoder, _ = s.getOrCreateEncoder(targetVideoCodecType)
		if settings.VideoEncoder == nil {
			return nil, fmt.Errorf("video encoder[%s] not found", targetVideoCodecType)
		}
	}

	return settings, nil
}

func (s *Stream) newFormat(fmtType mode.FmtType, fmtSupported mode.FmtSupported) (*Format, error) {
	if s.metadata == nil {
		return nil, errors.New("metadata is nil")
	}
	settings, err := s.generateFormatSettings(fmtType, fmtSupported)
	if err != nil {
		return nil, errors.Wrap(err, "generate format settings")
	}
	return NewFormat(s.ctx, fmtType, settings)
}

func (s *Stream) getOrCreateSubStream(fmtType mode.FmtType, fmtSupported mode.FmtSupported) (*subStream, error) {
	s.subStreamLock.Lock()
	defer s.subStreamLock.Unlock()
	if _, ok := s.subStreams[fmtType]; !ok {
		f, err := s.newFormat(fmtType, fmtSupported)
		if err != nil {
			return nil, errors.Wrap(err, "new format")
		}
		subStream, err := newSubStream(s.ctx, s.eventHandle, f)
		if err != nil {
			return nil, errors.Wrap(err, "new substream")
		}

		s.subStreams[fmtType] = subStream
	}

	return s.subStreams[fmtType], nil
}

func (s *Stream) GetSubStream(fmtType mode.FmtType) *subStream {
	s.subStreamLock.RLock()
	defer s.subStreamLock.RUnlock()
	return s.subStreams[fmtType]
}

func (s *Stream) GetSubStreams() []*subStream {
	s.subStreamLock.RLock()
	defer s.subStreamLock.RUnlock()
	subStreams := make([]*subStream, 0, len(s.subStreams))
	for _, subStream := range s.subStreams {
		subStreams = append(subStreams, subStream)
	}
	return subStreams
}

func (s *Stream) Subscribe(sub peer.Subscriber) error {
	func() {
		s.waitingSubsLock.Lock()
		defer s.waitingSubsLock.Unlock()
		if !s.actived {
			s.waitingSubs = append(s.waitingSubs, sub)
			return
		}
	}()

	fmtType := sub.Format()
	subVideoCodecSupported := sub.VideoCodecSupported()
	subAudioCodecSupported := sub.AudioCodecSupported()
	fmtSupported := mode.DupFmtSupported()

	// Retrieve the current format settings
	currentFmt := fmtSupported[fmtType]

	if subVideoCodecSupported != nil {
		currentFmt.VideoCodecTypes = subVideoCodecSupported
	}
	if subAudioCodecSupported != nil {
		currentFmt.AudioCodecTypes = subAudioCodecSupported
	}

	// Put the modified struct back into the map
	fmtSupported[fmtType] = currentFmt

	subStream, err := s.getOrCreateSubStream(fmtType, fmtSupported)
	if err != nil {
		return errors.Wrap(err, "get or create substream")
	}
	subStream.Subscribe(sub)
	return nil
}

func (s *Stream) Activate(metadata *Metadata) {
	s.metadata = metadata
	result, err := s.eventHandle.EmitEventWithResult(s.ctx, EventStreamActive, nil)
	if err != nil || result.Error != nil {
		return
	}
}

func (s *Stream) Unsubscribe(sub peer.Subscriber) {
	fmtType := sub.Format()
	subStream := s.GetSubStream(fmtType)
	if subStream == nil {
		return
	}
	subStream.Unsubscribe(sub)
}

func (s *Stream) Write(f *mode.Frame) error {
	subStreams := s.GetSubStreams()
	for _, subStream := range subStreams {
		subStream.Write(f)
	}
	return nil
}
