//go:build tinygo

package sharpmem

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
)

const (
	bitWriteCmd = 0x01
	bitVcom     = 0x02
	bitClear    = 0x04
)

// Device represents a Sharp Memory Display device.
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

type Config struct {
	Width  int16
	Height int16
}

// NewSPI creates a new connection.
// The SPI bus must have already been configured.
func NewSPI(bus drivers.SPI, csPin machine.Pin) Device {
	d := Device{
		bus:   bus,
		csPin: csPin,
	}
	return d
}

// Configure initializes the display with specified configuration.
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 64
	}

	d.initialize()
}

// initialize properly initializes the display and the in-memory image buffers.
func (d *Device) initialize() {
	d.csPin.Low()

	d.vcom = bitVcom

	d.bufferSize = d.width * d.height / 8
	d.buffer = make([]byte, d.bufferSize)

	d.lineDiff = make([]byte, bitfieldBufLen(d.height))

	for i := range d.buffer {
		d.buffer[i] = 0xff
	}

	var bytesPerLine = d.width / 8

	d.lineBuf = make([]byte, bytesPerLine+2)
}

// SetPixel enables or disables a pixel in the buffer
// color.RGBA{0, 0, 0, 255} is considered transparent, anything else
// will enable a pixel on the screen.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	i := x + y*d.width

	div := i / 8
	mod := uint8(i % 8)

	prev := hasBit(d.buffer[div], mod)
	var curr bool
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 255 {
		curr = true
	}

	if prev == curr {
		return
	}

	if curr {
		d.buffer[div] = setBit(d.buffer[div], mod)
	} else {
		d.buffer[div] = unsetBit(d.buffer[div], mod)
	}

	linediv := y / 8
	linemod := uint8(y % 8)
	d.lineDiff[linediv] = setBit(d.lineDiff[linediv], linemod)
}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	return d.width, d.height
}

// Display sends the whole buffer to the screen. If a line hasn't changed,
// it will not be transferred. It should ideally be called at >=1hz.
func (d *Device) Display() error {
	defer func() {
		for i := 0; i < len(d.lineDiff); i++ {
			d.lineDiff[i] = 0x00
		}
	}()

	// start transfer
	d.csPin.High()

	// send write command
	_, err := d.bus.Transfer(d.vcom | bitWriteCmd)
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

// Clear clears both the in-memory buffer and the display.
func (d *Device) Clear() error {
	d.ClearBuffer()
	return d.ClearDisplay()
}

// ClearBuffer clears the in-memory buffer. The display is not updated.
func (d *Device) ClearBuffer() {
	for i := 0; i < len(d.buffer); i++ {
		d.buffer[i] = 0xff
	}

	for i := 0; i < len(d.lineDiff); i++ {
		d.lineDiff[i] = 0x00
	}
}

// ClearDisplay clears the display. The in-memory buffer is not updated.
func (d *Device) ClearDisplay() error {
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

// toggleVcom toggles the VCOM, as is instructed by the datasheet.
// Toggling VCOM can help maintain the display's longevity. It should ideally
// be called at least once per second, preferably at 4-100 Hz.
// Toggling VCOM causes a tiny bit of flicker, but without it the pixels can
// be permanently damaged by the DC bias accumulating over time.
func (d *Device) toggleVcom() {
	if d.vcom != 0 {
		d.vcom = 0x00
	} else {
		d.vcom = bitVcom
	}
}
