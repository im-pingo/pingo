package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pingostack/pingos/core/stream"
	"github.com/pingostack/pingos/pkg/avframe"
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
	fmt.Printf("write frame: %+v\n", frame)
	if frame.IsAudio() {
		fmt.Printf("audio frame: %+v\n", frame)
		if frame.TTL != 5 {
			panic("audio frame ttl != 4")
		}
	} else {
		fmt.Printf("video frame: %+v\n", frame)
		if frame.TTL != 3 {
			panic("video frame ttl != 2")
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

func (p *TestPublisher) Metadata() *avframe.Metadata {
	return &avframe.Metadata{
		AudioCodecType: avframe.CodecTypeOPUS,
		VideoCodecType: avframe.CodecTypeH265,
		FmtType:        p.format,
	}
}

var frameCount int

func (p *TestPublisher) Read() (*avframe.Frame, error) {
	frameCount++
	if frameCount%2 != 0 {
		return &avframe.Frame{
			Fmt:    p.format,
			Codec:  avframe.CodecTypeOPUS,
			Ts:     uint64(time.Now().UnixNano()),
			Length: uint32(len([]byte("test"))),
			Data:   []byte("test"),
		}, nil
	}
	return &avframe.Frame{
		Fmt:    p.format,
		Codec:  avframe.CodecTypeH265,
		Ts:     uint64(time.Now().UnixNano()),
		Length: uint32(len([]byte("test"))),
		Data:   []byte("test"),
	}, nil
}

func (p *TestPublisher) Close() error {
	return nil
}

func main() {
	s := stream.NewStream(context.Background(), "live/test")
	s.Publish(&TestPublisher{
		id:     "test",
		format: avframe.FormatRtmp,
	})

	sub := &TestSubscriber{
		id:     "test",
		format: avframe.FormatRtpRtcp,
	}
	err := s.Subscribe(sub)

	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(1 * time.Second)
	for {
		sub.processor.Feedback(&avframe.Feedback{
			ID:    sub.ID(),
			Audio: true,
			Video: true,
		})
		return
	}

}
