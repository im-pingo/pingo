package avframe

import (
	"encoding/json"
	"slices"
)

type FmtType int8

const (
	FormatUnknown FmtType = iota
	FormatRaw
	FormatAudioSample
	FormatVideoSample
	FormatRtmp
	FormatFlv
	FormatMpegts
	FormatRtpRtcp
)

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
	case FormatAudioSample:
		return "audiosample"
	case FormatVideoSample:
		return "videosample"
	}
	return "unknown"
}

func FmtTypeFromString(s string) FmtType {
	switch s {
	case "raw":
		return FormatRaw
	case "rtmp":
		return FormatRtmp
	case "flv":
		return FormatFlv
	case "mpegts":
		return FormatMpegts
	case "rtprtcp":
		return FormatRtpRtcp
	case "audiosample":
		return FormatAudioSample
	case "videosample":
		return FormatVideoSample
	}
	return FormatUnknown
}

func (f FmtType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *FmtType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = FmtTypeFromString(s)
	return nil
}
