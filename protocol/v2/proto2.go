package v2

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/adammck/dynamixel/network"
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
	BroadcastIdent int = 0xFE // 254
)

type Proto2 struct {
	Network  io.ReadWriter
	buffered bool
}

func New(network io.ReadWriter) *Proto2 {
	return &Proto2{
		Network:  network,
		buffered: false,
	}
}

// SetBuffered puts the network in bufferred write mode, which means that the
// REG_WRITE instruction will be used, rather than WRITE_DATA. This causes calls
// to WriteData to be bufferred until the Action method is called, at which time
// they'll all be executed at once.
//
// This is very useful for synchronizing the movements of multiple servos.
func (p *Proto2) SetBuffered(buffered bool) {
	p.buffered = buffered
}

// See:
// http://support.robotis.com/en/product/dynamixel_pro/communication/instruction_status_packet.htm
func (p *Proto2) writeInstruction(ident int, instruction byte, params []byte) error {
	buf := new(bytes.Buffer)
	id := byte(ident & 0xFF)
	pLen := len(params) + 3

	// +------+------+------+----------+----+-------+-------+-------------+--------+-----+--------+-------+-------+
	// | 0xFF | 0xFF | 0xFD |   0x00   | ID | LEN_L | LEN_H |    INST     | Param1 | ... | ParamN | CRL_L | CRL_H |
	// +------+------+------+----------+----+-------+-------+-------------+--------+-----+--------+-------+-------+
	// |       Header       | Reserved | ID | Packet Length | Instruction |       Parameter       |   16bit CRC   |
	// +--------------------+----------+----+---------------+-------------+-----------------------+---------------+

	buf.Write([]byte{
		0xFF,                     // Header
		0xFF,                     // Header
		0xFD,                     // Header
		0x00,                     // Reserved
		id,                       // target ID
		byte(pLen & 0xFF),        // LSB: len(params) + 3
		byte((pLen >> 8) & 0xFF), // MSB: len(params) + 3
		instruction,              // instruction type (see const section)
	})

	// append n params
	buf.Write(params)

	// calculate checksum
	// TODO: Return two bytes from CRC rather than uint16?
	b := CRC(buf.Bytes())
	buf.WriteByte(byte(b & 0xFF))        // LSB
	buf.WriteByte(byte((b >> 8) & 0xFF)) // MSB

	// write to port
	_, err := buf.WriteTo(p.Network)
	if err != nil {
		return err
	}

	return nil
}

func (p *Proto2) readStatusPacket(expID int) ([]byte, error) {

	// +------+------+------+----------+----+-------+-------+-------------+-------+-------+-----+-------+-------+-------+
	// | 0xFF | 0xFF | 0xFD |   0x00   | ID | LEN_L | LEN_H |    0x55     | Error |Param1 | ... |ParamN | CRL_L | CRL_H |
	// +------+------+------+----------+----+-------+-------+-------------+-------+-------+-----+-------+-------+-------+
	// |       Header       | Reserved | ID | Packet Length | Instruction | Error |      Parameter      |   16bit CRC   |
	// +--------------------+----------+----+---------------+-------------+-------+---------------------+---------------+

	// Read the first nine bytes (up to Error), which should always be present.

	buf := make([]byte, 9)
	n, err := p.Network.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("reading packet header: %s", err)
	}
	if n != 9 {
		return nil, fmt.Errorf("reading packet header: expected %d bytes, got %d", 9, n)
	}

	// Check that this is a valid-looking packet, and that it's a status
	// response. If either of these are not met, we return early, even though
	// there might be trash left in the buffer, because we have no idea what's
	// going on. The bus probably needs to be flushed.
	//
	// Note that we don't check the fourth byte, which is reserved, but not part
	// of the spec. It's probably (?) zero, but might change in future.

	if buf[0] != 0xFF || buf[1] != 0xFF || buf[2] != 0xFD {
		return nil, fmt.Errorf("bad status packet header: 0x%02X 0x%02X 0x%02X", buf[0], buf[1], buf[2])
	}

	if buf[7] != Status {
		return nil, fmt.Errorf("bad status packet instruction: 0x%02X", buf[7])
	}

	actID := int(buf[4])
	plen := (int(buf[5]) | int(buf[6])<<8) - 4
	errByte := buf[8]

	// Now read the params, if there are any. We must do this before checking
	// for errors, to avoid leaving junk in the buffer.

	var pbuf []byte
	if plen > 0 {
		pbuf = make([]byte, plen)
		_, err = p.Network.Read(pbuf)
		if err != nil {
			return nil, fmt.Errorf("reading %d params: %s", plen, err)
		}
	}

	// Read the checksum, which is always two bytes.
	// TODO: Read this at the same time as the params.
	// TODO: Check it!

	buf = make([]byte, 2)
	n, err = p.Network.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("reading checksum: %s", err)
	}
	if n != 2 {
		return nil, fmt.Errorf("reading checksum: expected %d bytes, got %d", 2, n)
	}

	// Return an error if the packet contained one.

	if errByte != 0 {
		return nil, decodeError(errByte)
	}

	// Return an error if we received a packet with the wrong ID. This indicates
	// a concurrency issue (maybe clashing IDs on a single bus).

	if actID != expID {
		return nil, fmt.Errorf("expected status packet for %v, but got %v", expID, actID)
	}

	return pbuf, nil
}

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (p *Proto2) Ping(ident int) error {

	// HACK: Ping responses can take forever on XL-320s, but we don't want to
	//       raise the timeout for everything.
	nw, ok := p.Network.(*network.Network)
	if ok {
		ot := nw.Timeout
		nw.Timeout = 2 * time.Second
		defer func() {
			nw.Timeout = ot
		}()
	}

	err := p.writeInstruction(ident, Ping, nil)
	if err != nil {
		return err
	}

	// There's no way to disable the status packet for PING commands, so always
	// wait for it. That's how we know that the servo is responding.
	_, err = p.readStatusPacket(ident)
	if err != nil {
		return err
	}

	return nil
}

// ReadData reads a slice of n bytes from the control table of the given servo
// ID. Use the bytesToInt function to convert the output to something more
// useful.
func (p *Proto2) ReadData(ident int, addr int, n int) ([]byte, error) {
	params := []byte{
		byte(addr & 0xFF),        // LSB
		byte((addr >> 8) & 0xFF), // MSB
		byte(n & 0xFF),           // LSB
		byte((n >> 8) & 0xFF),    // MSB
	}

	err := p.writeInstruction(ident, ReadData, params)
	if err != nil {
		return []byte{}, err
	}

	buf, err := p.readStatusPacket(ident)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (p *Proto2) WriteData(ident int, addr int, data []byte, expectResponse bool) error {
	var instruction byte
	if p.buffered {
		instruction = RegWrite
	} else {
		instruction = WriteData
	}

	ps := make([]byte, len(data)+2)
	ps[0] = byte(addr & 0xFF)        // LSB
	ps[1] = byte((addr >> 8) & 0xFF) // MSB
	copy(ps[2:], data)

	err := p.writeInstruction(ident, instruction, ps)
	if err != nil {
		return err
	}

	if expectResponse {
		_, err = p.readStatusPacket(ident)
		if err != nil {
			return err
		}
	}

	return nil
}

// Action broadcasts the ACTION instruction, which initiates any previously
// bufferred instructions. Doesn't wait for a status packet in response, because
// they are not sent in response to broadcast instructions.
func (p *Proto2) Action() error {
	return p.writeInstruction(BroadcastIdent, Action, nil)
}
