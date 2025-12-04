package audio

// AudioSource is an interface for audio input sources
// Implementations can be: microphone, MP3 file, streaming URL, etc.
type AudioSource interface {
	// Read reads the next audio frame
	// Returns PCM samples in int16 format at the configured sample rate
	Read() ([]int16, error)

	// Start starts the audio source
	Start() error

	// Stop stops the audio source
	Stop() error

	// SampleRate returns the source's sample rate
	SampleRate() int

	// Channels returns the number of audio channels (1=mono, 2=stereo)
	Channels() int

	// IsRunning returns whether the source is currently running
	IsRunning() bool
}
