package sharpmem

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func Test_setBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(setBit(0x00, i), qt.Equals, v)
		c.Assert(setBit(0x00, (i+1)%8), qt.Not(qt.Equals), v)
	}
}

func Test_unsetBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(unsetBit(v, i), qt.Equals, uint8(0x00))
		c.Assert(unsetBit(v, (i+1)%8), qt.Not(qt.Equals), uint8(0x00))
	}
}

func Test_hasBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(hasBit(v, i), qt.Equals, true)
		c.Assert(hasBit(v, (i+1)%8), qt.Equals, false)
	}
}

func Test_bitfieldBufLen(t *testing.T) {
	c := qt.New(t)

	for i := 1; i < 536; i++ {
		requiredBufferSize := i / 8
		wouldOverflow := i % 8

		if wouldOverflow > 0 {
			requiredBufferSize += 1
		}

		c.Assert(bitfieldBufLen(int16(i)), qt.Equals, int16(requiredBufferSize))
	}
}
