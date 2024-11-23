package interceptor

import (
	"context"
	"fmt"

	"github.com/pingostack/pingos/core/plugin"
	"github.com/pingostack/pingos/pkg/avframe"
)

type Record struct {
	frame *avframe.Frame
}

func (r *Record) Priority() int {
	return 0
}

func (r *Record) Create(ctx context.Context) (plugin.Interceptor, error) {
	return r, nil
}

func (r *Record) Close() error {
	return nil
}

func (r *Record) Feedback(fb *avframe.Feedback) error {
	fmt.Println("record interceptor feedback", fb)
	return nil
}

func (r *Record) Format() avframe.FmtType {
	return avframe.FormatRtpRtcp
}

func (r *Record) Read() (*avframe.Frame, error) {
	fmt.Println("record interceptor read frame", r.frame)
	return r.frame, nil
}

func (r *Record) Write(frame *avframe.Frame) error {
	frame.TTL++
	fmt.Println("record interceptor write frame", frame)
	r.frame = frame
	return nil
}

func init() {
	plugin.RegisterInterceptorPlugin(avframe.FormatRtpRtcp, func(ctx context.Context) (plugin.Interceptor, error) {
		return &Record{}, nil
	})
}
