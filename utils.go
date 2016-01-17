package dynamixel

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

// btoi converts a bool to an int.
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// itob converts an int to a bool.
func itob(v int) bool {
	return (v != 0)
}

func low(i int) byte {
	return byte(i & 0xFF)
}

func high(i int) byte {
	return low(i >> 8)
}
