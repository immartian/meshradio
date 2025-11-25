package audio

import (
	"fmt"
	"sync"
	"time"
)

// StreamConfig holds audio stream configuration
type StreamConfig struct {
	SampleRate int
	Channels   int
	Bitrate    int
	FrameSize  int // in samples
}

// DefaultConfig returns sensible defaults for voice
func DefaultConfig() StreamConfig {
	return StreamConfig{
		SampleRate: 48000,  // 48kHz
		Channels:   1,      // Mono for voice
		Bitrate:    64000,  // 64kbps
		FrameSize:  960,    // 20ms at 48kHz
	}
}

// InputStream captures audio from microphone
type InputStream struct {
	config   StreamConfig
	running  bool
	mu       sync.Mutex
	frames   chan []byte
	stopChan chan struct{}
}

// NewInputStream creates a new audio input stream
func NewInputStream(config StreamConfig) *InputStream {
	return &InputStream{
		config:   config,
		frames:   make(chan []byte, 10), // Buffer 10 frames
		stopChan: make(chan struct{}),
	}
}

// Start begins capturing audio
func (in *InputStream) Start() error {
	in.mu.Lock()
	if in.running {
		in.mu.Unlock()
		return fmt.Errorf("stream already running")
	}
	in.running = true
	in.mu.Unlock()

	// For MVP, we'll simulate audio capture
	// In production, this would use PortAudio
	go in.captureLoop()

	return nil
}

// Stop stops capturing audio
func (in *InputStream) Stop() {
	in.mu.Lock()
	defer in.mu.Unlock()

	if !in.running {
		return
	}

	close(in.stopChan)
	in.running = false
}

// Read returns the next audio frame
func (in *InputStream) Read() ([]byte, error) {
	select {
	case frame := <-in.frames:
		return frame, nil
	case <-in.stopChan:
		return nil, fmt.Errorf("stream stopped")
	}
}

// captureLoop simulates audio capture
// TODO: Replace with real PortAudio capture
func (in *InputStream) captureLoop() {
	frameDuration := time.Duration(in.config.FrameSize*1000/in.config.SampleRate) * time.Millisecond
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Generate silence for MVP (2 bytes per sample, mono)
			frameSize := in.config.FrameSize * 2 * in.config.Channels
			frame := make([]byte, frameSize)

			// Try to send, drop if buffer full
			select {
			case in.frames <- frame:
			default:
				// Buffer full, drop frame
			}
		case <-in.stopChan:
			return
		}
	}
}

// OutputStream plays audio to speaker
type OutputStream struct {
	config   StreamConfig
	running  bool
	mu       sync.Mutex
	frames   chan []byte
	stopChan chan struct{}
}

// NewOutputStream creates a new audio output stream
func NewOutputStream(config StreamConfig) *OutputStream {
	return &OutputStream{
		config:   config,
		frames:   make(chan []byte, 10),
		stopChan: make(chan struct{}),
	}
}

// Start begins audio playback
func (out *OutputStream) Start() error {
	out.mu.Lock()
	if out.running {
		out.mu.Unlock()
		return fmt.Errorf("stream already running")
	}
	out.running = true
	out.mu.Unlock()

	// For MVP, we'll simulate playback
	// In production, this would use PortAudio
	go out.playbackLoop()

	return nil
}

// Stop stops audio playback
func (out *OutputStream) Stop() {
	out.mu.Lock()
	defer out.mu.Unlock()

	if !out.running {
		return
	}

	close(out.stopChan)
	out.running = false
}

// Write queues an audio frame for playback
func (out *OutputStream) Write(frame []byte) error {
	select {
	case out.frames <- frame:
		return nil
	case <-out.stopChan:
		return fmt.Errorf("stream stopped")
	default:
		// Buffer full, drop frame
		return nil
	}
}

// playbackLoop simulates audio playback
// TODO: Replace with real PortAudio playback
func (out *OutputStream) playbackLoop() {
	frameDuration := time.Duration(out.config.FrameSize*1000/out.config.SampleRate) * time.Millisecond

	for {
		select {
		case frame := <-out.frames:
			// In production, this would write to PortAudio
			_ = frame
			time.Sleep(frameDuration)
		case <-out.stopChan:
			return
		}
	}
}
