package avframe

import "encoding/json"

type CodecTypeSupported struct {
	VideoCodecTypes []CodecType
	AudioCodecTypes []CodecType
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
	CodecTypeAudioSample
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
	case CodecTypeAudioSample:
		return "audioSample"
	case CodecTypeOPUS:
		return "opus"
	case CodecTypeAAC:
		return "aac"
	case CodecTypeG711:
		return "g711"
	}
	return "unknown"
}

func CodecTypeFromString(s string) CodecType {
	switch s {
	case "yuv":
		return CodecTypeYUV
	case "h264":
		return CodecTypeH264
	case "h265":
		return CodecTypeH265
	case "vp8":
		return CodecTypeVP8
	case "vp9":
		return CodecTypeVP9
	case "opus":
		return CodecTypeOPUS
	case "aac":
		return CodecTypeAAC
	case "g711":
		return CodecTypeG711
	case "audioSample":
		return CodecTypeAudioSample
	}
	return CodecTypeUnknown
}

func (c CodecType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *CodecType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*c = CodecTypeFromString(s)
	return nil
}
