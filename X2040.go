package pertelian

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/google/gousb"
)

const (
	x2040Command  = byte(0xfe)
	x2040Clear    = byte(0x1)
	x2040LightOff = byte(0x2)
	x2040LightOn  = byte(0x3)
	x2040Entry    = byte(0x6)
	x2040Off      = byte(0x8)
	x2040On       = byte(0xc)
	x2040Init     = byte(0x38)
)

// PertelianX2040 keeps all the pointers and receives all the methods for interacting with your Pertelian X2040 display.
type PertelianX2040 struct {
	device    *gousb.Device
	iface     *gousb.Interface
	ep        *gousb.OutEndpoint
	ifaceDone func()
}

var (
	// ErrX2040UnknownWriteError is returned when there is an unexpected error writing to the device.
	ErrX2040UnknownWriteError = errors.New("unknown error writing to device")

	// ErrX2040DeviceNotFound is returned when attempting to open a Pertelian X2040, but no such device is found. (A device matching VID=0x0403,PID=0x6001 is expected)
	ErrX2040DeviceNotFound = errors.New("x2040 device not found")

	// ErrX2040OutOfRange is returned when attempting to write to a position that is outside the edges of the display, such as line 4 (0-3 are valid)
	ErrX2040OutOfRange = errors.New("target out of display range")

	// ErrX2040InvalidCharacterPosition is returned when attempting to access a custom display character that is out-of-bounds, such as character 7 (0-6 are valid)
	ErrX2040InvalidCharacterPosition = errors.New("invalid character position")

	// offsets hold the character offsets for each line in the display.
	offsets = [4]byte{
		0x80,
		0x80 + 0x40,
		0x80 + 0x14,
		0x80 + 0x54,
	}
)

// NewX2040 instantiates a new PertelianX2040 for you to play with.
func NewX2040(ctx *gousb.Context) (PertelianX2040, error) {
	pert := PertelianX2040{}

	device, err := ctx.OpenDeviceWithVIDPID(0x0403, 0x6001)
	if err != nil {
		return pert, fmt.Errorf("obtain device: %w", err)
	}
	if device == nil {
		return pert, ErrX2040DeviceNotFound
	}

	iface, done, err := device.DefaultInterface()
	if err != nil {
		return pert, fmt.Errorf("default interface: %w", err)
	}

	ep, err := iface.OutEndpoint(2)
	if err != nil {
		return pert, fmt.Errorf("open endpoint: %w", err)
	}
	pert.device = device
	pert.iface = iface
	pert.ifaceDone = done
	pert.ep = ep
	return pert, nil
}

// WriteGibberish writes directly to the device with no waiting for command processing, which *often* leads to corruption.
func (pert *PertelianX2040) WriteGibberish(data []byte) (int, error) {
	return pert.ep.Write(data)
}

