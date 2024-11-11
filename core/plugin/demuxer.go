package plugin

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pingostack/pingos/core/mode"
)

type Demuxer interface {
	io.Closer
	mode.Processor
}

type DemuxerPlugin struct {
	CreateDemuxer func(ctx context.Context) (Demuxer, error)
	FmtType       mode.FmtType
}

var demuxerTypes = map[mode.FmtType]*DemuxerPlugin{}
var demuxerLock sync.Mutex

func RegisterDemuxer(plugin *DemuxerPlugin) {
	demuxerLock.Lock()
	defer demuxerLock.Unlock()

	demuxerTypes[plugin.FmtType] = plugin
}

func CreateDemuxer(ctx context.Context, fmtType mode.FmtType) (demuxer Demuxer, err error) {
	plugin, ok := demuxerTypes[fmtType]
	if !ok {
		return nil, fmt.Errorf("demuxer %s not found", fmtType)
	}
	return plugin.CreateDemuxer(ctx)
}
