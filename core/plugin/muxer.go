package plugin

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pingostack/pingos/core/mode"
)

type Muxer interface {
	io.Closer
	mode.Processor
	GetFormatType() mode.FmtType
}

type MuxerPlugin struct {
	CreateMuxer func(ctx context.Context) (Muxer, error)
	FmtType     mode.FmtType
}

var muxerTypes = map[mode.FmtType]*MuxerPlugin{}
var muxerLock sync.Mutex

func RegisterMuxer(plugin *MuxerPlugin) {
	muxerLock.Lock()
	defer muxerLock.Unlock()

	muxerTypes[plugin.FmtType] = plugin
}

func CreateMuxer(ctx context.Context, fmtType mode.FmtType) (muxer Muxer, err error) {
	plugin, ok := muxerTypes[fmtType]
	if !ok {
		return nil, fmt.Errorf("muxer %s not found", fmtType)
	}
	return plugin.CreateMuxer(ctx)
}
