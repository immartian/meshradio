package audio

import (
	"fmt"

	"layeh.com/gopus"
)

// OpusCodec implements Opus audio compression
type OpusCodec struct {
	encoder   *gopus.Encoder
	decoder   *gopus.Decoder
	frameSize int
	channels  int
}

// NewOpusCodec creates a new Opus codec
func NewOpusCodec(sampleRate, channels, frameSize, bitrate int) (*OpusCodec, error) {
	// Create encoder
	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus encoder: %w", err)
	}

	// Set bitrate
	encoder.SetBitrate(bitrate)

	// Create decoder
	decoder, err := gopus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus decoder: %w", err)
	}

	return &OpusCodec{
		encoder:   encoder,
		decoder:   decoder,
		frameSize: frameSize,
		channels:  channels,
	}, nil
}

// Encode compresses PCM audio to Opus
func (c *OpusCodec) Encode(pcm []byte) ([]byte, error) {
	// Convert bytes to int16 samples
	samples := make([]int16, len(pcm)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(pcm[i*2]) | int16(pcm[i*2+1])<<8
	}

	// Encode with Opus
	encoded, err := c.encoder.Encode(samples, c.frameSize, 4000)
	if err != nil {
		fmt.Printf("🔴 Opus encode error: %v (input=%d samples)\n", err, len(samples))
		return nil, fmt.Errorf("Opus encode failed: %w", err)
	}

	fmt.Printf("🟢 Opus encoded: %d bytes → %d bytes (%.1fx compression)\n",
		len(pcm), len(encoded), float64(len(pcm))/float64(len(encoded)))
	return encoded, nil
}

// Decode decompresses Opus to PCM audio
func (c *OpusCodec) Decode(encoded []byte) ([]byte, error) {
	// Decode with Opus
	samples, err := c.decoder.Decode(encoded, c.frameSize, false)
	if err != nil {
		fmt.Printf("🔴 Opus decode error: %v (input=%d bytes)\n", err, len(encoded))
		return nil, fmt.Errorf("Opus decode failed: %w", err)
	}

	// Convert int16 samples back to bytes
	pcm := make([]byte, len(samples)*2)
	for i := 0; i < len(samples); i++ {
		pcm[i*2] = byte(samples[i])
		pcm[i*2+1] = byte(samples[i] >> 8)
	}

	fmt.Printf("🟢 Opus decoded: %d bytes → %d bytes (%d samples)\n",
		len(encoded), len(pcm), len(samples))
	return pcm, nil
}

// FrameSize returns the frame size in samples
func (c *OpusCodec) FrameSize() int {
	return c.frameSize
}

// Reset resets the codec state
func (c *OpusCodec) Reset() error {
	c.encoder.ResetState()
	c.decoder.ResetState()
	return nil
}
