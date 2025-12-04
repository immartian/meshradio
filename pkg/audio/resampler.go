package audio

// SimpleResampler performs basic linear interpolation resampling
// Not the highest quality, but pure Go with no dependencies
type SimpleResampler struct {
	fromRate int
	toRate   int
	channels int
	buffer   []int16
}

// NewSimpleResampler creates a simple resampler
func NewSimpleResampler(fromRate, toRate, channels int) *SimpleResampler {
	return &SimpleResampler{
		fromRate: fromRate,
		toRate:   toRate,
		channels: channels,
		buffer:   make([]int16, 0),
	}
}

// Resample converts sample rate using linear interpolation
func (r *SimpleResampler) Resample(input []int16) []int16 {
	if r.fromRate == r.toRate {
		return input
	}

	ratio := float64(r.toRate) / float64(r.fromRate)

	// Calculate number of frames (not samples)
	inputFrames := len(input) / r.channels
	outputFrames := int(float64(inputFrames) * ratio)
	output := make([]int16, outputFrames*r.channels)

	for i := 0; i < outputFrames; i++ {
		// Calculate position in input frames
		srcPos := float64(i) / ratio
		srcIndex := int(srcPos)
		frac := srcPos - float64(srcIndex)

		// Handle per-channel for stereo/mono
		for ch := 0; ch < r.channels; ch++ {
			srcIdx := srcIndex*r.channels + ch
			outIdx := i*r.channels + ch

			if srcIdx+r.channels < len(input) {
				// Linear interpolation between two samples
				sample1 := float64(input[srcIdx])
				sample2 := float64(input[srcIdx+r.channels])
				interpolated := sample1 + (sample2-sample1)*frac
				output[outIdx] = int16(interpolated)
			} else if srcIdx < len(input) {
				// Just use the last sample
				output[outIdx] = input[srcIdx]
			}
		}
	}

	return output
}
