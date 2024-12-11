package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pingostack/pingos/core/peer"
	"github.com/pingostack/pingos/core/stream"
	"github.com/pingostack/pingos/pkg/avframe"
	"github.com/pingostack/pingos/pkg/logger"
	_ "github.com/pingostack/pingos/test/plugins"
)

type TestSubscriber struct {
	id        string
	format    avframe.FmtType
	processor avframe.Processor
}

func (s *TestSubscriber) ID() string {
	return s.id
}

func (s *TestSubscriber) Format() avframe.FmtType {
	return s.format
}

func (s *TestSubscriber) Write(frame *avframe.Frame) error {
	writeCount++
	if frame.Fmt != s.format {
		logger.Fatal("frame format not match")
	}

	if frame.CodecType() != s.processor.Metadata().AudioCodecType && frame.CodecType() != s.processor.Metadata().VideoCodecType {
		logger.Fatal("frame codec not match")
	}

	logger.Infof("write frame: %s", frame)
	if frame.IsAudio() {
		if frame.TTL != 5 {
			logger.Fatalf("audio frame ttl(%d) != 5", frame.TTL)
		}
	} else {
		if frame.TTL != 3 {
			logger.Fatal("video frame ttl != 2")
		}
	}
	return nil
}

func (s *TestSubscriber) Close() error {
	return nil
}

func (s *TestSubscriber) AudioCodecSupported() []avframe.CodecType {
	return []avframe.CodecType{avframe.CodecTypeAAC}
}

func (s *TestSubscriber) VideoCodecSupported() []avframe.CodecType {
	return []avframe.CodecType{avframe.CodecTypeH265}
}

func (s *TestSubscriber) SetProcessor(processor avframe.Processor) {
	s.processor = processor
}

func (s *TestSubscriber) String() string {
	return fmt.Sprintf("TestSubscriber{id: %s, format: %s}", s.id, s.format)
}

type TestPublisher struct {
	id     string
	format avframe.FmtType
}

func (p *TestPublisher) ID() string {
	return p.id
}

func (p *TestPublisher) Format() avframe.FmtType {
	return p.format
}

func (p *TestPublisher) Metadata() avframe.Metadata {
	return avframe.Metadata{
		AudioCodecType: avframe.CodecTypeOPUS,
		VideoCodecType: avframe.CodecTypeH265,
		FmtType:        p.format,
	}
}

var frameCount int
var writeCount int
var readCount int

func (p *TestPublisher) Read() (*avframe.Frame, error) {
	readCount++
	time.Sleep(1 * time.Second)
	frameCount++
	if frameCount%2 != 0 {
		frame := &avframe.Frame{
			Fmt:         p.format,
			PayloadType: avframe.PayloadTypeAudio,
			Ts:          uint64(time.Now().UnixNano()),
			Length:      uint32(avframe.AudioHeader{}.Len()),
			Data:        make([]byte, avframe.AudioHeader{}.Len()),
		}
		frame.WriteAudioHeader(&avframe.AudioHeader{
			Codec: avframe.CodecTypeOPUS,
			Rate:  44100,
			Bits:  16,
		})
		return frame, nil
	}
	frame := &avframe.Frame{
		Fmt:         p.format,
		PayloadType: avframe.PayloadTypeVideo,
		Ts:          uint64(time.Now().UnixNano()),
		Length:      uint32(avframe.VideoHeader{}.Len()),
		Data:        make([]byte, avframe.VideoHeader{}.Len()),
	}
	frame.WriteVideoHeader(&avframe.VideoHeader{
		Codec:       avframe.CodecTypeH265,
		Orientation: 1,
	})
	return frame, nil
}

func (p *TestPublisher) Close() error {
	return nil
}

func main() {
	s := stream.NewStream(context.Background(), "live/test")

	sub := &TestSubscriber{
		id:     "test",
		format: avframe.FormatRtpRtcp,
	}
	ch := make(chan struct{})
	processor, err := s.Subscribe(sub, func(sub peer.Subscriber, processor avframe.Processor, err error) {
		if err != nil {
			logger.Info("subscribe error: ", err)
			logger.Fatal("subscribe error", err)
		} else {
			logger.Info("subscribe success, metadata: ", processor.Metadata())
		}
		sub.SetProcessor(processor)
		close(ch)
	})

	if err != nil {
		logger.Info(err)
		return
	}

	s.Publish(&TestPublisher{
		id:     "test",
		format: avframe.FormatRtmp,
	})
	logger.Info("publish success")

	if processor != nil {
		sub.SetProcessor(processor)
		logger.Fatal("processor is not nil")
	} else {
		<-ch
		processor = sub.processor
	}

	logger.Info("processor: ", processor.Metadata())

	logger.Info("start feedback")
	for i := 0; i < 10; i++ {
		if sub.processor != nil {
			logger.Infof("sub metadata: %+v", sub.processor.Metadata())

			sub.processor.Feedback(&avframe.Feedback{
				Type:  stream.FeedbackTypeSubscriberActive,
				Audio: true,
				Video: true,
				Data: &stream.FeedbackSubscriberActive{
					Subscriber: sub,
					Active:     true,
				},
			})
		}
		time.Sleep(1 * time.Second)
	}

	s.Close()

	<-s.Done()

	if writeCount != readCount {
		logger.Fatalf("write count(%d) != read count(%d)", writeCount, readCount)
	}

	logger.Infof("test pass, write count: %d, read count: %d", writeCount, readCount)
}
