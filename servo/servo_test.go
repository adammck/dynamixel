package servo

import (
	"testing"

	reg "github.com/adammck/dynamixel/registers"
	"github.com/stretchr/testify/assert"
)

var (
	roOneByte     = reg.RegName(1)
	roTwoByte     = reg.RegName(2)
	rwOneByte     = reg.RegName(3)
	rwTwoByte     = reg.RegName(4)
	invalidLength = reg.RegName(5)
	unsupported   = reg.RegName(6)
)

func TestGetRegister(t *testing.T) {
	m := reg.Map{
		roOneByte:     &reg.Register{Address: 0x01, Length: 1, Max: 1},
		roTwoByte:     &reg.Register{Address: 0x02, Length: 2, Max: 1},
		invalidLength: &reg.Register{Address: 0x03, Length: 3, Max: 1},
	}

	p, servo := servo(m, map[int]byte{})

	// read one byte
	p.controlTable[m[roOneByte].Address] = 0x10
	r, err := servo.getRegister(roOneByte)
	assert.Nil(t, err)
	assert.Equal(t, 0x10, r)

	// read two bytes
	p.controlTable[m[roTwoByte].Address] = 0x10
	p.controlTable[m[roTwoByte].Address+1] = 0x20
	r, err = servo.getRegister(roTwoByte)
	assert.Nil(t, err)
	assert.Equal(t, 0x2010, r) // 0x10(L) | 0x20(H)<<8

	// invalid register length
	r, err = servo.getRegister(invalidLength)
	assert.Error(t, err)
	assert.Equal(t, 0, r)

	// unsupported register (not present in control table)
	r, err = servo.getRegister(unsupported)
	assert.Error(t, err)
	assert.Equal(t, 0, r)
}

func TestSetRegister(t *testing.T) {
	m := reg.Map{
		roOneByte: &reg.Register{Address: 0x00, Length: 1, Access: reg.RO, Min: 0, Max: 1},
		rwOneByte: &reg.Register{Address: 0x01, Length: 1, Access: reg.RW, Min: 2, Max: 3},
		rwTwoByte: &reg.Register{Address: 0x02, Length: 2, Access: reg.RW, Min: 0, Max: 2048},
	}

	p, servo := servo(m, map[int]byte{})

	// read-only register can't be set
	err := servo.setRegister(roOneByte, 1)
	assert.Equal(t, byte(0), p.controlTable[0])
	assert.Error(t, err)

	// read/write single byte
	err = servo.setRegister(rwOneByte, 2)
	assert.NoError(t, err)
	assert.Equal(t, byte(2), p.controlTable[1], "control table should have been written")

	// read/write two bytes
	err = servo.setRegister(rwTwoByte, 1025)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), p.controlTable[2], "low byte of control table should have been written")
	assert.Equal(t, byte(0x04), p.controlTable[3], "high byte of control table should have been written")

	// write too-low value with one byte
	err = servo.setRegister(rwOneByte, 1)
	assert.EqualError(t, err, "value too low: 1 (min=2)")
	assert.Equal(t, byte(0x00), p.controlTable[4], "control table should NOT have been written")

	// write too-high value with one byte
	err = servo.setRegister(rwOneByte, 4)
	assert.EqualError(t, err, "value too high: 4 (max=3)")
	assert.Equal(t, byte(0x00), p.controlTable[5], "control table should NOT have been written")
}

func TestSetSetBuffered(t *testing.T) {
	m := reg.Map{
		rwOneByte: &reg.Register{Address: 0x01, Length: 1, Access: reg.RW, Min: 0, Max: 1},
	}

	// Fake servo in buffered mode
	p, servo := servo(m, map[int]byte{})
	servo.SetBuffered(true)

	err := servo.setRegister(rwOneByte, 1)
	assert.NoError(t, err)

	// Ensure that WriteData was not called
	assert.Equal(t, byte(0), p.controlTable[0x01], "control table should not have been written")

	// Ensure that RegWrite was called with the correct params
	if assert.Len(t, p.writeBuf, 1, "write buffer should have 1 element") {
		assert.Equal(t, writeEvent{0x01, []byte{1}}, p.writeBuf[0])
	}
}

func TestVoltage(t *testing.T) {

	// Fake servo which only supports PresentVoltage

	m := reg.Map{
		reg.PresentVoltage: {0x00, 1, reg.RO, 0, 0},
	}

	examples := map[byte]float64{
		95:  9.5,
		105: 10.5,
	}

	for input, exp := range examples {
		_, s := servo(m, map[int]byte{
			0x00: input,
		})

		val, err := s.Voltage()
		assert.NoError(t, err)
		assert.Equal(t, exp, val)
	}
}

// -----------------------------------------------------------------------------

type writeEvent struct {
	addr int
	data []byte
}

// MockNetwork provides a fake servo, with a control table which can be read
// from and written to like a real servo.
type mockProto struct {
	controlTable [50]byte
	writeBuf     []writeEvent
}

// servo returns a real Servo backed by a mock network, where the control table
// initially contains the given bytes. The control table is empty, except that
// the servo ID is 1, and the status return level is 2. (This is just to avoid
// having to specify the same values for every test.)
func servo(r reg.Map, b map[int]byte) (*mockProto, *Servo) {

	// Start with the minimal set of registers, which are required for anything
	// to work. Everything else is optional, so we leave it to the test(s).
	m := reg.Map{
		reg.ServoID:           {40, 1, reg.RW, 0, 252},
		reg.StatusReturnLevel: {41, 1, reg.RW, 0, 2},
	}

	// Add the given registers
	for k, v := range r {
		m[k] = v
	}

	// Pre-configure servo as #1, in verbose mode.
	p := &mockProto{}
	p.controlTable[m[reg.ServoID].Address] = byte(1)
	p.controlTable[m[reg.StatusReturnLevel].Address] = byte(2)

	// Add given control table values
	for addr, val := range b {
		p.controlTable[addr] = val
	}

	s := NewWithReturnLevel(p, m, 1, 2)
	return p, s
}

// Not implemented
func (p *mockProto) Ping(ident int) error {
	return nil
}

func (p *mockProto) ReadData(ident int, addr int, count int) ([]byte, error) {
	return p.controlTable[int(addr) : int(addr)+count], nil
}

func (p *mockProto) WriteData(ident int, address int, data []byte, expectResponse bool) error {
	for i, val := range data {
		p.controlTable[address+i] = val
	}

	return nil
}

func (p *mockProto) RegWrite(ident int, address int, data []byte, expectResponse bool) error {
	p.writeBuf = append(p.writeBuf, writeEvent{address, data})
	return nil
}

func (p *mockProto) Action() error {
	for _, ev := range p.writeBuf {
		p.WriteData(1, ev.addr, ev.data, false)
	}

	p.writeBuf = nil
	return nil
}

// Not implemented
func (p *mockProto) Log(string, ...interface{}) {
}
