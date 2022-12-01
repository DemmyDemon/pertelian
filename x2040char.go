package pertelian

import (
	"errors"
)

const charSize = 8

// PertelianX2040Character holds the 8 bytes, one for each line, that make up a custom character.
type PertelianX2040Character struct {
	// Lines is the byte array that actually contains the data.
	Lines [charSize]byte
}

var (
	// ErrX2040CharWant8Lines is returned if NewX2040Char() does not get exactly 8 strings.
	ErrX2040CharWant8Lines = errors.New("characters must be made up of exactly 8 lines")

	// ErrX2040CharWidth5 is returned when NewX2040Char() is given a string that is not exactly 5 runes long.
	ErrX2040CharWidth5 = errors.New("character lines must be exactly 5 runes long")
)

// NewX2040Char assists in creating a properly formatted custom character for the display.
// It takes exactly 8 strings of exactly 5 runes each. Any rune containing a space will be blank on the display, any rune that is not a space represents a filled dot on the display.
// The SetLineDrawing method contains an example of how to use this.
func NewX2040Char(lines ...string) (PertelianX2040Character, error) {
	char := PertelianX2040Character{
		Lines: [charSize]byte{},
	}
	if len(lines) != len(char.Lines) {
		return char, ErrX2040CharWant8Lines
	}
	for i, line := range lines {
		if len(line) != 5 {
			return char, ErrX2040CharWidth5
		}
		final := byte(0)
		for j := 0; j < 5; j++ {
			if line[j] != ' ' {
				final |= 1 << (4 - j)
			}
		}
		char.Lines[i] = final
	}
	return char, nil
}
