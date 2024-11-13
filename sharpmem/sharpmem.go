package sharpmem

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
)

const (
	bitWritecmd = 0x01
	bitVcom     = 0x02
	bitClear    = 0x04
)

type Device struct {
	bus        drivers.SPI
	csPin      machine.Pin
	buffer     []byte
	height     int16
	width      int16
	bufferSize int16
	lineDiff   []byte
	lineBuf    []byte
	vcom       uint8
}

func New(bus drivers.SPI, csPin machine.Pin) Device {
	d := Device{
		bus:    bus,
		csPin:  csPin,
		width:  160,
		height: 68,
	}

	d.csPin.Low()

	d.vcom = bitVcom

	d.bufferSize = d.width * d.height / 8
	d.buffer = make([]byte, d.bufferSize)

	d.lineDiff = make([]byte, d.height/8+1)

	for i := range d.buffer {
		d.buffer[i] = 0xff
	}

	var bytesPerLine = d.width / 8

	d.lineBuf = make([]byte, bytesPerLine+2)

	return d
}

func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	i := x + y*d.width

	div := i / 8
	mod := uint8(i % 8)

	curr := hasBit(d.buffer[div], mod)
	white := (c.R > 0 || c.G > 0 || c.B > 0) && c.A > 0
	var next bool
	if white {
		next = true
	}

	if next == curr {
		return
	}

	if white {
		d.buffer[div] = setBit(d.buffer[div], mod)
	} else {
		d.buffer[div] = clearBit(d.buffer[div], mod)
	}

	linediv := y / 8
	linemod := uint8(y % 8)
	d.lineDiff[linediv] = setBit(d.lineDiff[linediv], linemod)
}

func (d *Device) Size() (x, y int16) {
	return d.width, d.height
}

func (d *Device) Display() error {
	defer func() {
		for i := 0; i < len(d.lineDiff); i++ {
			d.lineDiff[i] = 0x00
		}
	}()

	// start transfer
	d.csPin.High()

	// send write command
	_, err := d.bus.Transfer(d.vcom | bitWritecmd)
	if err != nil {
		return err
	}

	d.toggleVcom()

	var bytesPerLine = d.width / 8
	var totalBytes = (d.width * d.height) / 8

	for i := int16(0); i < totalBytes; i += bytesPerLine {
		// first byte is current line (0-indexed)
		currentLine := (i + 1) / (d.width / 8)

		// skip rendering lines that haven't changed
		linediv := (currentLine) / 8
		linemod := uint8((currentLine) % 8)
		if !hasBit(d.lineDiff[linediv], linemod) {
			continue
		}

		d.lineBuf[0] = uint8(currentLine) + 1 // encode as 1-indexed

		// data bytes (copy to lineBuf to avoid modifications to buffer)
		copy(d.lineBuf[1:bytesPerLine+1], d.buffer[i:i+bytesPerLine])

		// last byte is always 0x00
		d.lineBuf[bytesPerLine+1] = 0x00

		// send the line data
		err = d.bus.Tx(d.lineBuf, nil)
		if err != nil {
			return err
		}
	}

	// trailer byte (always 0x00)
	_, err = d.bus.Transfer(0x00)
	if err != nil {
		return err
	}

	// end transfer
	d.csPin.Low()

	return nil
}

func (d *Device) Clear() error {
	// begin transaction
	d.csPin.High()

	b := []byte{
		d.vcom | bitClear,
		0x00,
	}

	err := d.bus.Tx(b, nil)
	if err != nil {
		return err
	}

	d.toggleVcom()

	// end transaction
	d.csPin.Low()

	return nil
}

// toggleVcom toggles the VCOM, as is instructed by the datasheet. Toggling VCOM
// can help maintain the display's longevity. It should ideally be called at
// least once per second.
func (d *Device) toggleVcom() {
	if d.vcom != 0 {
		d.vcom = 0x00
	} else {
		d.vcom = bitVcom
	}
}

func setBit(n uint8, pos uint8) uint8 {
	n |= 1 << pos
	return n
}

// Clears the bit at pos in n.
func clearBit(n uint8, pos uint8) uint8 {
	n &^= 1 << pos
	return n
}

func hasBit(n uint8, pos uint8) bool {
	n = n & (1 << pos)
	return n > 0
}
