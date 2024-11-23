package stream

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
	eventemitter "github.com/pingostack/pingos/pkg/eventemmiter"
	"github.com/pkg/errors"
)

const (
	// EventStreamActive eventemitter.EventID = iota
	// EventStreamInactive
	EventSubStreamEmpty eventemitter.EventID = iota
)

type Stream struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc

	noopProcessor *avframe.Transmit
	demuxer       *avframe.Transmit
	audioDecoders map[avframe.CodecType]*avframe.Transmit
	videoDecoders map[avframe.CodecType]*avframe.Transmit
	audioEncoders map[avframe.CodecType]*avframe.Transmit
	videoEncoders map[avframe.CodecType]*avframe.Transmit
	muxers        map[avframe.FmtType]*avframe.Transmit
	resourceLock  sync.RWMutex

	publisher     peer.Publisher
	publisherLock sync.RWMutex
	subStreams    map[avframe.FmtType]*subStream

	waitingSubs     []peer.Subscriber
	waitingSubsLock sync.RWMutex
	actived         atomic.Bool

	eventHandle eventemitter.EventEmitter
	onceActive  sync.Once
}

func NewStream(ctx context.Context, id string) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	stream := &Stream{
		id:            id,
		ctx:           ctx,
		cancel:        cancel,
		audioDecoders: make(map[avframe.CodecType]*avframe.Transmit),
		videoDecoders: make(map[avframe.CodecType]*avframe.Transmit),
		audioEncoders: make(map[avframe.CodecType]*avframe.Transmit),
		videoEncoders: make(map[avframe.CodecType]*avframe.Transmit),
		muxers:        make(map[avframe.FmtType]*avframe.Transmit),
		subStreams:    make(map[avframe.FmtType]*subStream),
		waitingSubs:   make([]peer.Subscriber, 0),
		eventHandle:   eventemitter.NewEventEmitter(ctx, 1024),
		onceActive:    sync.Once{},
	}

	stream.eventHandle.On(EventSubStreamEmpty, func(data interface{}) (interface{}, error) {
		ss := data.(*subStream)
		go func() {
			time.Sleep(10 * time.Second)
			if ss.CloseIfEmpty() {
				stream.deleteSubStream(ss.Format())
			}
		}()
		return nil, nil
	})

	return stream
}

func (s *Stream) deleteSubStream(fmtType avframe.FmtType) {
	s.resourceLock.Lock()
	defer s.resourceLock.Unlock()
	delete(s.subStreams, fmtType)
}

func (s *Stream) GetID() string {
	return s.id
}

func (s *Stream) getOrCreateMuxer(fmtType avframe.FmtType) (*avframe.Transmit, bool) {
	if muxer, found := s.muxers[fmtType]; found {
		return muxer, false
	}
	muxer, err := plugin.CreateMuxerPlugin(s.ctx, fmtType)
	if err != nil {
		return nil, false
	}
	transmit := avframe.NewTransmit(muxer)
	return transmit, true
}

func (s *Stream) getOrCreateDemuxer() (*avframe.Transmit, error) {
	if s.demuxer != nil {
		return s.demuxer, nil
	}
	metadata := s.publisher.Metadata()
	demuxer, err := plugin.CreateDemuxerPlugin(s.ctx, metadata.FmtType)
	if err != nil {
		panic(errors.Wrap(err, "create demuxer"))
	}
	transmit := avframe.NewTransmit(demuxer)
	return transmit, nil
}

func (s *Stream) getOrCreateDecoder(codecType avframe.CodecType) (*avframe.Transmit, bool) {
	if codecType.IsAudio() {
		if decoder, found := s.audioDecoders[codecType]; found {
			return decoder, false
		}
		decoder, err := plugin.CreateDecoderPlugin(s.ctx, codecType)
		if err != nil {
			panic(errors.Wrap(err, "create decoder"))
		}
		transmit := avframe.NewTransmit(decoder)
		return transmit, true
	}

	if decoder, found := s.videoDecoders[codecType]; found {
		return decoder, false
	}

	decoder, err := plugin.CreateDecoderPlugin(s.ctx, codecType)
	if err != nil {
		panic(errors.Wrap(err, "create decoder"))
	}
	transmit := avframe.NewTransmit(decoder)
	return transmit, true
}

func (s *Stream) getOrCreateEncoder(codecType avframe.CodecType) (*avframe.Transmit, bool) {
	if codecType.IsAudio() {
		if encoder, found := s.audioEncoders[codecType]; found {
			return encoder, false
		}

		encoder, err := plugin.CreateEncoderPlugin(s.ctx, codecType)
		if err != nil {
			panic(errors.Wrap(err, "create encoder"))
		}
		transmit := avframe.NewTransmit(encoder)
		return transmit, true
	}

	if encoder, found := s.videoEncoders[codecType]; found {
		return encoder, false
	}

	encoder, err := plugin.CreateEncoderPlugin(s.ctx, codecType)
	if err != nil {
		panic(errors.Wrap(err, "create encoder"))
	}
	transmit := avframe.NewTransmit(encoder)
	return transmit, true
}

