package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pingostack/pingos/core/mode"
	"github.com/pingostack/pingos/core/stream"
	_ "github.com/pingostack/pingos/plugins"
)

type TestSubscriber struct {
	id     string
	format mode.FmtType
}

func (s *TestSubscriber) ID() string {
	return s.id
}

func (s *TestSubscriber) Format() mode.FmtType {
	return s.format
}

func (s *TestSubscriber) Write(frame *mode.Frame) error {
	fmt.Printf("write frame: %+v\n", frame)
	return nil
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

func main() {
	s := stream.NewStream(context.Background(), "live/test")
	s.Activate(&stream.Metadata{
		AudioCodecType: mode.CodecTypeAAC,
		VideoCodecType: mode.CodecTypeH265,
		FmtType:        mode.FormatRtmp,
	})

	err := s.Subscribe(&TestSubscriber{
		id:     "test",
		format: mode.FormatRtpRtcp,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.Subscribe(&TestSubscriber{
		id:     "test2",
		format: mode.FormatRtpRtcp,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	s.Write(&mode.Frame{
		Fmt:    mode.FormatRtmp,
		Codec:  mode.CodecTypeH265,
		Ts:     uint64(time.Now().UnixNano()),
		Length: uint32(len([]byte("test"))),
		Data:   []byte("test"),
	})
}
