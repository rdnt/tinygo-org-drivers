package main

import (
	"image/color"
	"machine"
	"math/rand/v2"
	"time"

	"tinygo.org/x/drivers/sharpmem"
)

func initSPI() error {
	machine.P0_06.Configure(machine.PinConfig{Mode: machine.PinOutput})

	err := machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 2000000,
		SCK:       machine.P0_20,
		SDO:       machine.P0_17,
		SDI:       machine.P0_25,
		Mode:      0,
		LSBFirst:  true,
	})
	if err != nil {
		println("spi.Configure() failed, error:", err.Error())
		return err
	}

	return nil
}

func main() {
	time.Sleep(time.Second)

	err := initSPI()
	if err != nil {
		return
	}

	cfg := sharpmem.ConfigLS011B7DH03

	display := sharpmem.New(machine.SPI0, machine.P0_06)
	display.Configure(cfg)

	err = display.Clear()
	if err != nil {
		println("display.Clear() failed, error:", err.Error())
		return
	}

	for {

		x0 := int16(rand.IntN(int(cfg.Width - 7)))
		y0 := int16(rand.IntN(int(cfg.Height - 7)))

		for x2 := int16(0); x2 < 16; x2++ {
			x2 := x2
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

			if x2 >= 8 {
				x2 = x2 - 8
				c = color.RGBA{R: 0, G: 0, B: 0, A: 255}
			}

			for x := int16(0); x < x2; x++ {
				for y := int16(0); y < 8; y++ {
					display.SetPixel(x0+x, y0+y, c)
				}
			}

			err = display.Display()
			if err != nil {
				println("display.Display() failed, error:", err.Error())
				continue
			}

			time.Sleep(33 * time.Millisecond)
		}

		time.Sleep(33 * time.Millisecond)

		display.ClearBuffer()
	}
}