// Write writes directly to the device, pausing slightly after the first two characters.
// Note that this is done in one transmission per byte to minimize the chance of corruption.
func (pert *PertelianX2040) Write(data []byte) (int, error) {
	written := 0
	for i := 0; i < len(data); i++ {
		if i <= 2 {
			time.Sleep(1 * time.Microsecond)
		}
		w, err := pert.ep.Write(data[i : i+1])
		written += w
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

// inst sends a command prefix, and then the given byte as an instruction.
func (pert *PertelianX2040) inst(action byte) error {
	written, err := pert.Write([]byte{x2040Command, action})
	if err != nil {
		return err
	}
	if written != 2 {
		return ErrX2040UnknownWriteError
	}
	return nil
}

// do takes a list of actions, sending them to .inst one at a time.
func (pert *PertelianX2040) do(actions ...byte) error {
	for _, action := range actions {
		err := pert.inst(action)
		if err != nil {
			return err
		}
	}
	return nil
}

// Print sends a string entry to wherever the cursor happens to be.
func (pert *PertelianX2040) Print(text string) error {

	output := []byte{x2040Command, x2040Entry}
	output = append(output, text...)
	_, err := pert.Write(output)

	return err
}

// PrintAt sends a string entry to the given line and character.
func (pert *PertelianX2040) PrintAt(line uint8, char uint8, textString string) error {
	text := []byte(textString)
	if line > 3 {
		return ErrX2040OutOfRange
	}
	if len(text) > 20 {
		return ErrX2040OutOfRange
	}
	if char+uint8(len(text)) > 20 {
		return ErrX2040OutOfRange
	}
	offset := offsets[line] + char
	output := []byte{x2040Command, offset}
	output = append(output, text...)
	_, err := pert.Write(output)
	return err
}

// Centered attempts to send the given string so it's centered on the given line.
func (pert *PertelianX2040) Centered(line uint8, text string) error {
	if len(text) > 20 {
		return ErrX2040OutOfRange
	}
	offset := uint8((20 - len(text)) / 2)
	return pert.PrintAt(line, offset, text)
}

// Close closes down the interface and closes the device.
// **DOES NOT** clear the display, turn off the light, or any of that.
func (pert *PertelianX2040) Close() error {
	pert.ifaceDone()
	pert.iface = nil
	pert.ifaceDone = nil
	return pert.device.Close()
}

// On turns on the display, initializes it, clears any data already on there and turns the light on.
func (pert *PertelianX2040) On() error {
	return pert.do(x2040On, x2040Init, x2040Clear, x2040LightOn)
}

// Off turns off the light, and then the display.
func (pert *PertelianX2040) Off() error {
	return pert.do(x2040LightOff, x2040Off)
}

// Clear removes all data visible on the display.
func (pert *PertelianX2040) Clear() error {
	return pert.inst(x2040Clear)
}

// Blank overwrites the given line with all spaces, effectively blanking it.
func (pert *PertelianX2040) Blank(line uint8) error {
	if line > 3 {
		return ErrX2040OutOfRange
	}
	return pert.PrintAt(line, 0, strings.Repeat(" ", 20))
}

// Light sets the display light to the requested state.
func (pert *PertelianX2040) Light(state bool) error {
	if state {
		return pert.inst(x2040LightOn)
	} else {
		return pert.inst(x2040LightOff)
	}
}

// SetCharacter stores the given character in the display for later display.
// You get 7 slots, 0-6.
func (pert *PertelianX2040) SetCharacter(position uint8, char PertelianX2040Character) error {
	if position > 6 {
		return ErrX2040InvalidCharacterPosition
	}
	position *= 8
	position += 72
	output := []byte{x2040Command, position}
	output = append(output, char.Lines[0:8]...)
	_, err := pert.Write(output)
	return err
}

// GetCharacters returns a string built with the characters previously stored with SetCharacter.
func (pert *PertelianX2040) GetCharacters(slots ...uint8) string {
	length := len(slots)
	if length == 0 {
		return ""
	}
	if length > 20 {
		length = 20
	}
	output := make([]byte, length)
	for i := 0; i < len(slots); i++ {
		output[i] = slots[i] + 1
	}
	return string(output)
}

// SetLineDrawingCharacters stores a set of line drawing characters in the display.
func (pert *PertelianX2040) SetLineDrawingCharacters() {
	char0, _ := NewX2040Char(
		"     ",
		"     ",
		"     ",
		"   ##",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
	)
	pert.SetCharacter(0, char0)

	char1, _ := NewX2040Char(
		"     ",
		"     ",
		"     ",
		"#####",
		"     ",
		"     ",
		"     ",
		"     ",
	)
	pert.SetCharacter(1, char1)

	char2, _ := NewX2040Char(
		"     ",
		"     ",
		"     ",
		"##   ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
	)
	pert.SetCharacter(2, char2)

	char3, _ := NewX2040Char(
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
	)
	pert.SetCharacter(3, char3)

	char4, _ := NewX2040Char(
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"   ##",
		"     ",
		"     ",
		"     ",
	)
	pert.SetCharacter(4, char4)

	char5, _ := NewX2040Char(
		"     ",
		"     ",
		"     ",
		"     ",
		"#####",
		"     ",
		"     ",
		"     ",
	)
	pert.SetCharacter(5, char5)

	char6, _ := NewX2040Char(
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"##   ",
		"     ",
		"     ",
		"     ",
	)
	pert.SetCharacter(6, char6)

}

// Splash does some line drawing and text to look fancy.
func (pert *PertelianX2040) Splash() {
	pert.SetLineDrawingCharacters()

	pert.PrintAt(0, 0, pert.GetCharacters(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2))

	pert.PrintAt(1, 0, pert.GetCharacters(3))
	pert.Centered(1, "Pertelian  X2040")
	pert.PrintAt(1, 19, pert.GetCharacters(3))

	pert.PrintAt(2, 0, pert.GetCharacters(3))
	pert.Centered(2, runtime.Version())
	pert.PrintAt(2, 19, pert.GetCharacters(3))

	pert.PrintAt(3, 0, pert.GetCharacters(4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6))
}
