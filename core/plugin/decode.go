package plugin

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pingostack/pingos/core/mode"
)

type Decoder interface {
	io.Closer
	mode.Processor
	GetCodecType() mode.CodecType
}

type DecoderPlugin struct {
	CreateDecoder func(ctx context.Context) (Decoder, error)
	CodecType     mode.CodecType
}

var decoderTypes = map[mode.CodecType]*DecoderPlugin{}
var decoderLock sync.Mutex

func RegisterDecoder(plugin *DecoderPlugin) {
	decoderLock.Lock()
	defer decoderLock.Unlock()

	decoderTypes[plugin.CodecType] = plugin
}

func CreateDecoder(ctx context.Context, codecType mode.CodecType) (decoder Decoder, err error) {
	plugin, ok := decoderTypes[codecType]
	if !ok {
		return nil, fmt.Errorf("decoder %s not found", codecType)
	}
	return plugin.CreateDecoder(ctx)
}
