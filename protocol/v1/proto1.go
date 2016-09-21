package v1

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

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
	Serial io.ReadWriteCloser

	// The time to wait for a single read to complete before giving up.
	Timeout time.Duration

	// Optional Logger (which only implements Printf) to log network traffic. If
	// nil (the default), nothing is logged.
	Logger iface.Logger

	// Whether the network is currently in bufferred write mode.
	buffered bool
}

func New(serial io.ReadWriteCloser) *Proto1 {
	return &Proto1{
		Serial:   serial,
		Timeout:  128 * time.Millisecond,
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
func (n *Proto1) SetBuffered(buffered bool) {
	n.buffered = buffered
}

func (n *Proto1) SetLogger(logger iface.Logger) {
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

// This stuff is generic to all Dynamixels. See:
//
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_instruction.htm

func (n *Proto1) WriteInstruction(ident byte, instruction byte, params ...byte) error {
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

	n.Logf(">> %#v\n", buf.Bytes())
	_, err := buf.WriteTo(n.Serial)

	if err != nil {
		return err
	}

	return nil
}

// read receives the next n bytes from the network, blocking if they're not
// immediately available. Returns a slice containing the bytes read. If the
// network timeout is reached, returns the bytes read so far (which might be
// none) and an error.
func (n *Proto1) read(count int) ([]byte, error) {
	start := time.Now()
	buf := make([]byte, count)
	retry := 1 * time.Millisecond
	m := 0

	for m < count {
		nn, err := n.Serial.Read(buf[m:])
		m += nn

		// It's okay if we reached the end of the available bytes. They're
		// probably just not available yet. Other errors are fatal.
		if err != nil && err != io.EOF {
			return buf, err
		}

		// If the timeout has been exceeded, abort.
		if time.Since(start) >= n.Timeout {
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

func (n *Proto1) ReadStatusPacket(expectIdent byte) ([]byte, error) {

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

	headerBuf, headerErr := n.read(3)
	n.Logf("<< %#v (header, ident)\n", headerBuf)
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

		identBuf, identErr := n.read(1)
		n.Logf("<< %#v (ident retry)\n", identBuf)
		if identErr != nil {
			return []byte{}, identErr
		}

		resIdent = identBuf[0]
	}

	// The next two bytes are always present, so just read them.

	paramCountAndErrBitsBuf, pcebErr := n.read(2)
	n.Logf("<< %#v (p+2, errbits)\n", paramCountAndErrBitsBuf)
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
		paramsBuf, paramsErr = n.read(int(numParams))
		n.Logf("<< %#v (params)\n", paramsBuf)
		if paramsErr != nil {
			return []byte{}, paramsErr
		}
	}

	// read the checksum, which is always one byte
	// TODO: check the checksum

	checksumBuf, checksumErr := n.read(1)

	n.Logf("<< %#v (checksum)\n", checksumBuf)
	if checksumErr != nil {
		return []byte{}, checksumErr
	}

	// return an error if the packet contained one.

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

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (n *Proto1) Ping(ident byte) error {
	n.Logf("Ping(%d)", ident)

	writeErr := n.WriteInstruction(ident, Ping)
	if writeErr != nil {
		return writeErr
	}

	// There's no way to disable the status packet for PING commands, so always
	// wait for it. That's how we know that the servo is responding.
	_, readErr := n.ReadStatusPacket(ident)
	if readErr != nil {
		return readErr
	}

	return nil
}

// ReadData reads a slice of count bytes from the control table of the given
// servo ID. Use the bytesToInt function to convert the output to something more
// useful.
func (n *Proto1) ReadData(ident int, addr int, count int) ([]byte, error) {
	ib := utils.Low(ident)

	params := []byte{
		utils.Low(addr),
		byte(count),
	}

	err := n.WriteInstruction(ib, ReadData, params...)
	if err != nil {
		return []byte{}, err
	}

	buf, err := n.ReadStatusPacket(ib)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (n *Proto1) WriteData(ident int, address int, data []byte, expectStausPacket bool) error {
	ib := utils.Low(ident)

	var instruction byte

	if n.buffered {
		instruction = RegWrite
	} else {
		instruction = WriteData
	}

	// Params is dest address followed by the data.
	p := make([]byte, len(data)+1)
	p[0] = utils.Low(address)
	copy(p[1:], data)

	writeErr := n.WriteInstruction(ib, instruction, p...)
	if writeErr != nil {
		return writeErr
	}

	if expectStausPacket {
		_, readErr := n.ReadStatusPacket(ib)
		if readErr != nil {
			return readErr
		}
	}

	return nil
}

// Action broadcasts the ACTION instruction, which initiates any previously
// bufferred instructions. Doesn't wait for a status packet in response, because
// they are not sent in response to broadcast instructions.
func (n *Proto1) Action() error {
	return n.WriteInstruction(BroadcastIdent, Action)
}

// Logf writes a message to the network logger, unless it's nil.
func (n *Proto1) Logf(format string, v ...interface{}) {
	if n.Logger != nil {
		n.Logger.Printf(format, v)
	}
}
