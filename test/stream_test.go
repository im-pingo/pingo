package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/stream"
	_ "github.com/pingostack/pingos/plugins"
)

type TestSubscriber struct {
	id     string
	format mode.FmtType
	frame  *mode.Frame
}

func (s *TestSubscriber) ID() string {
	return s.id
}

func (s *TestSubscriber) Format() mode.FmtType {
	return s.format
}

func (s *TestSubscriber) Write(frame *mode.Frame) error {
	fmt.Printf("subscriber write frame: %+v\n", frame)
	s.frame = frame
	return nil
}

func (s *TestSubscriber) Read() *mode.Frame {
	return s.frame
}

func (s *TestSubscriber) Close() error {
	return nil
}

func (s *TestSubscriber) AudioCodecSupported() []mode.CodecType {
	return []mode.CodecType{mode.CodecTypeAAC}
}

func (s *TestSubscriber) VideoCodecSupported() []mode.CodecType {
	return []mode.CodecType{mode.CodecTypeH264}
}

func TestStream_Transcode(t *testing.T) {
	s := stream.NewStream(context.Background(), "live/test")
	s.Activate(&stream.Metadata{
		AudioCodecType: mode.CodecTypeAAC,
		VideoCodecType: mode.CodecTypeH265,
		FmtType:        mode.FormatRtmp,
	})

	sub := &TestSubscriber{
		id:     "test",
		format: mode.FormatRtpRtcp,
	}
	err := s.Subscribe(sub)
	if err != nil {
		t.Fatal(err)
	}

	s.Write(&mode.Frame{
		Fmt:    mode.FormatRtmp,
		Codec:  mode.CodecTypeH265,
		Ts:     uint64(time.Now().UnixNano()),
		Length: uint32(len([]byte("test"))),
		Data:   []byte("test"),
	})

	frame := sub.Read()
	fmt.Printf("subscriber	read frame: %+v\n", frame)

	if frame.TTL != 4 {
		t.Fatalf("frame ttl expect 4, but got %d", frame.TTL)
	}
}
