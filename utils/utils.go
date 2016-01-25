package utils

import (
	"fmt"
)

// bytesToInt converts a slice of bytes to an int. Only 8- and 16-bit uints are
// supported, because those are the only thing the control tables contain.
func bytesToInt(b []byte) (int, error) {

	switch len(b) {
	case 1:
		return int(b[0]), nil

	case 2:
		return int(b[0]) | int(b[1])<<8, nil

	default:
		return 0, fmt.Errorf("invalid read length %d", len(b))

	}
}

// BoolToInt converts a bool to an int.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IntToBool converts an int to a bool.
func IntToBool(v int) bool {
	return (v != 0)
}

func Low(i int) byte {
	return byte(i & 0xFF)
}

func High(i int) byte {
	return Low(i >> 8)
}
