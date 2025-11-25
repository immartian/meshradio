package protocol

import (
	"encoding/binary"
)

// Audio codec types
const (
	CodecOpus uint8 = 0x01
	CodecFLAC uint8 = 0x02
	CodecAAC  uint8 = 0x03
	CodecMP3  uint8 = 0x04
)

// AudioPacket represents an audio stream packet payload
type AudioPacket struct {
	CodecType      uint8
	SampleRate     uint8  // Encoded value (48kHz = 48, 44.1kHz = 44, etc.)
	Channels       uint8
	Bitrate        uint8  // In kbps (64 = 64kbps)
	FrameTimestamp uint32
	AudioData      []byte
}

// MarshalAudioPayload encodes audio packet to bytes
func MarshalAudioPayload(ap *AudioPacket) []byte {
	payloadSize := 8 + len(ap.AudioData)
	buf := make([]byte, payloadSize)

	buf[0] = ap.CodecType
	buf[1] = ap.SampleRate
	buf[2] = ap.Channels
	buf[3] = ap.Bitrate
	binary.BigEndian.PutUint32(buf[4:8], ap.FrameTimestamp)
	copy(buf[8:], ap.AudioData)

	return buf
}

// UnmarshalAudioPayload decodes audio packet from bytes
func UnmarshalAudioPayload(data []byte) (*AudioPacket, error) {
	if len(data) < 8 {
		return nil, ErrInvalidPayload
	}

	ap := &AudioPacket{
		CodecType:      data[0],
		SampleRate:     data[1],
		Channels:       data[2],
		Bitrate:        data[3],
		FrameTimestamp: binary.BigEndian.Uint32(data[4:8]),
	}

	if len(data) > 8 {
		ap.AudioData = make([]byte, len(data)-8)
		copy(ap.AudioData, data[8:])
	}

	return ap, nil
}
