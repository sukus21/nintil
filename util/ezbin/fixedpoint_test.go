package ezbin

import (
	"testing"
)

func TestDecode(t *testing.T) {
	if DecodeFixedPoint(0x08, 0, 1, 3) != 1 {
		t.Fail()
	}
	if DecodeFixedPoint(0x30, 0, 3, 5) != 1.5 {
		t.Fail()
	}
	if DecodeFixedPoint(0xF0, 1, 3, 4) != -1 {
		t.Fail()
	}
}

func TestEncode(t *testing.T) {
	if EncodeFixedPoint(1, 0, 1, 3) != 0x08 {
		t.Fail()
	}
	if EncodeFixedPoint(1.5, 0, 3, 5) != 0x30 {
		t.Fail()
	}
	if EncodeFixedPoint(-1, 1, 3, 4) != 0xF0 {
		t.Fail()
	}
}
