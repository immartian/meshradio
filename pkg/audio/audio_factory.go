package audio

import (
	"fmt"
	"os"
)

// AudioBackend indicates which audio backend is available
type AudioBackend int

const (
	BackendSimulated AudioBackend = iota
	BackendPortAudio
)

// GetBackend returns the available audio backend
func GetBackend() AudioBackend {
	// Check if PortAudio is available (will be set by build tags)
	if hasPortAudio {
		return BackendPortAudio
	}
	return BackendSimulated
}

// NewInputStreamAuto creates the best available input stream
func NewInputStreamAuto(config StreamConfig) (*InputStream, error) {
	// Always use simulated for now (PortAudio requires build tags)
	fmt.Fprintln(os.Stderr, "🔇 Using simulated audio (rebuild with -tags portaudio for real audio)")
	return NewInputStream(config), nil
}

// NewOutputStreamAuto creates the best available output stream
func NewOutputStreamAuto(config StreamConfig) (*OutputStream, error) {
	// Always use simulated for now (PortAudio requires build tags)
	fmt.Fprintln(os.Stderr, "🔇 Using simulated audio (rebuild with -tags portaudio for real audio)")
	return NewOutputStream(config), nil
}

// NewCodecAuto creates the best available codec
func NewCodecAuto(sampleRate, channels, frameSize int) (Codec, error) {
	// Always use dummy codec for now (Opus requires build tags)
	fmt.Fprintln(os.Stderr, "⚠️  Using dummy codec (rebuild with -tags opus for compression)")
	return NewDummyCodec(frameSize), nil
}

// GetAudioInfo returns information about audio capabilities
func GetAudioInfo() string {
	backend := GetBackend()

	info := "Audio Backend: "
	if backend == BackendPortAudio {
		info += "PortAudio (Real Audio) ✅\n"
	} else {
		info += "Simulated ⚠️\n"
	}

	info += "Codec: "
	if hasOpus {
		info += "Opus ✅\n"
	} else {
		info += "None (Pass-through) ⚠️\n"
	}

	return info
}
