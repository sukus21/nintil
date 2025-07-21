package ezbin

import (
	"io"
	"math"
)

func ReadFixedPoint(r io.Reader, signBits int, wholeBits int, fractBits int) (float64, error) {
	totalBits := signBits + wholeBits + fractBits

	// Validate bit count in fixed-point field
	if totalBits&0x7 != 0 || totalBits > 64 {
		return 0, ErrInvalidFixedPointSize
	}
	if signBits != 0 && signBits != 1 {
		return 0, ErrInvalidFixedPointSign
	}
	if wholeBits < 0 || fractBits < 0 {
		return 0, ErrInvalidFixedPointBits
	}

	// Read bits
	rawData, err := ReadBytesAsInt(r, totalBits/8)
	if err != nil {
		return 0, err
	}

	return DecodeFixedPoint(rawData, signBits, wholeBits, fractBits), nil
}

func DecodeFixedPoint(rawData uint64, signBits int, wholeBits int, fractBits int) float64 {
	totalBits := signBits + wholeBits + fractBits

	negative := signBits == 1 && (rawData&(1<<(totalBits-1))) != 0
	if negative {
		rawData ^= (1 << totalBits) - 1
		rawData += 1
		wholeBits += 1
	}

	fraction := rawData & ((1 << fractBits) - 1)
	whole := (rawData >> uint64(fractBits)) & ((1 << wholeBits) - 1)

	// Convert to float
	final := float64(whole)
	final += float64(fraction) / float64(int(1)<<fractBits)
	if negative {
		final = -final
	}

	return final
}

func EncodeFixedPoint(float float64, signBits int, wholeBits int, fractBits int) uint64 {
	isNegative := float < 0
	float = math.Abs(float)
	wholef, fractf := math.Modf(float)

	fractf *= float64(int(1) << fractBits)
	fract := uint64(math.Floor(fractf))
	whole := uint64(wholef) & ((1 << wholeBits) - 1)

	final := fract + (whole << fractBits)
	if isNegative {
		final = -final
		final &= (1 << (signBits + wholeBits + fractBits)) - 1
	}

	return final
}
