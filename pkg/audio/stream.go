package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/meshradio/meshradio/pkg/logging"
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
		SampleRate: 48000,  // 48kHz for music quality
		Channels:   2,      // Stereo for music
		Bitrate:    128000, // 128kbps for good music quality
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
	config         StreamConfig
	running        bool
	mu             sync.Mutex
	frames         chan []byte
	stopChan       chan struct{}
	ctx            *malgo.AllocatedContext
	device         *malgo.Device
	callbackCount  uint64          // Track callback invocations
	framesConsumed uint64          // Track frames consumed from buffer
	logger         *logging.Logger // Logger for playback debugging
}

// NewOutputStream creates a new audio output stream
func NewOutputStream(config StreamConfig) *OutputStream {
	// Create logger for playback
	logger, err := logging.NewLogger(logging.DefaultConfig("playback"))
	if err != nil {
		fmt.Printf("Warning: Failed to create playback logger: %v\n", err)
		// Continue without logger rather than failing
	}

	return &OutputStream{
		config:   config,
		frames:   make(chan []byte, 150), // 150 frames = 3 seconds buffer for network jitter
		stopChan: make(chan struct{}),
		logger:   logger,
	}
}

// Start begins audio playback
func (out *OutputStream) Start() error {
	out.mu.Lock()
	defer out.mu.Unlock()

	if out.running {
		return fmt.Errorf("stream already running")
	}

	// Log playback start
	if out.logger != nil {
		out.logger.Info("Starting audio playback: %d Hz, %d channels, frameSize=%d",
			out.config.SampleRate, out.config.Channels, out.config.FrameSize)
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
	deviceConfig.PeriodSizeInFrames = uint32(out.config.FrameSize)
	deviceConfig.Periods = 4 // 4 periods for smoother playback
	deviceConfig.Alsa.NoMMap = 1

	// Data callback - called when device needs audio data
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		out.callbackCount++

		// Calculate expected bytes (framecount is frames, not samples)
		expectedBytes := int(framecount) * out.config.Channels * 2 // 2 bytes per sample (int16)

		// Use select with timeout instead of immediate default to prevent race conditions
		// while still avoiding deadlock at startup
		select {
		case frame := <-out.frames:
			out.framesConsumed++
			// Ensure we copy the right amount of data
			if len(frame) >= expectedBytes {
				copy(pOutputSample, frame[:expectedBytes])
			} else {
				// Frame too small, copy what we have and pad with silence
				copy(pOutputSample, frame)
				for i := len(frame); i < expectedBytes; i++ {
					pOutputSample[i] = 0
				}
			}

			// Log every 250 callbacks (every ~5 seconds)
			if out.callbackCount%250 == 0 {
				if out.logger != nil {
					out.logger.Debug("Playback: callback=%d, buffer=%d/%d",
						out.callbackCount, len(out.frames), cap(out.frames))
				}
			}
		case <-time.After(5 * time.Millisecond):
			// Timeout: no frame available after 5ms, output silence
			// This is a genuine underrun (not a race condition)
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}

			// Log underrun less frequently to avoid spam
			if out.callbackCount%50 == 0 {
				if out.logger != nil {
					out.logger.Warn("Playback underrun: callback=%d, buffer=%d/%d, timeout waiting for frame",
						out.callbackCount, len(out.frames), cap(out.frames))
				}
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
	fmt.Printf("🔊 Audio playback started (%d Hz, %d channels)\n",
		out.config.SampleRate, out.config.Channels)
	return nil
}

// Stop stops audio playback
func (out *OutputStream) Stop() {
	out.mu.Lock()
	defer out.mu.Unlock()

	if !out.running {
		return
	}

	// Mark as not running first to prevent new operations
	out.running = false

	// Close stop channel to signal goroutines
	close(out.stopChan)

	// Stop and clean up audio device with timeout to prevent hanging
	if out.device != nil || out.ctx != nil {
		done := make(chan struct{})
		go func() {
			defer close(done)
			if out.device != nil {
				out.device.Uninit()
				out.device = nil
			}
			if out.ctx != nil {
				out.ctx.Uninit()
				out.ctx = nil
			}
		}()

		// Wait for cleanup with 1 second timeout
		select {
		case <-done:
			// Cleanup completed
		case <-time.After(1 * time.Second):
			if out.logger != nil {
				out.logger.Warn("Audio device cleanup timed out")
			}
		}
	}

	// Close logger
	if out.logger != nil {
		out.logger.Info("Audio playback stopped")
		out.logger.Close()
		out.logger = nil
	}

	fmt.Println("Audio playback stopped")
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
		if out.logger != nil {
			out.logger.Warn("Audio buffer full, dropping frame (buffer=%d/%d)", len(out.frames), cap(out.frames))
		}
		return nil
	}
}

// Note: Playback is now handled by malgo's data callback in Start()
// The callback reads from out.frames channel and copies data to the audio device
