package mode

import (
	"encoding/binary"
	"slices"
)

type FmtType int8

const (
	FormatRaw FmtType = iota
	FormatRtmp
	FormatFlv
	FormatMpegts
	FormatRtpRtcp
)

type CodecTypeSupported struct {
	VideoCodecTypes []CodecType
	AudioCodecTypes []CodecType
}

type FmtSupported map[FmtType]CodecTypeSupported

func (fs FmtSupported) GetSupportedVideoCodecTypes(f FmtType) []CodecType {
	supported, ok := fs[f]
	if !ok {
		return nil
	}
	return supported.VideoCodecTypes
}

func (fs FmtSupported) GetSuitableVideoCodecType(f FmtType, codecType CodecType) CodecType {
	supported := fs.GetSupportedVideoCodecTypes(f)
	if slices.Contains(supported, codecType) {
		return codecType
	}
	if len(supported) > 0 {
		return supported[0]
	}
	return CodecTypeUnknown
}

func (fs FmtSupported) GetSuitableAudioCodecType(f FmtType, codecType CodecType) CodecType {
	supported := fs.GetSupportedAudioCodecTypes(f)
	if slices.Contains(supported, codecType) {
		return codecType
	}
	if len(supported) > 0 {
		return supported[0]
	}
	return CodecTypeUnknown
}

func (fs FmtSupported) GetSupportedAudioCodecTypes(f FmtType) []CodecType {
	supported, ok := fs[f]
	if !ok {
		return nil
	}
	return supported.AudioCodecTypes
}

var fmtSupported = FmtSupported{
	FormatFlv: {
		VideoCodecTypes: []CodecType{
			CodecTypeH264,
			CodecTypeH265,
		},
		AudioCodecTypes: []CodecType{
			CodecTypeAAC,
			CodecTypeMP3,
		},
	},
	FormatRtmp: {
		VideoCodecTypes: []CodecType{
			CodecTypeH264,
			CodecTypeH265,
		},
		AudioCodecTypes: []CodecType{
			CodecTypeAAC,
			CodecTypeMP3,
		},
	},
	FormatMpegts: {
		VideoCodecTypes: []CodecType{
			CodecTypeH264,
			CodecTypeH265,
		},
		AudioCodecTypes: []CodecType{
			CodecTypeAAC,
			CodecTypeMP3,
		},
	},
	FormatRtpRtcp: {
		VideoCodecTypes: []CodecType{
			CodecTypeVP8,
			CodecTypeVP9,
			CodecTypeH264,
			CodecTypeH265,
		},
		AudioCodecTypes: []CodecType{
			CodecTypeAAC,
			CodecTypeMP3,
			CodecTypeOPUS,
			CodecTypeG711,
		},
	},
}

func GetSupportedVideoCodecTypes(f FmtType) []CodecType {
	return fmtSupported.GetSupportedVideoCodecTypes(f)
}

func GetSuitableVideoCodecType(f FmtType, codecType CodecType) CodecType {
	return fmtSupported.GetSuitableVideoCodecType(f, codecType)
}

func GetSupportedAudioCodecTypes(f FmtType) []CodecType {
	return fmtSupported.GetSupportedAudioCodecTypes(f)
}

func GetSuitableAudioCodecType(f FmtType, codecType CodecType) CodecType {
	return fmtSupported.GetSuitableAudioCodecType(f, codecType)
}

func DupFmtSupported() FmtSupported {
	result := make(FmtSupported)
	for k, v := range fmtSupported {
		result[k] = v
	}
	return result
}

func (f FmtType) String() string {
	switch f {
	case FormatRaw:
		return "raw"
	case FormatRtmp:
		return "rtmp"
	case FormatFlv:
		return "flv"
	case FormatMpegts:
		return "mpegts"
	case FormatRtpRtcp:
		return "rtprtcp"
	}
	return "unknown"
}

type CodecType uint8

const (
	CodecTypeUnknown CodecType = iota
	CodecTypeYUV
	CodecTypeH264
	CodecTypeH265
	CodecTypeVP8
	CodecTypeVP9
	CodecTypeVideoCount
	CodecTypeOPUS
	CodecTypeAAC
	CodecTypeMP3
	CodecTypeG711
	CodecTypeAudioCount
)

func (c CodecType) IsVideo() bool {
	return c > CodecTypeUnknown && c < CodecTypeVideoCount
}

func (c CodecType) IsAudio() bool {
	return c > CodecTypeVideoCount && c < CodecTypeAudioCount
}

func (c CodecType) String() string {
	switch c {
	case CodecTypeYUV:
		return "yuv"
	case CodecTypeH264:
		return "h264"
	case CodecTypeH265:
		return "h265"
	case CodecTypeVP8:
		return "vp8"
	case CodecTypeVP9:
		return "vp9"
	case CodecTypeOPUS:
		return "opus"
	case CodecTypeAAC:
		return "aac"
	case CodecTypeG711:
		return "g711"
	}
	return "unknown"
}

type Frame struct {
	Fmt    FmtType
	Codec  CodecType
	Ts     uint64
	Length uint32
	TTL    uint8
	Data   []byte
}

func (f *Frame) IsVideo() bool {
	return f.Codec.IsVideo()
}

func (f *Frame) IsAudio() bool {
	return f.Codec.IsAudio()
}

func (f *Frame) Reset() {
	f.Fmt = FormatRaw
	f.Codec = CodecTypeUnknown
	f.Length = 0
	f.Data = nil
}

func (f *Frame) Dup() *Frame {
	return &Frame{
		Fmt:    f.Fmt,
		Codec:  f.Codec,
		Ts:     f.Ts,
		Length: f.Length,
		Data:   append([]byte(nil), f.Data...),
	}
}

func (f *Frame) CopyFrom(src *Frame) {
	f.Fmt = src.Fmt
	f.Codec = src.Codec
	f.Ts = src.Ts
	f.Length = src.Length
	f.Data = append([]byte(nil), src.Data...)
}

func (f *Frame) CopyTo(dst []byte) []byte {
	dst = append(dst, byte(f.Fmt))
	dst = append(dst, byte(f.Codec))
	binary.BigEndian.PutUint64(dst[2:10], f.Ts)
	binary.BigEndian.PutUint32(dst[10:14], f.Length)
	dst = append(dst, f.Data...)
	return dst
}

func (f *Frame) Len() int {
	return 10 + len(f.Data)
}

func ParseFrame(p []byte) *Frame {
	// parse frame header, fmt[1 byte], codec[1 byte], ts[8 bytes]
	fmtType := p[0]
	codecType := p[1]
	ts := binary.BigEndian.Uint64(p[2:10])
	length := binary.BigEndian.Uint32(p[10:14])

	return &Frame{
		Fmt:    FmtType(fmtType),
		Codec:  CodecType(codecType),
		Ts:     ts,
		Length: length,
		Data:   p[14:],
	}
}
