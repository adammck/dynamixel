package v2

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/adammck/dynamixel/iface"
	"github.com/adammck/dynamixel/utils"
)

const (

	// Instruction Types
	Ping         byte = 0x01
	ReadData     byte = 0x02
	WriteData    byte = 0x03
	RegWrite     byte = 0x04
	Action       byte = 0x05
	FactoryReset byte = 0x06 // All data to factory default settings
	Reboot       byte = 0x08 // Reboot device
	Status       byte = 0x55
	SyncRead     byte = 0x82
	SyncWrite    byte = 0x83
	BulkRead     byte = 0x92
	BulkWrite    byte = 0x93

	// Send an instruction to all servos
	BroadcastIdent byte = 0xFE // 254
)

type Proto2 struct {
	Serial io.ReadWriteCloser

	// The time to wait for a single read to complete before giving up.
	Timeout time.Duration

	// Optional Logger (which only implements Printf) to log network traffic. If
	// nil (the default), nothing is logged.
	Logger iface.Logger

	// Whether the network is currently in bufferred write mode.
	buffered bool
}

func New(serial io.ReadWriteCloser) *Proto2 {
	return &Proto2{
		Serial:   serial,
		Timeout:  10 * time.Millisecond,
		Logger:   nil,
		buffered: false,
	}
}

// SetBuffered puts the network in bufferred write mode, which means that the
// REG_WRITE instruction will be used, rather than WRITE_DATA. This causes calls
// to WriteData to be bufferred until the Action method is called, at which time
// they'll all be executed at once.
//
// This is very useful for synchronizing the movements of multiple servos.
func (n *Proto2) SetBuffered(buffered bool) {
	n.buffered = buffered
}

func (n *Proto2) SetLogger(logger iface.Logger) {
	n.Logger = logger
}

// DecodeStartusError Converts an error byte (as included in a status packet)
// into an error object with a friendly error message. We can't be too specific
// about it, because any combination of errors might occur at the same time.
//
// See: http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm#Status_Packet
func DecodeStatusError(errBits byte) error {
	str := []string{}

	if errBits&1 == 1 {
		str = append(str, "input voltage")
	}

	if errBits&2 == 2 {
		str = append(str, "angle limit")
	}

	if errBits&4 == 4 {
		str = append(str, "overheating")
	}

	if errBits&8 == 8 {
		str = append(str, "range")
	}

	if errBits&16 == 16 {
		str = append(str, "checksum")
	}

	if errBits&32 == 32 {
		str = append(str, "overload")
	}

	if errBits&64 == 64 {
		str = append(str, "instruction")
	}

	if errBits&128 == 128 {
		str = append(str, "unknown")
	}

	return fmt.Errorf("status error(s): %s", strings.Join(str, ", "))
}

func DecodePacketError(err byte) error {
	s := ""

	switch err {
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
		s = fmt.Sprintf("unknown error: 0x%X", err)
	}

	return errors.New(s)
}

// See:
// http://support.robotis.com/en/product/dynamixel_pro/communication/instruction_status_packet.htm
func (network *Proto2) WriteInstruction(ident uint8, instruction byte, params ...byte) error {
	buf := new(bytes.Buffer)
	paramsLength := byte(len(params) + 3)

	// +------+------+------+----------+----+-------+-------+-------------+--------+-----+--------+-------+-------+
	// | 0xFF | 0xFF | 0xFD |   0x00   | ID | LEN_L | LEN_H |    INST     | Param1 | ... | ParamN | CRL_L | CRL_H |
	// +------+------+------+----------+----+-------+-------+-------------+--------+-----+--------+-------+-------+
	// |       Header       | Reserved | ID | Packet Length | Instruction |       Parameter       |   16bit CRC   |
	// +--------------------+----------+----+---------------+-------------+-----------------------+---------------+

	buf.Write([]byte{
		0xFF,                       // Header
		0xFF,                       // Header
		0xFD,                       // Header
		0x00,                       // Reserved
		byte(ident),                // target ID
		paramsLength & 0xFF,        // LSB: len(params) + 3
		(paramsLength >> 8) & 0xFF, // MSB: len(params) + 3
		instruction,                // instruction type (see const section)
	})

	// append n params
	buf.Write(params)

	// calculate checksum
	// TODO: Return two bytes from CRC rather than uint16?
	b := CRC(buf.Bytes())
	buf.WriteByte(byte(b & 0xFF))        // LSB
	buf.WriteByte(byte((b >> 8) & 0xFF)) // MSB

	// write to port
	network.Logf(">> %#v\n", buf.Bytes())
	_, err := buf.WriteTo(network.Serial)
	if err != nil {
		return err
	}

	return nil
}

