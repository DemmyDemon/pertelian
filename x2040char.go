package pertelian

import (
	"errors"
)

const charSize = 8

type PertelianX2040Character struct {
	Lines [charSize]byte
}

var ErrX2040CharWant8Lines = errors.New("characters must be made up of exactly 8 lines")
var ErrX2040CharWidth5 = errors.New("character lines must be exactly 5 runes long")

// NewX2040Char assists in creating a properly formatted custom character for the display.
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
