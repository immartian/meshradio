// +build opus

package audio

import (
	"fmt"

	"gopkg.in/hraban/opus.v2"
)

// OpusCodec implements Opus audio encoding/decoding
type OpusCodec struct {
	encoder   *opus.Encoder
	decoder   *opus.Decoder
	frameSize int
	channels  int
}

// NewOpusCodec creates a new Opus codec
func NewOpusCodec(sampleRate, channels, frameSize int) (*OpusCodec, error) {
	// Create encoder
	encoder, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus encoder: %w", err)
	}

	// Set encoding parameters for voice
	encoder.SetBitrate(64000)                        // 64 kbps
	encoder.SetComplexity(10)                        // Max quality
	encoder.SetDTX(true)                             // Discontinuous transmission
	encoder.SetInBandFEC(true)                       // Forward error correction
	encoder.SetPacketLossPerc(10)                    // Expected 10% packet loss
	encoder.SetMaxBandwidth(opus.Fullband)           // Full bandwidth
	encoder.SetVBR(true)                             // Variable bitrate
	encoder.SetVBRConstraint(false)                  // Unconstrained VBR
	encoder.SetSignal(opus.SignalVoice)              // Optimize for voice

	// Create decoder
	decoder, err := opus.NewDecoder(sampleRate, channels)
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

// Encode compresses PCM audio
func (c *OpusCodec) Encode(pcm []byte) ([]byte, error) {
	// Convert bytes to int16 PCM
	pcmSamples := make([]int16, len(pcm)/2)
	for i := 0; i < len(pcmSamples); i++ {
		pcmSamples[i] = int16(pcm[i*2]) | int16(pcm[i*2+1])<<8
	}

	// Encode with Opus
	data := make([]byte, 4000) // Max Opus frame size
	n, err := c.encoder.Encode(pcmSamples, data)
	if err != nil {
		return nil, fmt.Errorf("Opus encode error: %w", err)
	}

	return data[:n], nil
}

// Decode decompresses Opus audio
func (c *OpusCodec) Decode(encoded []byte) ([]byte, error) {
	// Decode with Opus
	pcmSamples := make([]int16, c.frameSize*c.channels)
	n, err := c.decoder.Decode(encoded, pcmSamples)
	if err != nil {
		return nil, fmt.Errorf("Opus decode error: %w", err)
	}

	// Convert int16 to bytes
	pcm := make([]byte, n*2*c.channels)
	for i := 0; i < n*c.channels; i++ {
		pcm[i*2] = byte(pcmSamples[i])
		pcm[i*2+1] = byte(pcmSamples[i] >> 8)
	}

	return pcm, nil
}

// FrameSize returns the frame size
func (c *OpusCodec) FrameSize() int {
	return c.frameSize
}

// Reset resets the codec state
func (c *OpusCodec) Reset() error {
	if err := c.encoder.ResetState(); err != nil {
		return err
	}
	if err := c.decoder.ResetState(); err != nil {
		return err
	}
	return nil
}

// SetBitrate sets the encoding bitrate (in bps)
func (c *OpusCodec) SetBitrate(bitrate int) error {
	return c.encoder.SetBitrate(bitrate)
}

// SetComplexity sets encoding complexity (0-10)
func (c *OpusCodec) SetComplexity(complexity int) error {
	return c.encoder.SetComplexity(complexity)
}
