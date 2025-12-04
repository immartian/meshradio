package audio

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// MicrophoneSource reads audio from the microphone
type MicrophoneSource struct {
	config  StreamConfig
	stream  *InputStream
	running bool
	mu      sync.Mutex
}

// NewMicrophoneSource creates a new microphone audio source
func NewMicrophoneSource(config StreamConfig) *MicrophoneSource {
	return &MicrophoneSource{
		config:  config,
		stream:  NewInputStream(config),
		running: false,
	}
}

// Start starts the microphone source
func (m *MicrophoneSource) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("microphone source already running")
	}

	if err := m.stream.Start(); err != nil {
		return err
	}

	m.running = true
	return nil
}

// Stop stops the microphone source
func (m *MicrophoneSource) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.stream.Stop()
	m.running = false
	return nil
}

// Read reads the next audio frame
func (m *MicrophoneSource) Read() ([]int16, error) {
	// Read bytes from stream
	frameBytes, err := m.stream.Read()
	if err != nil {
		return nil, err
	}

	// Convert bytes to int16 samples
	samples := make([]int16, len(frameBytes)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(frameBytes[i*2 : i*2+2]))
	}

	return samples, nil
}

// SampleRate returns the sample rate
func (m *MicrophoneSource) SampleRate() int {
	return m.config.SampleRate
}

// Channels returns the number of channels
func (m *MicrophoneSource) Channels() int {
	return m.config.Channels
}

// IsRunning returns whether the source is running
func (m *MicrophoneSource) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
