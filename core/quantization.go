package core

import "math"

// Quantizer defines the interface for vector quantization
// Quantization reduces the memory footprint of vectors by compressing them
// from 32-bit floats to smaller representations like 8-bit integers

type Quantizer interface {
	// Quantize compresses a float32 vector to a compressed representation
	Quantize(vector []float32) []byte

	// Dequantize decompresses a compressed representation back to float32
	Dequantize(data []byte) []float32

	// GetCompressedSize returns the size in bytes of a quantized vector
	GetCompressedSize(dim int) int
}

// SQ8Quantizer implements scalar quantization to 8-bit integers
// It scales each vector element to the range [0, 255] based on the
// min and max values of the vector

type SQ8Quantizer struct{}

// NewSQ8Quantizer creates a new SQ8 quantizer
func NewSQ8Quantizer() *SQ8Quantizer {
	return &SQ8Quantizer{}
}

// Quantize compresses a float32 vector to 8-bit integers
func (q *SQ8Quantizer) Quantize(vector []float32) []byte {
	if len(vector) == 0 {
		return []byte{}
	}

	// Find min and max values
	min := vector[0]
	max := vector[0]
	for _, v := range vector {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Calculate scale factor
	range_ := max - min
	scale := float32(255.0)
	if range_ > 0 {
		scale = 255.0 / range_
	}

	// Quantize each element
	result := make([]byte, len(vector)+8) // 4 bytes for min, 4 bytes for max

	// Store min and max as float32 (little-endian)
	storeFloat32(result[0:4], min)
	storeFloat32(result[4:8], max)

	// Quantize each element
	for i, v := range vector {
		// Scale to [0, 255], then shift to [-128, 127]
		qval := (v - min) * float32(scale)
		if qval < 0 {
			qval = 0
		} else if qval > 255 {
			qval = 255
		}
		result[i+8] = byte(qval)
	}

	return result
}

// Dequantize decompresses an 8-bit integer vector back to float32
func (q *SQ8Quantizer) Dequantize(data []byte) []float32 {
	if len(data) < 8 {
		return []float32{}
	}

	// Read min and max
	min := readFloat32(data[0:4])
	max := readFloat32(data[4:8])

	// Calculate scale factor
	range_ := max - min
	scale := float32(1.0)
	if range_ > 0 {
		scale = range_ / 255.0
	}

	// Dequantize each element
	vector := make([]float32, len(data)-8)
	for i := 8; i < len(data); i++ {
		// Just scale the byte value
		qval := float32(data[i])
		vector[i-8] = min + qval*float32(scale)
	}

	return vector
}

// GetCompressedSize returns the size in bytes of a quantized vector
func (q *SQ8Quantizer) GetCompressedSize(dim int) int {
	return dim + 8 // 1 byte per dimension + 8 bytes for min/max
}

// storeFloat32 stores a float32 as 4 bytes (little-endian)
func storeFloat32(buf []byte, val float32) {
	bits := math.Float32bits(val)
	buf[0] = byte(bits)
	buf[1] = byte(bits >> 8)
	buf[2] = byte(bits >> 16)
	buf[3] = byte(bits >> 24)
}

// readFloat32 reads a float32 from 4 bytes (little-endian)
func readFloat32(buf []byte) float32 {
	bits := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
	return math.Float32frombits(bits)
}