func (s *Stream) createSubStream(subFmtType avframe.FmtType, fmtSupported avframe.FmtSupported) (ss *subStream, err error) {
	var muxerTransmit *avframe.Transmit
	var demuxerTransmit *avframe.Transmit
	var audioDecoderTransmit *avframe.Transmit
	var videoDecoderTransmit *avframe.Transmit
	var audioEncoderTransmit *avframe.Transmit
	var videoEncoderTransmit *avframe.Transmit

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		if err != nil {
			if muxerTransmit != nil {
				muxerTransmit.Close()
			}
			if demuxerTransmit != nil {
				demuxerTransmit.Close()
			}
			if audioDecoderTransmit != nil {
				audioDecoderTransmit.Close()
			}
			if videoDecoderTransmit != nil {
				videoDecoderTransmit.Close()
			}
			if audioEncoderTransmit != nil {
				audioEncoderTransmit.Close()
			}
			if videoEncoderTransmit != nil {
				videoEncoderTransmit.Close()
			}
		}
	}()

	if s.noopProcessor == nil {
		return nil, fmt.Errorf("stream not activated")
	}

	create := func(processor *avframe.Transmit) (ss *subStream, err error) {
		ss, err = newSubStream(s.ctx, processor, WithOnSubscriberEmpty(func(sub peer.Subscriber) {
			s.eventHandle.EmitEvent(s.ctx, EventSubStreamEmpty, sub)
		}))
		return
	}

	metadata := s.publisher.Metadata()
	if metadata.FmtType == subFmtType {
		return create(s.noopProcessor)
	}

	created := false
	if muxerTransmit, created = s.getOrCreateMuxer(subFmtType); muxerTransmit == nil {
		return nil, fmt.Errorf("muxer[%s] not found", subFmtType)
	}
	if !created {
		return create(muxerTransmit)
	}

	demuxerTransmit, _ = s.getOrCreateDemuxer()
	if demuxerTransmit == nil {
		return nil, fmt.Errorf("demuxer not found")
	}

	// get suitable codec type for target fmt type and add decoder
	var targetAudioCodecType avframe.CodecType
	var targetVideoCodecType avframe.CodecType
	if fmtSupported != nil {
		targetAudioCodecType = fmtSupported.GetSuitableAudioCodecType(subFmtType, metadata.AudioCodecType)
		targetVideoCodecType = fmtSupported.GetSuitableVideoCodecType(subFmtType, metadata.VideoCodecType)
	} else {
		targetAudioCodecType = avframe.GetSuitableAudioCodecType(subFmtType, metadata.AudioCodecType)
		targetVideoCodecType = avframe.GetSuitableVideoCodecType(subFmtType, metadata.VideoCodecType)
	}

	if targetAudioCodecType == avframe.CodecTypeUnknown || targetVideoCodecType == avframe.CodecTypeUnknown {
		return nil, fmt.Errorf("no suitable codec type for target fmt type[%s]", subFmtType)
	}

	if targetAudioCodecType != metadata.AudioCodecType {
		audioDecoderTransmit, _ = s.getOrCreateDecoder(metadata.AudioCodecType)
		if audioDecoderTransmit == nil {
			return nil, fmt.Errorf("audio decoder[%s] not found", metadata.AudioCodecType)
		}
		audioEncoderTransmit, _ = s.getOrCreateEncoder(targetAudioCodecType)
		if audioEncoderTransmit == nil {
			return nil, fmt.Errorf("audio encoder[%s] not found", targetAudioCodecType)
		}
	}

	if targetVideoCodecType != metadata.VideoCodecType {
		videoDecoderTransmit, _ = s.getOrCreateDecoder(metadata.VideoCodecType)
		if videoDecoderTransmit == nil {
			return nil, fmt.Errorf("video decoder[%s] not found", metadata.VideoCodecType)
		}
		videoEncoderTransmit, _ = s.getOrCreateEncoder(targetVideoCodecType)
		if videoEncoderTransmit == nil {
			return nil, fmt.Errorf("video encoder[%s] not found", targetVideoCodecType)
		}
	}

	s.muxers[subFmtType] = muxerTransmit

	if s.demuxer == nil {
		s.demuxer = demuxerTransmit
		s.noopProcessor.AddNextAudioTransmit(demuxerTransmit)
		s.noopProcessor.AddNextVideoTransmit(demuxerTransmit)
	}

	if audioDecoderTransmit != nil {
		if audioDecoderTransmit.PrevAudioTransmit() == nil {
			demuxerTransmit.AddNextAudioTransmit(audioDecoderTransmit)
			s.audioDecoders[targetAudioCodecType] = audioDecoderTransmit
		}
		if audioEncoderTransmit.PrevAudioTransmit() == nil {
			audioDecoderTransmit.AddNextAudioTransmit(audioEncoderTransmit)
			s.audioEncoders[targetAudioCodecType] = audioEncoderTransmit
		}
		audioEncoderTransmit.AddNextAudioTransmit(muxerTransmit)
	} else {
		demuxerTransmit.AddNextAudioTransmit(muxerTransmit)
	}

	if videoDecoderTransmit != nil {
		if videoDecoderTransmit.PrevVideoTransmit() == nil {
			demuxerTransmit.AddNextVideoTransmit(videoDecoderTransmit)
			s.videoDecoders[targetVideoCodecType] = videoDecoderTransmit
		}
		if videoEncoderTransmit.PrevVideoTransmit() == nil {
			videoDecoderTransmit.AddNextVideoTransmit(videoEncoderTransmit)
			s.videoEncoders[targetVideoCodecType] = videoEncoderTransmit
		}
		videoEncoderTransmit.AddNextVideoTransmit(muxerTransmit)
	} else {
		demuxerTransmit.AddNextVideoTransmit(muxerTransmit)
	}

	return create(muxerTransmit)
}

