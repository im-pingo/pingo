package avframe

import (
	"encoding/json"
	"fmt"
)

type Metadata struct {
	AudioCodecType CodecType `json:"audio_codec_type"`
	VideoCodecType CodecType `json:"video_codec_type"`
	FmtType        FmtType   `json:"fmt_type"`
}

func (m *Metadata) Marshal() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m *Metadata) Unmarshal(data []byte) {
	json.Unmarshal(data, m)
}

func (m *Metadata) String() string {
	return fmt.Sprintf("Metadata{AudioCodecType: %s, VideoCodecType: %s, FmtType: %s}", m.AudioCodecType, m.VideoCodecType, m.FmtType)
}
