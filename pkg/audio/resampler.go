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
	outputLen := int(float64(len(input)) * ratio)
	output := make([]int16, outputLen)

	for i := 0; i < outputLen; i++ {
		// Calculate position in input array
		srcPos := float64(i) / ratio
		srcIndex := int(srcPos)
		frac := srcPos - float64(srcIndex)

		// Handle per-channel for stereo/mono
		for ch := 0; ch < r.channels; ch++ {
			idx := srcIndex*r.channels + ch

			if idx+r.channels < len(input) {
				// Linear interpolation between two samples
				sample1 := float64(input[idx])
				sample2 := float64(input[idx+r.channels])
				interpolated := sample1 + (sample2-sample1)*frac
				output[i*r.channels+ch] = int16(interpolated)
			} else if idx < len(input) {
				// Just use the last sample
				output[i*r.channels+ch] = input[idx]
			}
		}
	}

	return output
}
