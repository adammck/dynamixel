package dynamixel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockNetwork struct {
	controlTable [50]byte
}

func (n *mockNetwork) Ping(ident uint8) error {
	return nil
}

func (n *mockNetwork) ReadData(ident uint8, addr byte, count int) ([]byte, error) {
	return n.controlTable[int(addr) : int(addr)+count], nil
}

// TODO: Move this into Servo?
func (n *mockNetwork) ReadInt(ident uint8, addr byte, count int) (int, error) {
	return 0, nil
}

func (n *mockNetwork) WriteData(ident uint8, expectStausPacket bool, params ...byte) error {
	addr := int(params[0])

	for i, val := range params[1:] {
		n.controlTable[addr+i] = val
	}

	return nil
}

func (n *mockNetwork) Log(string, ...interface{}) {}

func TestUpdateCache(t *testing.T) {
	n := &mockNetwork{}
	n.controlTable = [50]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
	}

	servo := NewServo(n, 1)
	err := servo.updateCache()

	assert.Nil(t, err)
	assert.Equal(t, servo.cache, n.controlTable)
}

func TestGetRegister(t *testing.T) {
	n := &mockNetwork{}
	servo := NewServo(n, 1)
	servo.cache = [50]byte{
		0x99, 0x10, 0x20, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// invalid register length
	x, err := servo.getRegister(Register{0x00, 3, ro, true})
	assert.Error(t, err)
	assert.Equal(t, 0, x)

	// one byte (cached)
	a, err := servo.getRegister(Register{0x00, 1, ro, true})
	assert.Nil(t, err)
	assert.Equal(t, 0x99, a)

	// two bytes (cached)
	b, err := servo.getRegister(Register{0x01, 2, ro, true})
	assert.Nil(t, err)
	assert.Equal(t, 0x2010, b) // 0x10(L) | 0x20(H)<<8

	// one byte (immediate)
	servo.cache[0x02] = 0x77
	n.controlTable[0x02] = 0x88
	c, err := servo.getRegister(Register{0x02, 1, ro, false})
	assert.Nil(t, err)
	assert.Equal(t, 0x88, c)
	assert.Equal(t, 0x88, servo.cache[0x02], "servo cache should have been updated")
}

func TestSetRegister(t *testing.T) {
	n := &mockNetwork{}
	servo := NewServo(n, 1)

	// read only register can't be set
	err := servo.setRegister(Register{0x00, 1, ro, true}, 1)
	assert.Equal(t, 0x00, n.controlTable[0])
	assert.Equal(t, 0x00, servo.cache[0])
	assert.Error(t, err)

	// read/write single byte
	err = servo.setRegister(Register{0x01, 1, rw, true}, 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, n.controlTable[1], "control table should have been written")
	assert.Equal(t, 99, servo.cache[1], "servo cache should have been updated")

	// read/write two bytes
	err = servo.setRegister(Register{0x02, 2, rw, true}, 4097)
	assert.NoError(t, err)
	assert.Equal(t, 0x01, n.controlTable[2], "low byte of control table should have been written")
	assert.Equal(t, 0x10, n.controlTable[3], "high byte of control table should have been written")
	assert.Equal(t, 0x01, servo.cache[2], "low byte of servo cache should have been updated")
	assert.Equal(t, 0x10, servo.cache[3], "high byte of servo cache should have been updated")
}

func TestModelNumber(t *testing.T) {
	n := &mockNetwork{}
	n.controlTable[0] = 0x02
	n.controlTable[1] = 0x01
	servo := NewServo(n, 1)
	servo.updateCache()
	val, err := servo.ModelNumber()
	assert.NoError(t, err)
	assert.Equal(t, 258, val)
}
