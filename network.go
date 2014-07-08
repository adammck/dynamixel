package dynamixel

import (
	"bytes"
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
)

type DynamixelNetwork struct {
	Serial io.ReadWriteCloser
}

func NewNetwork(serial io.ReadWriteCloser) *DynamixelNetwork {
	return &DynamixelNetwork{
		Serial: serial,
	}
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
		fmt.Printf("bad header! buf: %#v\n", buf)
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
		fmt.Printf("err code! buf: %#v\n", errBits)
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
	writeErr := n.WriteInstruction(ident, WriteData, params...)
	if writeErr != nil {
		return writeErr
	}

	_, readErr := n.ReadStatusPacket(ident)
	if readErr != nil {
		return readErr
	}

	return nil
}
