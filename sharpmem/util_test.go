package sharpmem

import (
	"image/color"
	"math/rand/v2"
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

	for i := int16(1); i < 536; i++ {
		requiredBufferSize := i / 8
		wouldOverflow := i % 8

		if wouldOverflow > 0 {
			requiredBufferSize += 1
		}

		c.Assert(bitfieldBufLen(i), qt.Equals, requiredBufferSize)
	}
}

type mockBus struct{}

func (m mockBus) Tx(w, r []byte) error {
	return nil
}

func (m mockBus) Transfer(b byte) (byte, error) {
	return b, nil
}

type mockPin struct{}

func (m mockPin) High() {
}

func (m mockPin) Low() {
}

func Test_Device(t *testing.T) {
	c := qt.New(t)

	cfgs := []Config{
		{Width: 128, Height: 128}, // LS010B7DH04, LS013B7DH03
		{Width: 160, Height: 68},  // LS011B7DH03
		{Width: 184, Height: 38},  // LS012B7DD01
		{Width: 144, Height: 168}, // LS013B7DH05
		{Width: 230, Height: 303}, // LS018B7DH02
		{Width: 400, Height: 240}, // LS027B7DH01, LS027B7DH01A
		{Width: 336, Height: 536}, //LS032B7DD02
		{Width: 320, Height: 240}, //LS044Q7DH01
	}

	cfgLen := len(cfgs)
	for i := 0; i < cfgLen; i++ {
		cfgs = append(cfgs, Config{
			Width:                cfgs[i].Width,
			Height:               cfgs[i].Height,
			DisableOptimizations: true,
		})
	}

	spi := mockBus{}
	pin := mockPin{}
	display := New(spi, pin)

	for _, cfg := range cfgs {
		display.Configure(cfg)

		x, y := display.Size()
		c.Assert(x, qt.Equals, cfg.Width)
		c.Assert(y, qt.Equals, cfg.Height)

		for i := 0; i < 10; i++ {
			x := int16(rand.IntN(int(cfg.Width)))
			y := int16(rand.IntN(int(cfg.Height)))
			display.SetPixel(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}

		for i := 0; i < 10; i++ {
			x := int16(rand.IntN(int(cfg.Width)))
			y := int16(rand.IntN(int(cfg.Height)))
			display.SetPixel(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}

		err := display.Display()
		c.Assert(err, qt.Equals, nil)

		err = display.ClearDisplay()
		c.Assert(err, qt.Equals, nil)

		display.ClearBuffer()
	}
}
