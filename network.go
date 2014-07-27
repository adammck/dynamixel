package dynamixel

import (
	"bytes"
	"strings"
	"encoding/binary"
	"fmt"
	"io"
)

const (

	// Instruction Types
	Ping      byte = 0x01
	ReadData  byte = 0x02
	WriteData byte = 0x03
	RegWrite  byte = 0x04
	Action    byte = 0x05
	Reset     byte = 0x06
	SyncWrite byte = 0x83

	// Send an instruction to all servos
	BroadcastIdent byte = 0xFE // 254
)

type DynamixelNetwork struct {
	Serial   io.ReadWriteCloser
	Buffered bool
}

func NewNetwork(serial io.ReadWriteCloser) *DynamixelNetwork {
	return &DynamixelNetwork{
		Serial: serial,
		Buffered: false,
	}
}

//
// Puts the network in bufferred write mode, which means that the REG_WRITE
// instruction will be used, rather than WRITE_DATA. This causes calls to
// WriteData to be bufferred until the Action method is called, at which time
// they'll all be executed at once.
//
// This is very useful for synchronizing the movements of multiple servos.
//
func (n *DynamixelNetwork) SetBuffered(buffered bool) {
	n.Buffered = buffered
}

//
// Converts an error byte (as included in a status packet) into an error object
// with a friendly error message. We can't be too specific about it, because any
// combination of errors might occur at the same time.
//
// See: http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm#Status_Packet
//
func DecodeStatusError(errBits byte) error {
	str := []string{}

	if(errBits & 1 == 1) {
		str = append(str, "input voltage")
	}

	if(errBits & 2 == 2) {
		str = append(str, "angle limit")
	}

	if(errBits & 4 == 4) {
		str = append(str, "overheating")
	}

	if(errBits & 8 == 8) {
		str = append(str, "range")
	}

	if(errBits & 16 == 16) {
		str = append(str, "checksum")
	}

	if(errBits & 32 == 32) {
		str = append(str, "overload")
	}

	if(errBits & 64 == 64) {
		str = append(str, "instruction")
	}

	if(errBits & 128 == 128) {
		str = append(str, "unknown")
	}

	return fmt.Errorf("status error(s): %s", strings.Join(str, ", "))
}

//
// This stuff is generic to all Dynamixels. See:
//
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_instruction.htm
//
func (n *DynamixelNetwork) WriteInstruction(ident uint8, instruction byte, params ...byte) error {

	buf := new(bytes.Buffer)
	paramsLength := byte(len(params) + 2)

	// build instruction packet

	buf.Write([]byte{
		0xFF,               // header
		0xFF,               // header
		byte(ident),        // target Dynamixel ID
		byte(paramsLength), // len(params) + 2
		instruction,        // instruction type (read/write/etc)
	})

	buf.Write(params)

	// calculate checksum

	sum := ident + paramsLength + instruction

	for _, value := range params {
		sum += value
	}

	buf.WriteByte(byte((^sum) & 0xFF))

	// write to port

	_, err := buf.WriteTo(n.Serial)

	if err != nil {
		return err
	}

	return nil
}

func (network *DynamixelNetwork) ReadStatusPacket(expectIdent uint8) ([]byte, error) {

	//
	// Status packets are similar to instruction packet:
	//
	// +------+------+-------+----------+---------+--------+--------+----------+
	// | 0xFF | 0xFF | ident | params+2 | errBits | param1 | param2 | checksum |
	// +------+------+-------+----------+---------+--------+--------+----------+
	//
	// We don't know the length, because responses can have variable numbers of
	// parameters.
	//

	// read the first five bytes, which are always present. we don't know how many
	// parameters follow that, until we've read buf[3]

	buf := make([]byte, 5)
	n, err := network.Serial.Read(buf)
	if n == 0 && err != nil {
		return []byte{}, err
	}

	if buf[0] != 0xFF || buf[1] != 0xFF {
		return []byte{}, fmt.Errorf("bad status packet header: %x %x", buf[0], buf[1])
	}

	resIdent := uint8(buf[2])
	numParams := uint8(buf[3]) - 2
	errBits := buf[4]

	// now read the params, if there are any. we must do this before checking for
	// errors, to avoid leaving junk in the buffer.

	paramsBuf := make([]byte, numParams)
	if numParams > 0 {
		n2, err2 := network.Serial.Read(paramsBuf)
		if n2 == 0 && err2 != nil {
			return []byte{}, err2
		}
	}

	// read the checksum, which is always one byte
	// TODO: check the checksum

	checksumBuf := make([]byte, 1)
	n3, err3 := network.Serial.Read(checksumBuf)
	if n3 == 0 && err3 != nil {
		return []byte{}, err3
	}

	// return an error if the packet contained one.
	// TODO: decode the error bit(s) and return a proper message!

	if errBits != 0x0 {
		return []byte{}, DecodeStatusError(errBits)
	}

	// return an error if we received a packet with the wrong ID. this indicates
	// a concurrency issue (maybe clashing IDs on a single bus).

	if resIdent != expectIdent {
		return []byte{}, fmt.Errorf("expected status packet for %v, but got %v", expectIdent, resIdent)
	}

	// omg, nothing went wrong

	return paramsBuf, nil
}

func (n *DynamixelNetwork) ReadData(ident uint8, startAddress byte, length int) (uint16, error) {
	params := []byte{startAddress, byte(length)}
	err1 := n.WriteInstruction(ident, ReadData, params...)
	if err1 != nil {
		return 0, err1
	}

	buf, err2 := n.ReadStatusPacket(ident)
	if err2 != nil {
		return 0, err2
	}
	var val uint16
	err3 := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &val)
	if err3 != nil {
		return 0, err3
	}

	return val, nil
}

func (n *DynamixelNetwork) WriteData(ident uint8, params ...byte) error {
	var instruction byte

	if(n.Buffered) {
		instruction = RegWrite
	} else {
		instruction = WriteData
	}

	writeErr := n.WriteInstruction(ident, instruction, params...)
	if writeErr != nil {
		return writeErr
	}

	_, readErr := n.ReadStatusPacket(ident)
	if readErr != nil {
		return readErr
	}

	return nil
}

//
// Broadcasts the ACTION instruction, which initiates any previously bufferred
// instructions.
//
// Doesn't wait for a status packet in response, because they are not sent in
// response to broadcast instructions.
//
func (n *DynamixelNetwork) Action() error {
	return n.WriteInstruction(BroadcastIdent, Action)
}
