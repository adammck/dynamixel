package v1

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
		{0x1, "status error: input voltage"},
		{0x2, "status error: angle limit"},
		{0x4, "status error: overheating"},
		{0x8, "status error: range"},
		{0x10, "status error: checksum"},
		{0x20, "status error: overload"},
		{0x40, "status error: instruction"},
		{0x80, "status error: unknown"},
		{0x7, "status errors: input voltage, angle limit, overheating"},
		{0xFF, "status errors: input voltage, angle limit, overheating, range, checksum, overload, instruction, unknown"},
	}

	for _, eg := range examples {
		act := decodeError(eg.input)
		assert.EqualError(t, act, eg.output)
	}
}
