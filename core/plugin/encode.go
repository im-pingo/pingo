package plugin

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pingostack/pingos/core/mode"
)

type Encoder interface {
	io.Closer
	mode.Processor
	GetCodecType() mode.CodecType
}

type EncoderPlugin struct {
	CreateEncoder func(ctx context.Context) (Encoder, error)
	CodecType     mode.CodecType
}

var encoderTypes = map[mode.CodecType]*EncoderPlugin{}
var encoderLock sync.Mutex

func RegisterEncoder(plugin *EncoderPlugin) {
	encoderLock.Lock()
	defer encoderLock.Unlock()

	encoderTypes[plugin.CodecType] = plugin
}

func CreateEncoder(ctx context.Context, codecType mode.CodecType) (encoder Encoder, err error) {
	plugin, ok := encoderTypes[codecType]
	if !ok {
		return nil, fmt.Errorf("encoder %s not found", codecType)
	}
	return plugin.CreateEncoder(ctx)
}
