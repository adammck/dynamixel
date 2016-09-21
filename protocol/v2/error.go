package v2

import (
	"errors"
	"fmt"
)

func decodeError(b byte) error {
	s := ""

	switch b {
	case 0x01:
		s = "result fail"

	case 0x02:
		s = "instruction error"

	case 0x03:
		s = "crc error"

	case 0x04:
		s = "data range error"

	case 0x05:
		s = "data length error"

	case 0x06:
		s = "data limit error"

	case 0x07:
		s = "access error"

	default:
		s = fmt.Sprintf("unknown error: 0x%X", b)
	}

	return errors.New(s)
}
