package v1

import (
	"bytes"
	"fmt"
	"io"

	"github.com/adammck/dynamixel/utils"
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
	BroadcastIdent int = 0xFE // 254
)

type Proto1 struct {
	Network  io.ReadWriter
	buffered bool
}

func New(network io.ReadWriter) *Proto1 {
	return &Proto1{
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
func (p *Proto1) SetBuffered(buffered bool) {
	p.buffered = buffered
}

// This stuff is generic to all Dynamixels. See:
//
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_instruction.htm

func (p *Proto1) writeInstruction(ident int, instruction byte, params []byte) error {
	buf := new(bytes.Buffer)
	id := byte(ident & 0xFF)
	pLen := byte(len(params) + 2)

	// build instruction packet

	buf.Write([]byte{
		0xFF,        // header
		0xFF,        // header
		id,          // target Dynamixel ID
		pLen,        // len(params) + 2
		instruction, // instruction type (read/write/etc)
	})

	buf.Write(params)

	// calculate checksum

	sum := id + pLen + instruction
	for _, value := range params {
		sum += value
	}

	buf.WriteByte(byte((^sum) & 0xFF))

	// write to port
	_, err := buf.WriteTo(p.Network)
	if err != nil {
		return err
	}

	return nil
}

func (p *Proto1) readStatusPacket(expID int) ([]byte, error) {

	//
	// Status packets are similar to instruction packet:
	//
	// +------+------+ ------ +-------+----------+---------+--------+--------+----------+
	// | 0xFF | 0xFF |  0xFF  | ident | params+2 | errBits | param1 | param2 | checksum |
	// +------+------+ ------ +-------+----------+---------+--------+--------+----------+
	//
	//                  ^-- sometimes, but it shouldn't be there?!
	//
	// We don't know the length, because responses can have variable numbers of
	// parameters.
	//

	// Read the first three bytes, which are always present. Hopefully, the
	// first two are the header, and the third is the ID of the servo which this
	// packet refers to. But sometimes, the third byte is another 0xFF. I don't
	// know why, and I can't seem to find any useful information on the matter.

	buf := make([]byte, 3)
	_, err := p.Network.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	if buf[0] != 0xFF || buf[1] != 0xFF {
		return []byte{}, fmt.Errorf("bad status packet header: %x %x", buf[0], buf[1])
	}

	// The third byte should be the ident. But if an extra header byte has shown
	// up, ignore it and read another byte to replace it.

	actID := int(buf[2])
	if actID == 255 {

		buf = make([]byte, 1)
		_, identErr := p.Network.Read(buf)
		if identErr != nil {
			return []byte{}, identErr
		}

		actID = int(buf[0])
	}

	// The next two bytes are always present, so just read them.

	buf = make([]byte, 2)
	_, err = p.Network.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	plen := uint8(buf[0]) - 2
	errBits := buf[1]
	pbuf := make([]byte, plen)

	// now read the params, if there are any. we must do this before checking
	// for errors, to avoid leaving junk in the buffer.

	if plen > 0 {
		pbuf = make([]byte, int(plen))
		_, err = p.Network.Read(pbuf)
		if err != nil {
			return []byte{}, err
		}
	}

	// read the checksum, which is always one byte
	// TODO: check the checksum

	buf = make([]byte, 1)
	_, err = p.Network.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	// return an error if the packet contained one.

	if errBits != 0x0 {
		return []byte{}, decodeError(errBits)
	}

	// return an error if we received a packet with the wrong ID. this indicates
	// a concurrency issue (maybe clashing IDs on a single bus).

	if actID != expID {
		return []byte{}, fmt.Errorf("expected status packet for %v, but got %v", expID, actID)
	}

	// omg, nothing went wrong

	return pbuf, nil
}

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (p *Proto1) Ping(ident int) error {
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

// ReadData reads a slice of count bytes from the control table of the given
// servo ID. Use the bytesToInt function to convert the output to something more
// useful.
func (p *Proto1) ReadData(ident int, addr int, count int) ([]byte, error) {
	params := []byte{
		utils.Low(addr),
		byte(count),
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

func (p *Proto1) WriteData(ident int, address int, data []byte, expectResponse bool) error {
	var instruction byte
	if p.buffered {
		instruction = RegWrite
	} else {
		instruction = WriteData
	}

	// Params is dest address followed by the data.
	ps := make([]byte, len(data)+1)
	ps[0] = utils.Low(address)
	copy(ps[1:], data)

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
func (p *Proto1) Action() error {
	return p.writeInstruction(BroadcastIdent, Action, nil)
}
