// +build portaudio

package audio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
)

// PortAudioInputStream captures audio using PortAudio
type PortAudioInputStream struct {
	config StreamConfig
	stream *portaudio.Stream
	buffer []int16
	frames chan []byte
}

// NewPortAudioInputStream creates a real audio input stream
func NewPortAudioInputStream(config StreamConfig) (*PortAudioInputStream, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	bufferSize := config.FrameSize * config.Channels

	in := &PortAudioInputStream{
		config: config,
		buffer: make([]int16, bufferSize),
		frames: make(chan []byte, 10),
	}

	// Open input stream
	stream, err := portaudio.OpenDefaultStream(
		config.Channels, // input channels
		0,               // output channels
		float64(config.SampleRate),
		config.FrameSize,
		in.buffer,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	in.stream = stream
	return in, nil
}

// Start begins capturing audio
func (in *PortAudioInputStream) Start() error {
	if err := in.stream.Start(); err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	go in.captureLoop()
	return nil
}

// Stop stops capturing audio
func (in *PortAudioInputStream) Stop() {
	if in.stream != nil {
		in.stream.Stop()
		in.stream.Close()
	}
	portaudio.Terminate()
}

// Read returns the next audio frame
func (in *PortAudioInputStream) Read() ([]byte, error) {
	frame, ok := <-in.frames
	if !ok {
		return nil, fmt.Errorf("stream closed")
	}
	return frame, nil
}

// captureLoop continuously captures audio from microphone
func (in *PortAudioInputStream) captureLoop() {
	for {
		// Read from PortAudio
		err := in.stream.Read()
		if err != nil {
			continue
		}

		// Convert int16 to bytes
		frameBytes := make([]byte, len(in.buffer)*2)
		for i, sample := range in.buffer {
			frameBytes[i*2] = byte(sample)
			frameBytes[i*2+1] = byte(sample >> 8)
		}

		// Send to channel
		select {
		case in.frames <- frameBytes:
		default:
			// Buffer full, drop frame
		}
	}
}

// PortAudioOutputStream plays audio using PortAudio
type PortAudioOutputStream struct {
	config StreamConfig
	stream *portaudio.Stream
	buffer []int16
	frames chan []byte
}

// NewPortAudioOutputStream creates a real audio output stream
func NewPortAudioOutputStream(config StreamConfig) (*PortAudioOutputStream, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	bufferSize := config.FrameSize * config.Channels

	out := &PortAudioOutputStream{
		config: config,
		buffer: make([]int16, bufferSize),
		frames: make(chan []byte, 10),
	}

	// Open output stream
	stream, err := portaudio.OpenDefaultStream(
		0,               // input channels
		config.Channels, // output channels
		float64(config.SampleRate),
		config.FrameSize,
		out.buffer,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	out.stream = stream
	return out, nil
}

// Start begins audio playback
func (out *PortAudioOutputStream) Start() error {
	if err := out.stream.Start(); err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	go out.playbackLoop()
	return nil
}

// Stop stops audio playback
func (out *PortAudioOutputStream) Stop() {
	if out.stream != nil {
		out.stream.Stop()
		out.stream.Close()
	}
	portaudio.Terminate()
}

// Write queues an audio frame for playback
func (out *PortAudioOutputStream) Write(frame []byte) error {
	select {
	case out.frames <- frame:
		return nil
	default:
		// Buffer full, drop frame
		return nil
	}
}

// playbackLoop continuously plays audio to speakers
func (out *PortAudioOutputStream) playbackLoop() {
	for {
		frame, ok := <-out.frames
		if !ok {
			return
		}

		// Convert bytes to int16
		for i := 0; i < len(out.buffer) && i*2+1 < len(frame); i++ {
			out.buffer[i] = int16(frame[i*2]) | int16(frame[i*2+1])<<8
		}

		// Write to PortAudio
		if err := out.stream.Write(); err != nil {
			continue
		}
	}
}

// ListInputDevices returns available input devices
func ListInputDevices() ([]string, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	var inputs []string
	for _, dev := range devices {
		if dev.MaxInputChannels > 0 {
			inputs = append(inputs, dev.Name)
		}
	}

	return inputs, nil
}

// ListOutputDevices returns available output devices
func ListOutputDevices() ([]string, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	var outputs []string
	for _, dev := range devices {
		if dev.MaxOutputChannels > 0 {
			outputs = append(outputs, dev.Name)
		}
	}

	return outputs, nil
}
