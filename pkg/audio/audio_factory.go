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
func NewInputStreamAuto(config StreamConfig) (interface{}, error) {
	backend := GetBackend()

	switch backend {
	case BackendPortAudio:
		fmt.Fprintln(os.Stderr, "🎤 Using real microphone (PortAudio)")
		return NewPortAudioInputStream(config)

	case BackendSimulated:
		fmt.Fprintln(os.Stderr, "🔇 Using simulated audio (install PortAudio for real audio)")
		return NewInputStream(config), nil
	}

	return NewInputStream(config), nil
}

// NewOutputStreamAuto creates the best available output stream
func NewOutputStreamAuto(config StreamConfig) (interface{}, error) {
	backend := GetBackend()

	switch backend {
	case BackendPortAudio:
		fmt.Fprintln(os.Stderr, "🔊 Using real speakers (PortAudio)")
		return NewPortAudioOutputStream(config)

	case BackendSimulated:
		fmt.Fprintln(os.Stderr, "🔇 Using simulated audio (install PortAudio for real audio)")
		return NewOutputStream(config), nil
	}

	return NewOutputStream(config), nil
}

// NewCodecAuto creates the best available codec
func NewCodecAuto(sampleRate, channels, frameSize int) (Codec, error) {
	if hasOpus {
		fmt.Fprintln(os.Stderr, "🎵 Using Opus codec")
		return NewOpusCodec(sampleRate, channels, frameSize)
	}

	fmt.Fprintln(os.Stderr, "⚠️  Using dummy codec (install Opus for compression)")
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