// read receives the next n bytes from the network, blocking if they're not
// immediately available. Returns a slice containing the bytes read. If the
// network timeout is reached, returns the bytes read so far (which might be
// none) and an error.
func (network *Proto2) read(n int) ([]byte, error) {
	start := time.Now()
	buf := make([]byte, n)
	retry := 1 * time.Millisecond
	m := 0

	for m < n {
		nn, err := network.Serial.Read(buf[m:])
		m += nn

		network.Logf("~~ n=%d, m=%d, nn=%d, err=%s\n", n, m, nn, err)

		// It's okay if we reached the end of the available bytes. They're
		// probably just not available yet. Other errors are fatal.
		if err != nil && err != io.EOF {
			return buf, err
		}

		// If the timeout has been exceeded, abort.
		if time.Since(start) >= network.Timeout {
			return buf, fmt.Errorf("read timed out")
		}

		// If no bytes were read, back off exponentially. This is just to avoid
		// flooding the network with retries if a servo isn't responding.
		if nn == 0 {
			time.Sleep(retry)
			retry *= 2
		}
	}

	return buf, nil
}

func (network *Proto2) ReadStatusPacket(expectIdent uint8) ([]byte, error) {

	// +------+------+------+----------+----+-------+-------+-------------+-------+-------+-----+-------+-------+-------+
	// | 0xFF | 0xFF | 0xFD |   0x00   | ID | LEN_L | LEN_H |    0x55     | Error |Param1 | ... |ParamN | CRL_L | CRL_H |
	// +------+------+------+----------+----+-------+-------+-------------+-------+-------+-----+-------+-------+-------+
	// |       Header       | Reserved | ID | Packet Length | Instruction | Error |      Parameter      |   16bit CRC   |
	// +--------------------+----------+----+---------------+-------------+-------+---------------------+---------------+

	buf, err := network.read(9)
	network.Logf("<< %#v\n", buf)
	if err != nil {
		return nil, err
	}

	// Check that this is a valid-looking packet, and that it's a status
	// response. If either of these are not met, we return early, even though
	// there might be trash left in the buffer, because we have no idea what's
	// going on. The bus probably needs to be flushed.
	//
	// Note that we don't check the fourth byte, which is reserved, but not part
	// of the spec. It's probably (?) zero, but might change in future.

	if buf[0] != 0xFF || buf[1] != 0xFF || buf[2] != 0xFD {
		return nil, fmt.Errorf("bad status packet header: %x", buf[0:2])
	}

	if buf[7] != Status {
		return nil, fmt.Errorf("bad status packet instruction: %x", buf[7])
	}

	resIdent := uint8(buf[4])
	numParams := (int(buf[5]) | int(buf[6])<<8) - 4
	errBits := buf[8]

	// Now read the params, if there are any. We must do this before checking
	// for errors, to avoid leaving junk in the buffer.

	pbuf := make([]byte, numParams)
	if numParams > 0 {
		pbuf, err = network.read(numParams)
		network.Logf("<< %#v (params)\n", pbuf)
		if err != nil {
			return nil, err
		}
	}

	// Read the checksum, which is always two bytes.
	// TODO: Read this at the same time as the params.
	// TODO: Check it!

	buf, err = network.read(2)
	network.Logf("<< %#v (checksum)\n", buf)
	if err != nil {
		return nil, err
	}

	// Return an error if the packet contained one.

	if errBits != 0 {
		network.Logf("EE %v\n", errBits)
		return nil, DecodePacketError(errBits)
	}

	// Return an error if we received a packet with the wrong ID. This indicates
	// a concurrency issue (maybe clashing IDs on a single bus).

	if resIdent != expectIdent {
		return nil, fmt.Errorf("expected status packet for %v, but got %v", expectIdent, resIdent)
	}

	return pbuf, nil
}

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (n *Proto2) Ping(ident uint8) error {

	// HACK: Ping responses can take forever on XL-320s, but we don't want to raise the timeout for everything.
	ot := n.Timeout
	n.Timeout = 2 * time.Second
	defer func() {
		n.Timeout = ot
	}()

	n.Logf("-- Ping %d\n", ident)
	t := time.Now()

	err := n.WriteInstruction(ident, Ping)
	if err != nil {
		return err
	}

	// There's no way to disable the status packet for PING commands, so always
	// wait for it. That's how we know that the servo is responding.
	_, err = n.ReadStatusPacket(ident)
	n.Logf("++ %s\n", time.Since(t))
	if err != nil {
		return err
	}

	return nil
}

