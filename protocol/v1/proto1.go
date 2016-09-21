package v1

import (
	"bytes"
	"fmt"

	"github.com/adammck/dynamixel/iface"
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
	BroadcastIdent byte = 0xFE // 254
)

type Proto1 struct {
	Network iface.Networker

	// Whether the network is currently in bufferred write mode.
	buffered bool
}

func New(network iface.Networker) *Proto1 {
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

func (p *Proto1) writeInstruction(ident byte, instruction byte, params ...byte) error {
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

	sum := byte(ident) + paramsLength + instruction

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

func (p *Proto1) readStatusPacket(expectIdent byte) ([]byte, error) {

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

	headerBuf, headerErr := p.Network.Read(3)
	if headerErr != nil {
		return []byte{}, headerErr
	}

	if headerBuf[0] != 0xFF || headerBuf[1] != 0xFF {
		return []byte{}, fmt.Errorf("bad status packet header: %x %x", headerBuf[0], headerBuf[1])
	}

	// The third byte should be the ident. But if an extra header byte has shown
	// up, ignore it and read another byte to replace it.

	resIdent := headerBuf[2]
	if resIdent == 255 {

		identBuf, identErr := p.Network.Read(1)
		if identErr != nil {
			return []byte{}, identErr
		}

		resIdent = identBuf[0]
	}

	// The next two bytes are always present, so just read them.

	paramCountAndErrBitsBuf, pcebErr := p.Network.Read(2)
	if pcebErr != nil {
		return []byte{}, pcebErr
	}

	numParams := uint8(paramCountAndErrBitsBuf[0]) - 2
	errBits := paramCountAndErrBitsBuf[1]
	paramsBuf := make([]byte, numParams)

	// now read the params, if there are any. we must do this before checking
	// for errors, to avoid leaving junk in the buffer.

	if numParams > 0 {
		var paramsErr error
		paramsBuf, paramsErr = p.Network.Read(int(numParams))
		if paramsErr != nil {
			return []byte{}, paramsErr
		}
	}

	// read the checksum, which is always one byte
	// TODO: check the checksum

	_, checksumErr := p.Network.Read(1)
	if checksumErr != nil {
		return []byte{}, checksumErr
	}

	// return an error if the packet contained one.

	if errBits != 0x0 {
		return []byte{}, decodeError(errBits)
	}

	// return an error if we received a packet with the wrong ID. this indicates
	// a concurrency issue (maybe clashing IDs on a single bus).

	if resIdent != expectIdent {
		return []byte{}, fmt.Errorf("expected status packet for %v, but got %v", expectIdent, resIdent)
	}

	// omg, nothing went wrong

	return paramsBuf, nil
}

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (p *Proto1) Ping(ident int) error {
	ib := utils.Low(ident)

	writeErr := p.writeInstruction(ib, Ping)
	if writeErr != nil {
		return writeErr
	}

	// There's no way to disable the status packet for PING commands, so always
	// wait for it. That's how we know that the servo is responding.
	_, readErr := p.readStatusPacket(ib)
	if readErr != nil {
		return readErr
	}

	return nil
}

// ReadData reads a slice of count bytes from the control table of the given
// servo ID. Use the bytesToInt function to convert the output to something more
// useful.
func (p *Proto1) ReadData(ident int, addr int, count int) ([]byte, error) {
	ib := utils.Low(ident)

	params := []byte{
		utils.Low(addr),
		byte(count),
	}

	err := p.writeInstruction(ib, ReadData, params...)
	if err != nil {
		return []byte{}, err
	}

	buf, err := p.readStatusPacket(ib)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (p *Proto1) WriteData(ident int, address int, data []byte, expectResponse bool) error {
	ib := utils.Low(ident)

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

	writeErr := p.writeInstruction(ib, instruction, ps...)
	if writeErr != nil {
		return writeErr
	}

	if expectResponse {
		_, readErr := p.readStatusPacket(ib)
		if readErr != nil {
			return readErr
		}
	}

	return nil
}

// Action broadcasts the ACTION instruction, which initiates any previously
// bufferred instructions. Doesn't wait for a status packet in response, because
// they are not sent in response to broadcast instructions.
func (p *Proto1) Action() error {
	return p.writeInstruction(BroadcastIdent, Action)
}
