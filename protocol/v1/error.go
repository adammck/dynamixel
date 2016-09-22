package v1

import (
	"fmt"
	"strings"
)

// DecodeStartusError Converts an error byte (as included in a status packet)
// into an error object with a friendly error message. We can't be too specific
// about it, because any combination of errors might occur at the same time.
//
// See: http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm#Status_Packet
func decodeError(b byte) error {
	str := []string{}

	if b == 0 {
		return fmt.Errorf("no error")
	}

	if b&1 == 1 {
		str = append(str, "input voltage")
	}

	if b&2 == 2 {
		str = append(str, "angle limit")
	}

	if b&4 == 4 {
		str = append(str, "overheating")
	}

	if b&8 == 8 {
		str = append(str, "range")
	}

	if b&16 == 16 {
		str = append(str, "checksum")
	}

	if b&32 == 32 {
		str = append(str, "overload")
	}

	if b&64 == 64 {
		str = append(str, "instruction")
	}

	if b&128 == 128 {
		str = append(str, "unknown")
	}

	s := ""
	if len(str) > 1 {
		s = "s"
	}

	return fmt.Errorf("status error%s: %s", s, strings.Join(str, ", "))
}
