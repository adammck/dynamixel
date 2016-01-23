package dynamixel

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

func TestBtoi(t *testing.T) {
	assert.Equal(t, 1, btoi(true))
	assert.Equal(t, 0, btoi(false))
}

func TestItob(t *testing.T) {
	assert.Equal(t, true, itob(-1))
	assert.Equal(t, false, itob(0))
	assert.Equal(t, true, itob(1))
	assert.Equal(t, true, itob(2))
}

func TestLow(t *testing.T) {
	assert.Equal(t, byte(0), low(0))
	assert.Equal(t, byte(0x1), low(1))
	assert.Equal(t, byte(0xff), low(1023))
	assert.Equal(t, byte(0x0), low(1024))
	assert.Equal(t, byte(0x1), low(1025))
}

func TestHigh(t *testing.T) {
	assert.Equal(t, byte(0), high(0))
	assert.Equal(t, byte(0), high(1))
	assert.Equal(t, byte(0x3), high(1023))
	assert.Equal(t, byte(0x4), high(1024))
	assert.Equal(t, byte(0x4), high(1025))
}
