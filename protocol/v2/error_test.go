package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeError(t *testing.T) {
	examples := []struct {
		input  byte
		output string
	}{
		{0x0, "no error"},
		{0x1, "result fail"},
		{0x2, "instruction error"},
		{0x3, "crc error"},
		{0x4, "data range error"},
		{0x5, "data length error"},
		{0x6, "data limit error"},
		{0x7, "access error"},
		{0x8, "unknown error: 0x08"},
	}

	for _, eg := range examples {
		act := decodeError(eg.input)
		assert.EqualError(t, act, eg.output)
	}
}