// ReadData reads a slice of n bytes from the control table of the given servo
// ID. Use the bytesToInt function to convert the output to something more
// useful.
func (network *Proto2) ReadData(ident int, addr int, n int) ([]byte, error) {
	network.Logf("-- ReadData %d, 0x%x, %d\n", ident, addr, n)
	ib := utils.Low(ident)
	t := time.Now()

	params := []byte{
		byte(addr & 0xFF),        // LSB
		byte((addr >> 8) & 0xFF), // MSB
		byte(n & 0xFF),           // LSB
		byte((n >> 8) & 0xFF),    // MSB
	}

	err := network.WriteInstruction(ib, ReadData, params...)
	if err != nil {
		return []byte{}, err
	}

	buf, err := network.ReadStatusPacket(ib)
	network.Logf("++ %s\n", time.Since(t))
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (n *Proto2) WriteData(ident int, addr int, params []byte, expectStausPacket bool) error {
	n.Logf("-- WriteData: ident=%d, expectStausPacket=%t, addr=0x%x, params=%v\n", ident, expectStausPacket, addr, params)
	ib := utils.Low(ident)

	var instruction byte
	t := time.Now()

	if n.buffered {
		instruction = RegWrite
	} else {
		instruction = WriteData
	}

	p := make([]byte, len(params)+2)
	p[0] = byte(addr & 0xFF)        // LSB
	p[1] = byte((addr >> 8) & 0xFF) // MSB
	copy(p[2:], params)

	writeErr := n.WriteInstruction(ib, instruction, p...)
	if writeErr != nil {
		return writeErr
	}

	if expectStausPacket {
		_, err := n.ReadStatusPacket(ib)
		n.Logf("++ %s\n", time.Since(t))
		if err != nil {
			return err
		}
	} else {
		n.Logf("++ %s\n", time.Since(t))
	}

	return nil
}

// Action broadcasts the ACTION instruction, which initiates any previously
// bufferred instructions. Doesn't wait for a status packet in response, because
// they are not sent in response to broadcast instructions.
func (n *Proto2) Action() error {
	return n.WriteInstruction(BroadcastIdent, Action)
}

func (nw *Proto2) Flush() {
	buf := make([]byte, 128)
	var n int

	for {
		n, _ = nw.Serial.Read(buf)
		nw.Logf(".. %v\n", buf)
		if n == 0 {
			break
		}
	}
}

func (n *Proto2) FactoryReset() error {
	panic("not implemented: Network.FactoryReset")
}

func (n *Proto2) Reboot() error {
	panic("not implemented: Network.Reboot")
}

func (n *Proto2) SyncRead() error {
	panic("not implemented: Network.SyncRead")
}

func (n *Proto2) SyncWrite() error {
	panic("not implemented: Network.SyncWrite")
}

func (n *Proto2) BulkRead() error {
	panic("not implemented: Network.BulkRead")
}

func (n *Proto2) BulkWrite() error {
	panic("not implemented: Network.BulkWrite")
}

// Logf writes a message to the network logger, unless it's nil.
func (n *Proto2) Logf(format string, v ...interface{}) {
	if n.Logger != nil {
		n.Logger.Printf(format, v...)
	}
}
