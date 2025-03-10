package nsb

type Fixed16 uint16

func (nf Fixed16) Float32() float32 {
	f := nf
	if nf&(1<<15) != 0 {
		f = (-nf) + 1
	}
	fract := f & ((1 << 12) - 1)
	whole := f >> 12

	flt := float32(whole) + float32(fract)/(1<<12)
	if nf&(1<<15) != 0 {
		flt = -flt
	}
	return flt
}

func (nf Fixed16) Float64() float64 {
	f := nf
	if nf&(1<<15) != 0 {
		f = (-nf) + 1
	}
	fract := f & ((1 << 12) - 1)
	whole := f >> 12

	flt := float64(whole) + float64(fract)/(1<<12)
	if nf&(1<<15) != 0 {
		flt = -flt
	}
	return flt
}

type Fixed32 uint32

func (nf Fixed32) Float32() float32 {
	f := nf
	if nf&(1<<31) != 0 {
		f = (-nf) + 1
	}
	fract := f & ((1 << 12) - 1)
	whole := f >> 12

	flt := float32(whole) + float32(fract)/(1<<12)
	if nf&(1<<31) != 0 {
		flt = -flt
	}
	return flt
}

func (nf Fixed32) Float64() float64 {
	f := nf
	if nf&(1<<31) != 0 {
		f = (-nf) + 1
	}
	fract := f & ((1 << 12) - 1)
	whole := f >> 12

	flt := float64(whole) + float64(fract)/(1<<12)
	if nf&(1<<31) != 0 {
		flt = -flt
	}
	return flt
}