func (s *Stream) getOrCreateSubStream(fmtType avframe.FmtType, fmtSupported avframe.FmtSupported) (*subStream, error) {
	s.resourceLock.Lock()
	defer s.resourceLock.Unlock()
	if _, ok := s.subStreams[fmtType]; !ok {
		subStream, err := s.createSubStream(fmtType, fmtSupported)
		if err != nil {
			return nil, errors.Wrap(err, "new substream")
		}

		s.subStreams[fmtType] = subStream
	}

	return s.subStreams[fmtType], nil
}

func (s *Stream) GetSubStream(fmtType avframe.FmtType) *subStream {
	s.resourceLock.RLock()
	defer s.resourceLock.RUnlock()
	return s.subStreams[fmtType]
}

func (s *Stream) GetSubStreams() []*subStream {
	s.resourceLock.RLock()
	defer s.resourceLock.RUnlock()
	subStreams := make([]*subStream, 0, len(s.subStreams))
	for _, subStream := range s.subStreams {
		subStreams = append(subStreams, subStream)
	}
	return subStreams
}

func (s *Stream) Subscribe(sub peer.Subscriber) error {
	s.waitingSubsLock.Lock()
	if !s.actived.Load() {
		s.waitingSubs = append(s.waitingSubs, sub)
		s.waitingSubsLock.Unlock()
		return nil
	}
	s.waitingSubsLock.Unlock()

	fmtType := sub.Format()
	subVideoCodecSupported := sub.VideoCodecSupported()
	subAudioCodecSupported := sub.AudioCodecSupported()
	fmtSupported := avframe.DupFmtSupported()

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

func (s *Stream) Publish(publisher peer.Publisher) {
	s.publisherLock.Lock()
	defer s.publisherLock.Unlock()
	s.publisher = publisher
	if s.noopProcessor == nil {
		s.noopProcessor = avframe.NewTransmit(avframe.NewNoopProcessor(publisher.Format()))
	}

	s.start()
}

func (s *Stream) Unsubscribe(sub peer.Subscriber) {
	fmtType := sub.Format()
	subStream := s.GetSubStream(fmtType)
	if subStream == nil {
		return
	}
	subStream.Unsubscribe(sub)
}

func (s *Stream) active() {
	s.waitingSubsLock.Lock()
	subs := []peer.Subscriber{}
	s.actived.Store(true)
	subs = append(subs, s.waitingSubs...)
	s.waitingSubs = []peer.Subscriber{}
	s.waitingSubsLock.Unlock()

	for _, sub := range subs {
		s.Subscribe(sub)
	}
}

func (s *Stream) inactive() {
	s.actived.Store(false)
}

func (s *Stream) start() {
	fmt.Println("stream start")
	go func() {
		s.active()
		defer s.inactive()
		for {
			s.publisherLock.RLock()
			publisher := s.publisher
			s.publisherLock.RUnlock()
			if publisher == nil {
				return
			}

			frame, err := publisher.Read()
			if err != nil {
				return
			}

			fmt.Println("publisher read frame", frame)
			s.noopProcessor.Write(frame)

			time.Sleep(1 * time.Second)
		}
	}()
}
