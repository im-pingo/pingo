package plugin

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/pingostack/pingos/pkg/avframe"
)

type Plugin[T any, K comparable] struct {
	Create func(ctx context.Context) (T, error)
	Type   K
}

// demuxer plugin
type Demuxer interface {
	avframe.Processor
	GetFormatType() avframe.FmtType
}

var demuxerPlugins = map[avframe.FmtType]*Plugin[Demuxer, avframe.FmtType]{}
var demuxerPluginLock sync.Mutex

func RegisterDemuxerPlugin(typ avframe.FmtType, create func(ctx context.Context) (Demuxer, error)) error {
	plugin := &Plugin[Demuxer, avframe.FmtType]{
		Create: create,
		Type:   typ,
	}

	demuxerPluginLock.Lock()
	defer demuxerPluginLock.Unlock()
	if _, ok := demuxerPlugins[plugin.Type]; ok {
		return fmt.Errorf("demuxer %s already registered", plugin.Type)
	}

	demuxerPlugins[plugin.Type] = plugin
	return nil
}

func CreateDemuxerPlugin(ctx context.Context, fmtType avframe.FmtType) (demuxer Demuxer, err error) {
	demuxerPluginLock.Lock()
	defer demuxerPluginLock.Unlock()

	plugin, ok := demuxerPlugins[fmtType]
	if !ok {
		return nil, fmt.Errorf("demuxer %s not found", fmtType)
	}
	return plugin.Create(ctx)
}

// muxer plugin
type Muxer interface {
	avframe.Processor
	GetFormatType() avframe.FmtType
}

var muxerPlugins = map[avframe.FmtType]*Plugin[Muxer, avframe.FmtType]{}
var muxerPluginLock sync.Mutex

func RegisterMuxerPlugin(typ avframe.FmtType, create func(ctx context.Context) (Muxer, error)) error {
	plugin := &Plugin[Muxer, avframe.FmtType]{
		Create: create,
		Type:   typ,
	}

	muxerPluginLock.Lock()
	defer muxerPluginLock.Unlock()
	if _, ok := muxerPlugins[plugin.Type]; ok {
		return fmt.Errorf("muxer %s already registered", plugin.Type)
	}

	muxerPlugins[plugin.Type] = plugin
	return nil
}

func CreateMuxerPlugin(ctx context.Context, fmtType avframe.FmtType) (muxer Muxer, err error) {
	muxerPluginLock.Lock()
	defer muxerPluginLock.Unlock()

	plugin, ok := muxerPlugins[fmtType]
	if !ok {
		return nil, fmt.Errorf("muxer %s not found", fmtType)
	}
	return plugin.Create(ctx)
}

// decoder plugin
type Decoder interface {
	avframe.Processor
	GetCodecType() avframe.CodecType
}

var decoderPlugins = map[avframe.CodecType]*Plugin[Decoder, avframe.CodecType]{}
var decoderPluginLock sync.Mutex

func RegisterDecoderPlugin(typ avframe.CodecType, create func(ctx context.Context) (Decoder, error)) error {
	plugin := &Plugin[Decoder, avframe.CodecType]{
		Create: create,
		Type:   typ,
	}

	decoderPluginLock.Lock()
	defer decoderPluginLock.Unlock()
	if _, ok := decoderPlugins[plugin.Type]; ok {
		return fmt.Errorf("decoder %s already registered", plugin.Type)
	}

	decoderPlugins[plugin.Type] = plugin
	return nil
}

func CreateDecoderPlugin(ctx context.Context, codecType avframe.CodecType) (decoder Decoder, err error) {
	decoderPluginLock.Lock()
	defer decoderPluginLock.Unlock()

	plugin, ok := decoderPlugins[codecType]
	if !ok {
		return nil, fmt.Errorf("decoder %s not found", codecType)
	}
	return plugin.Create(ctx)
}

// encoder plugin
type Encoder interface {
	avframe.Processor
	GetCodecType() avframe.CodecType
}

var encoderPlugins = map[avframe.CodecType]*Plugin[Encoder, avframe.CodecType]{}
var encoderPluginLock sync.Mutex

func RegisterEncoderPlugin(typ avframe.CodecType, create func(ctx context.Context) (Encoder, error)) error {
	plugin := &Plugin[Encoder, avframe.CodecType]{
		Create: create,
		Type:   typ,
	}

	encoderPluginLock.Lock()
	defer encoderPluginLock.Unlock()
	if _, ok := encoderPlugins[plugin.Type]; ok {
		return fmt.Errorf("encoder %s already registered", plugin.Type)
	}

	encoderPlugins[plugin.Type] = plugin
	return nil
}

func CreateEncoderPlugin(ctx context.Context, codecType avframe.CodecType) (encoder Encoder, err error) {
	encoderPluginLock.Lock()
	defer encoderPluginLock.Unlock()

	plugin, ok := encoderPlugins[codecType]
	if !ok {
		return nil, fmt.Errorf("encoder %s not found", codecType)
	}
	return plugin.Create(ctx)
}

type Interceptor interface {
	avframe.Processor
	Priority() int
}

var interceptorPlugins = map[avframe.FmtType][]*Plugin[Interceptor, avframe.FmtType]{}
var interceptorPluginLock sync.Mutex

func RegisterInterceptorPlugin(typ avframe.FmtType, create func(ctx context.Context) (Interceptor, error)) error {
	plugin := &Plugin[Interceptor, avframe.FmtType]{
		Create: create,
		Type:   typ,
	}

	interceptorPluginLock.Lock()
	defer interceptorPluginLock.Unlock()
	if _, ok := interceptorPlugins[plugin.Type]; ok {
		return fmt.Errorf("interceptor %s already registered", plugin.Type)
	}

	interceptorPlugins[plugin.Type] = append(interceptorPlugins[plugin.Type], plugin)
	return nil
}

func CreateInterceptorPlugins(ctx context.Context, typ avframe.FmtType) (interceptors []Interceptor, err error) {
	interceptorPluginLock.Lock()
	defer interceptorPluginLock.Unlock()

	plugins, ok := interceptorPlugins[typ]
	if !ok {
		return nil, fmt.Errorf("interceptor %s not found", typ)
	}
	for _, plugin := range plugins {
		interceptor, err := plugin.Create(ctx)
		if err != nil {
			return nil, err
		}
		interceptors = append(interceptors, interceptor)
	}

	// sort by priority, higher priority first
	sort.Slice(interceptors, func(i, j int) bool {
		return interceptors[i].Priority() > interceptors[j].Priority()
	})

	return
}
