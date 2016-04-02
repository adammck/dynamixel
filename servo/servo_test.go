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

	n, servo := servo(m, map[int]byte{})

	// read one byte
	n.controlTable[m[roOneByte].Address] = 0x10
	r, err := servo.getRegister(roOneByte)
	assert.Nil(t, err)
	assert.Equal(t, 0x10, r)

	// read two bytes
	n.controlTable[m[roTwoByte].Address] = 0x10
	n.controlTable[m[roTwoByte].Address+1] = 0x20
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

	n, servo := servo(m, map[int]byte{})

	// read-only register can't be set
	err := servo.setRegister(roOneByte, 1)
	assert.Equal(t, byte(0), n.controlTable[0])
	assert.Error(t, err)

	// read/write single byte
	err = servo.setRegister(rwOneByte, 2)
	assert.NoError(t, err)
	assert.Equal(t, byte(2), n.controlTable[1], "control table should have been written")

	// read/write two bytes
	err = servo.setRegister(rwTwoByte, 1025)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), n.controlTable[2], "low byte of control table should have been written")
	assert.Equal(t, byte(0x04), n.controlTable[3], "high byte of control table should have been written")

	// write too-low value with one byte
	err = servo.setRegister(rwOneByte, 1)
	assert.EqualError(t, err, "value too low: 1 (min=2)")
	assert.Equal(t, byte(0x00), n.controlTable[4], "control table should NOT have been written")

	// write too-high value with one byte
	err = servo.setRegister(rwOneByte, 4)
	assert.EqualError(t, err, "value too high: 4 (max=3)")
	assert.Equal(t, byte(0x00), n.controlTable[5], "control table should NOT have been written")
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

// MockNetwork provides a fake servo, with a control table which can be read
// from and written to like a real servo.
type mockNetwork struct {
	controlTable [50]byte
}

// servo returns a real Servo backed by a mock network, where the control table
// initially contains the given bytes. The control table is empty, except that
// the servo ID is 1, and the status return level is 2. (This is just to avoid
// having to specify the same values for every test.)
func servo(r reg.Map, b map[int]byte) (*mockNetwork, *Servo) {

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
	n := &mockNetwork{}
	n.controlTable[m[reg.ServoID].Address] = byte(1)
	n.controlTable[m[reg.StatusReturnLevel].Address] = byte(2)

	// Add given control table values
	for addr, val := range b {
		n.controlTable[addr] = val
	}

	s := New(n, 1, m)
	return n, s
}

func (n *mockNetwork) Ping(ident uint8) error {
	return nil
}

func (n *mockNetwork) ReadData(ident uint8, addr byte, count int) ([]byte, error) {
	return n.controlTable[int(addr) : int(addr)+count], nil
}

func (n *mockNetwork) WriteData(ident uint8, expectStausPacket bool, params ...byte) error {
	addr := int(params[0])

	for i, val := range params[1:] {
		n.controlTable[addr+i] = val
	}

	return nil
}

func (n *mockNetwork) Log(string, ...interface{}) {
}
