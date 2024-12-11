package avframe

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type AudioHeader struct {
	Codec CodecType
	Rate  uint16
	Bits  uint8
}

func (h AudioHeader) Len() int {
	return 4
}

func (h AudioHeader) String() string {
	return fmt.Sprintf("AudioHeader{Codec: %s, Rate: %d, Bits: %d}", h.Codec, h.Rate, h.Bits)
}

type VideoHeader struct {
	Codec       CodecType
	Orientation uint8
}

func (h VideoHeader) Len() int {
	return 2
}

func (h VideoHeader) String() string {
	return fmt.Sprintf("VideoHeader{Codec: %s, Orientation: %d}", h.Codec, h.Orientation)
}

type Frame struct {
	Fmt         FmtType
	PayloadType PayloadType
	Ts          uint64
	Length      uint32
	TTL         uint8
	Attributes  map[string]string
	Data        []byte
}

type PayloadType uint8

const (
	PayloadTypeUnknown PayloadType = iota
	PayloadTypeMetadata
	PayloadTypeAudio
	PayloadTypeVideo
	PayloadTypeData
)

func (p PayloadType) String() string {
	switch p {
	case PayloadTypeAudio:
		return "audio"
	case PayloadTypeVideo:
		return "video"
	case PayloadTypeData:
		return "data"
	case PayloadTypeMetadata:
		return "metadata"
	}
	return "unknown"
}

func PayloadTypeFromString(s string) PayloadType {
	switch s {
	case "audio":
		return PayloadTypeAudio
	case "video":
		return PayloadTypeVideo
	case "data":
		return PayloadTypeData
	case "metadata":
		return PayloadTypeMetadata
	}
	return PayloadTypeUnknown
}

func (p *PayloadType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = PayloadTypeFromString(s)
	return nil
}

func (p PayloadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (f *Frame) IsVideo() bool {
	return f.PayloadType == PayloadTypeVideo
}

func (f *Frame) IsAudio() bool {
	return f.PayloadType == PayloadTypeAudio
}

func (f *Frame) IsMetadata() bool {
	return f.PayloadType == PayloadTypeMetadata
}

func (f *Frame) Reset() {
	f.Fmt = FormatRaw
	f.PayloadType = PayloadTypeUnknown
	f.Length = 0
	f.Data = nil
}

func (f *Frame) Dup() *Frame {
	return &Frame{
		Fmt:         f.Fmt,
		PayloadType: f.PayloadType,
		Ts:          f.Ts,
		Length:      f.Length,
		Data:        append([]byte(nil), f.Data...),
	}
}

func (f *Frame) CopyFrom(src *Frame) {
	f.Fmt = src.Fmt
	f.PayloadType = src.PayloadType
	f.Ts = src.Ts
	f.Length = src.Length
	f.Data = append([]byte(nil), src.Data...)
}

func (f *Frame) CopyTo(dst []byte) []byte {
	dst = append(dst, byte(f.Fmt))
	dst = append(dst, byte(f.PayloadType))
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
	payloadType := p[1]
	ts := binary.BigEndian.Uint64(p[2:10])
	length := binary.BigEndian.Uint32(p[10:14])

	return &Frame{
		Fmt:         FmtType(fmtType),
		PayloadType: PayloadType(payloadType),
		Ts:          ts,
		Length:      length,
		Data:        p[14:],
	}
}

func (f *Frame) GetVideoHeader() *VideoHeader {
	if !f.IsVideo() {
		return nil
	}

	header := &VideoHeader{}
	header.Codec = CodecType(f.Data[0])
	header.Orientation = f.Data[1]
	return header
}

func (f *Frame) GetAudioHeader() *AudioHeader {
	if !f.IsAudio() {
		return nil
	}

	header := &AudioHeader{}
	header.Codec = CodecType(f.Data[0])
	header.Rate = binary.BigEndian.Uint16(f.Data[1:3])
	header.Bits = f.Data[3]
	return header
}

func (f *Frame) WriteAudioHeader(header *AudioHeader) {
	f.Data[0] = byte(header.Codec)
	binary.BigEndian.PutUint16(f.Data[1:3], header.Rate)
	f.Data[3] = header.Bits
}

func (f *Frame) WriteVideoHeader(header *VideoHeader) {
	f.Data[0] = byte(header.Codec)
	f.Data[1] = header.Orientation
}

func (f *Frame) WritePayload(payload []byte) {
	f.Data = append(f.Data, payload...)
}

func (f *Frame) CodecType() CodecType {
	if f.IsAudio() {
		return f.GetAudioHeader().Codec
	}
	return f.GetVideoHeader().Codec
}

func (f *Frame) String() string {
	if f.IsAudio() {
		return fmt.Sprintf("Frame{Fmt: %s, PayloadType: %s, Ts: %d, Length: %d, TTL %d, AudioHeader: %s}", f.Fmt, f.PayloadType, f.Ts, f.Length, f.TTL, f.GetAudioHeader())
	} else if f.IsVideo() {
		return fmt.Sprintf("Frame{Fmt: %s, PayloadType: %s, Ts: %d, Length: %d, TTL %d, VideoHeader: %s}", f.Fmt, f.PayloadType, f.Ts, f.Length, f.TTL, f.GetVideoHeader())
	} else if f.IsMetadata() {
		metadata := Metadata{}
		metadata.Unmarshal(f.Data)
		return fmt.Sprintf("Frame{Fmt: %s, PayloadType: %s, Ts: %d, Length: %d, TTL %d, Metadata: %s}", f.Fmt, f.PayloadType, f.Ts, f.Length, f.TTL, metadata)
	} else {
		return fmt.Sprintf("Frame{Fmt: %s, PayloadType: %s, Ts: %d, Length: %d, TTL %d, Data: %v}", f.Fmt, f.PayloadType, f.Ts, f.Length, f.TTL, f.Data)
	}
}
