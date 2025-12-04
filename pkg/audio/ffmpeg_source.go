package audio

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// FFmpegSource reads audio from MP3 files using ffmpeg
type FFmpegSource struct {
	config           StreamConfig
	filepath         string
	cmd              *exec.Cmd
	stdout           io.ReadCloser
	running          bool
	mu               sync.Mutex
	consecutiveSilent int
}

// NewFFmpegSource creates a new FFmpeg-based audio source
func NewFFmpegSource(filepath string, config StreamConfig) (*FFmpegSource, error) {
	fmt.Printf("📁 Opening MP3 with ffmpeg: %s\n", filepath)

	return &FFmpegSource{
		config:   config,
		filepath: filepath,
	}, nil
}

// Start starts the FFmpeg source
func (f *FFmpegSource) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return fmt.Errorf("FFmpeg source already running")
	}

	// Build ffmpeg command to decode MP3 to raw PCM
	// -i input.mp3: input file
	// -f s16le: output format (signed 16-bit little-endian)
	// -ar 48000: output sample rate
	// -ac 2: output channels (stereo)
	// pipe:1: output to stdout
	args := []string{
		"-i", f.filepath,
		"-f", "s16le",
		"-ar", fmt.Sprintf("%d", f.config.SampleRate),
		"-ac", fmt.Sprintf("%d", f.config.Channels),
		"-loglevel", "quiet",
		"pipe:1",
	}

	f.cmd = exec.Command("ffmpeg", args...)

	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	f.stdout = stdout

	if err := f.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	f.running = true
	fmt.Printf("✅ FFmpeg decoder started: %d Hz, %d channels\n", f.config.SampleRate, f.config.Channels)

	return nil
}

// Stop stops the FFmpeg source
func (f *FFmpegSource) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return nil
	}

	f.running = false

	if f.stdout != nil {
		f.stdout.Close()
	}

	if f.cmd != nil && f.cmd.Process != nil {
		f.cmd.Process.Kill()
		f.cmd.Wait()
	}

	return nil
}

// Read reads the next audio frame
func (f *FFmpegSource) Read() ([]int16, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return nil, fmt.Errorf("FFmpeg source not running")
	}

	// Calculate bytes needed for one frame
	// FrameSize is samples per channel, multiply by channels and 2 bytes per sample
	samplesNeeded := f.config.FrameSize
	if f.config.Channels == 2 {
		samplesNeeded *= 2
	}
	bytesNeeded := samplesNeeded * 2 // 2 bytes per int16 sample

	// Read PCM data from ffmpeg
	pcmBytes := make([]byte, bytesNeeded)
	n, err := io.ReadFull(f.stdout, pcmBytes)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		// End of file
		f.running = false
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read from ffmpeg: %w", err)
	}

	// Convert bytes to int16 samples
	samples := make([]int16, n/2)
	for i := 0; i < n/2; i++ {
		samples[i] = int16(pcmBytes[i*2]) | int16(pcmBytes[i*2+1])<<8
	}

	// Detect silence to auto-stop
	nonZero := 0
	maxAbs := int16(0)
	for _, s := range samples {
		if s != 0 {
			nonZero++
		}
		if abs(s) > maxAbs {
			maxAbs = abs(s)
		}
	}

	if nonZero == 0 && maxAbs == 0 {
		f.consecutiveSilent++
		if f.consecutiveSilent >= 50 {
			fmt.Printf("⚠️  Detected %d consecutive silent frames, ending stream\n", f.consecutiveSilent)
			f.running = false
			return nil, io.EOF
		}
	} else {
		f.consecutiveSilent = 0
	}

	return samples, nil
}

// SampleRate returns the configured output sample rate
func (f *FFmpegSource) SampleRate() int {
	return f.config.SampleRate
}

// Channels returns the number of output channels
func (f *FFmpegSource) Channels() int {
	return f.config.Channels
}

// IsRunning returns whether the source is running
func (f *FFmpegSource) IsRunning() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.running
}
