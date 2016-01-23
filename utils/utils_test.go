package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBytesToInt(t *testing.T) {
	examples := []struct {
		input  []byte
		err    bool
		output int
	}{
		{[]byte{0}, false, 0},
		{[]byte{1}, false, 1},
		{[]byte{0, 1}, false, 256},
		{[]byte{0, 2}, false, 512},
		{[]byte{1, 0}, false, 1},
		{[]byte{2, 0}, false, 2},
		{[]byte{1, 1}, false, 257},
		{[]byte{2, 2}, false, 514},
		{[]byte{0, 0, 0}, true, 0},
	}

	for _, eg := range examples {
		act, err := bytesToInt(eg.input)
		assert.Equal(t, act, eg.output)

		if eg.err {
			assert.Error(t, err)
		}
	}
}

func TestBoolToInt(t *testing.T) {
	assert.Equal(t, 1, BoolToInt(true))
	assert.Equal(t, 0, BoolToInt(false))
}

func TestItob(t *testing.T) {
	assert.Equal(t, true, IntToBool(-1))
	assert.Equal(t, false, IntToBool(0))
	assert.Equal(t, true, IntToBool(1))
	assert.Equal(t, true, IntToBool(2))
}

func TestLow(t *testing.T) {
	assert.Equal(t, byte(0), Low(0))
	assert.Equal(t, byte(0x1), Low(1))
	assert.Equal(t, byte(0xff), Low(1023))
	assert.Equal(t, byte(0x0), Low(1024))
	assert.Equal(t, byte(0x1), Low(1025))
}

func TestHigh(t *testing.T) {
	assert.Equal(t, byte(0), High(0))
	assert.Equal(t, byte(0), High(1))
	assert.Equal(t, byte(0x3), High(1023))
	assert.Equal(t, byte(0x4), High(1024))
	assert.Equal(t, byte(0x4), High(1025))
}
