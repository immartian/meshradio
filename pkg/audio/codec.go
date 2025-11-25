package audio

// Codec interface for audio encoding/decoding
type Codec interface {
	Encode(pcm []byte) ([]byte, error)
	Decode(encoded []byte) ([]byte, error)
	FrameSize() int
	Reset() error
}

// DummyCodec is a pass-through codec for MVP
// In production, this would be replaced with Opus
type DummyCodec struct {
	frameSize int
}

// NewDummyCodec creates a new dummy codec
func NewDummyCodec(frameSize int) *DummyCodec {
	return &DummyCodec{
		frameSize: frameSize,
	}
}

// Encode passes through the data (no encoding for MVP)
func (c *DummyCodec) Encode(pcm []byte) ([]byte, error) {
	// In production, this would use libopus to encode
	return pcm, nil
}

// Decode passes through the data (no decoding for MVP)
func (c *DummyCodec) Decode(encoded []byte) ([]byte, error) {
	// In production, this would use libopus to decode
	return encoded, nil
}

// FrameSize returns the frame size
func (c *DummyCodec) FrameSize() int {
	return c.frameSize
}

// Reset resets the codec state
func (c *DummyCodec) Reset() error {
	return nil
}
