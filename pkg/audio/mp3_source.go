package audio

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hajimehoshi/go-mp3"
)

// MP3Source reads audio from MP3 files
type MP3Source struct {
	config       StreamConfig
	file         *os.File
	decoder      *mp3.Decoder
	resampler    *SimpleResampler
	running      bool
	mu           sync.Mutex
	buffer       []int16
	sampleRate   int
	channels     int
	needResample bool
}

// NewMP3Source creates a new MP3 audio source
func NewMP3Source(filepath string, config StreamConfig) (*MP3Source, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	srcSampleRate := decoder.SampleRate()
	srcChannels := 2 // MP3 is always stereo in go-mp3

	// Check if we need resampling
	needResample := srcSampleRate != config.SampleRate

	var resampler *SimpleResampler
	if needResample {
		// Create resampler
		resampler = NewSimpleResampler(srcSampleRate, config.SampleRate, srcChannels)
	}

	return &MP3Source{
		config:       config,
		file:         file,
		decoder:      decoder,
		resampler:    resampler,
		running:      false,
		buffer:       make([]int16, 0),
		sampleRate:   srcSampleRate,
		channels:     srcChannels,
		needResample: needResample,
	}, nil
}

// Start starts the MP3 source
func (m *MP3Source) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("MP3 source already running")
	}

	m.running = true
	return nil
}

// Stop stops the MP3 source
func (m *MP3Source) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false

	if m.file != nil {
		m.file.Close()
	}

	return nil
}

// Read reads the next audio frame
func (m *MP3Source) Read() ([]int16, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil, fmt.Errorf("MP3 source not running")
	}

	// Calculate how many samples we need
	// FrameSize is per channel, so for stereo we need FrameSize * 2
	samplesNeeded := m.config.FrameSize
	if m.config.Channels == 2 {
		samplesNeeded *= 2
	}

	readCount := 0
	// Read and decode MP3 data
	for len(m.buffer) < samplesNeeded {
		// Read a chunk of MP3 data
		chunk := make([]byte, 4096)
		n, err := m.decoder.Read(chunk)
		if err == io.EOF {
			// End of file - pad with silence if needed
			if len(m.buffer) > 0 {
				// Pad to frame size
				for len(m.buffer) < samplesNeeded {
					m.buffer = append(m.buffer, 0)
				}
				break
			}
			return nil, io.EOF
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read MP3: %w", err)
		}

		// Debug: Check raw bytes read
		if readCount == 0 {
			fmt.Printf("🔍 Raw bytes read (first 32): %v\n", chunk[:min(32, n)])
		}

		// Convert bytes to int16 samples
		samples := make([]int16, n/2)
		for i := 0; i < n/2; i++ {
			samples[i] = int16(chunk[i*2]) | int16(chunk[i*2+1])<<8
		}

		// Debug: Check sample values
		if readCount == 0 {
			nonZero := 0
			maxAbs := int16(0)
			fmt.Printf("🔍 First 16 samples: %v\n", samples[:min(16, len(samples))])
			for _, s := range samples[:min(100, len(samples))] {
				if s != 0 {
					nonZero++
				}
				if abs(s) > maxAbs {
					maxAbs = abs(s)
				}
			}
			fmt.Printf("🎵 MP3 decode: read %d bytes → %d samples, nonZero=%d/100, maxAbs=%d, needResample=%v\n",
				n, len(samples), nonZero, maxAbs, m.needResample)
			fmt.Printf("🔍 MP3 decoder info: sampleRate=%d, srcChannels=%d, targetRate=%d\n",
				m.sampleRate, m.channels, m.config.SampleRate)
		}
		readCount++

		// Resample if needed
		if m.needResample {
			resampled := m.resampler.Resample(samples)
			m.buffer = append(m.buffer, resampled...)
		} else {
			m.buffer = append(m.buffer, samples...)
		}
	}

	// Extract frame
	frame := make([]int16, samplesNeeded)
	copy(frame, m.buffer[:samplesNeeded])
	m.buffer = m.buffer[samplesNeeded:]

	// Debug: Check final frame values (only first few times)
	if readCount > 0 {
		nonZero := 0
		maxAbs := int16(0)
		for _, s := range frame[:min(100, len(frame))] {
			if s != 0 {
				nonZero++
			}
			if abs(s) > maxAbs {
				maxAbs = abs(s)
			}
		}
		fmt.Printf("🎵 MP3 frame output: %d samples, nonZero=%d/100, maxAbs=%d\n",
			len(frame), nonZero, maxAbs)
	}

	// Convert stereo to mono if needed
	if m.channels == 2 && m.config.Channels == 1 {
		mono := make([]int16, len(frame)/2)
		for i := 0; i < len(mono); i++ {
			// Average left and right channels
			left := int32(frame[i*2])
			right := int32(frame[i*2+1])
			mono[i] = int16((left + right) / 2)
		}
		return mono, nil
	}

	return frame, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

// SampleRate returns the configured output sample rate
func (m *MP3Source) SampleRate() int {
	return m.config.SampleRate
}

// Channels returns the number of output channels
func (m *MP3Source) Channels() int {
	return m.config.Channels
}

// IsRunning returns whether the source is running
func (m *MP3Source) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
