package g3d

import "github.com/go-gl/mathgl/mgl64"

func getPivotMatrix(f boneMatrixFlags, a float64, b float64) mgl64.Mat3 {
	i := float64(1)
	if f.getNegOne() {
		b = -1
	}

	c := b
	if f.getNegC() {
		c = -b
	}

	d := a
	if f.getNegD() {
		d = -a
	}

	switch f.getForm() {
	case 0:
		return mgl64.Mat3{
			i, 0, 0,
			0, a, c,
			0, b, d,
		}.Transpose()
	case 1:
		return mgl64.Mat3{
			0, a, c,
			i, 0, 0,
			0, b, d,
		}.Transpose()
	case 2:
		return mgl64.Mat3{
			0, a, c,
			0, b, d,
			i, 0, 0,
		}.Transpose()
	case 3:
		return mgl64.Mat3{
			0, i, 0,
			a, 0, c,
			b, 0, d,
		}.Transpose()
	case 4:
		return mgl64.Mat3{
			a, 0, c,
			0, i, 0,
			b, 0, d,
		}.Transpose()
	case 5:
		return mgl64.Mat3{
			a, 0, c,
			b, 0, d,
			0, i, 0,
		}.Transpose()
	case 6:
		return mgl64.Mat3{
			0, 0, i,
			a, c, 0,
			b, d, 0,
		}.Transpose()
	case 7:
		return mgl64.Mat3{
			a, c, 0,
			0, 0, i,
			b, d, 0,
		}.Transpose()
	case 8:
		return mgl64.Mat3{
			a, c, 0,
			b, d, 0,
			0, 0, i,
		}.Transpose()
	default:
		panic("getPivotMatrix: invalid form")
	}
}
