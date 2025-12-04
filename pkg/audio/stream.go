package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
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
	ctx      *malgo.AllocatedContext
	device   *malgo.Device
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
	defer out.mu.Unlock()

	if out.running {
		return fmt.Errorf("stream already running")
	}

	// Initialize malgo context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize audio context: %w", err)
	}
	out.ctx = ctx

	// Configure playback device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(out.config.Channels)
	deviceConfig.SampleRate = uint32(out.config.SampleRate)
	deviceConfig.Alsa.NoMMap = 1

	// Data callback - called when device needs audio data
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		select {
		case frame := <-out.frames:
			// Copy frame data to output buffer
			copy(pOutputSample, frame)
		default:
			// No frame available, output silence
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}
		}
	}

	// Initialize device
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: onRecvFrames,
	})
	if err != nil {
		ctx.Uninit()
		return fmt.Errorf("failed to initialize playback device: %w", err)
	}
	out.device = device

	// Start the device
	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		return fmt.Errorf("failed to start playback device: %w", err)
	}

	out.running = true
	fmt.Println("🔊 Audio playback started")
	return nil
}

// Stop stops audio playback
func (out *OutputStream) Stop() {
	out.mu.Lock()
	defer out.mu.Unlock()

	if !out.running {
		return
	}

	// Stop and clean up audio device
	if out.device != nil {
		out.device.Uninit()
		out.device = nil
	}

	if out.ctx != nil {
		out.ctx.Uninit()
		out.ctx = nil
	}

	close(out.stopChan)
	out.running = false
	fmt.Println("🔇 Audio playback stopped")
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

// Note: Playback is now handled by malgo's data callback in Start()
// The callback reads from out.frames channel and copies data to the audio device
