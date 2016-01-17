package dynamixel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCacheIsPopulated(t *testing.T) {

	// create the mock network first, which contains the (mock) remote control
	// table. this should be read into the cache when the servo is allocated.

	n := &mockNetwork{
		[50]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
			0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13,
			0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D,
			0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30, 0x31,
		},
	}

	servo := NewServo(n, 1)
	assert.Equal(t, servo.cache, n.controlTable)
}

func TestGetRegister(t *testing.T) {
	n, servo := servo(map[int]byte{})

	// pre-populate the CACHE, not the control table
	servo.cache[0x00] = byte(0x99)
	servo.cache[0x01] = byte(0x10)
	servo.cache[0x02] = byte(0x20)
	servo.cache[0x03] = byte(0x88)

	// invalid register length
	x, err := servo.getRegister(Register{0x00, 3, ro, true})
	assert.Error(t, err)
	assert.Equal(t, 0, x)

	// one byte (cached)
	a, err := servo.getRegister(Register{0x00, 1, ro, true})
	assert.Nil(t, err)
	assert.Equal(t, 0x99, a)

	// two bytes (cached)
	b, err := servo.getRegister(Register{byte(1), 2, ro, true})
	assert.Nil(t, err)
	assert.Equal(t, 0x2010, b) // 0x10(L) | 0x20(H)<<8

	// one byte (immediate)
	servo.cache[0x02] = 0x77
	n.controlTable[0x02] = 0x88
	c, err := servo.getRegister(Register{0x02, 1, ro, false})
	assert.Nil(t, err)
	assert.Equal(t, 0x88, c)
	assert.Equal(t, byte(0x88), servo.cache[0x02], "servo cache should have been updated")
}

func TestSetRegister(t *testing.T) {
	n, servo := servo(map[int]byte{})

	// read only register can't be set
	err := servo.setRegister(Register{0x00, 1, ro, true}, 1)
	assert.Equal(t, byte(0), n.controlTable[0])
	assert.Equal(t, byte(0), servo.cache[0])
	assert.Error(t, err)

	// read/write single byte
	err = servo.setRegister(Register{byte(1), 1, rw, true}, 99)
	assert.NoError(t, err)
	assert.Equal(t, byte(99), n.controlTable[1], "control table should have been written")
	assert.Equal(t, byte(99), servo.cache[1], "servo cache should have been updated")

	// read/write two bytes
	err = servo.setRegister(Register{0x02, 2, rw, true}, 4097)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), n.controlTable[2], "low byte of control table should have been written")
	assert.Equal(t, byte(0x10), n.controlTable[3], "high byte of control table should have been written")
	assert.Equal(t, byte(0x01), servo.cache[2], "low byte of servo cache should have been updated")
	assert.Equal(t, byte(0x10), servo.cache[3], "high byte of servo cache should have been updated")
}

// -- Registers

func TestModelNumber(t *testing.T) {
	_, s := servo(map[int]byte{
		0x00: byte(2), // L
		0x01: byte(1), // H
	})

	val, err := s.ModelNumber()
	assert.NoError(t, err)
	assert.Equal(t, 258, val)
}

func TestFirmwareVersion(t *testing.T) {
	_, s := servo(map[int]byte{
		0x02: byte(99),
	})

	val, err := s.FirmwareVersion()
	assert.NoError(t, err)
	assert.Equal(t, 99, val)
}

func TestTorqueEnable(t *testing.T) {
	_, s := servo(map[int]byte{
		0x18: 0,
	})

	val, err := s.TorqueEnable()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	s.cache[0x18] = 1
	val, err = s.TorqueEnable()
	assert.NoError(t, err)
	assert.Equal(t, true, val)
}

func TestSetTorqueEnable(t *testing.T) {
	n, s := servo(map[int]byte{})

	err := s.SetTorqueEnable(true)
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n.controlTable[0x18])
	assert.Equal(t, byte(1), s.cache[0x18])

	err = s.SetTorqueEnable(false)
	assert.NoError(t, err)
	assert.Equal(t, byte(0), n.controlTable[0x18])
	assert.Equal(t, byte(0), s.cache[0x18])
}

func TestLED(t *testing.T) {
	n := &mockNetwork{}
	s := NewServo(n, 1)

	s.cache[0x19] = 0
	val, err := s.LED()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	s.cache[0x19] = 1
	val, err = s.LED()
	assert.NoError(t, err)
	assert.Equal(t, true, val)
}

func TestSetLED(t *testing.T) {
	n := &mockNetwork{}
	s := NewServo(n, 1)

	err := s.SetLED(true)
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n.controlTable[0x19])
	assert.Equal(t, byte(1), s.cache[0x19])

	err = s.SetLED(false)
	assert.NoError(t, err)
	assert.Equal(t, byte(0), n.controlTable[0x19])
	assert.Equal(t, byte(0), s.cache[0x19])
}

func TestMovingSpeed(t *testing.T) {
	_, s := servo(map[int]byte{
		0x20: 0xff, // L
		0x21: 0x03, // H
	})

	val, err := s.MovingSpeed()
	assert.NoError(t, err)
	assert.Equal(t, 1023, val)
}

func TestSetMovingSpeed(t *testing.T) {
	n := &mockNetwork{}
	s := NewServo(n, 1)

	err := s.SetMovingSpeed(513)
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n.controlTable[0x20]) // L
	assert.Equal(t, byte(2), n.controlTable[0x21]) // H
	assert.Equal(t, byte(1), s.cache[0x20])        // L
	assert.Equal(t, byte(2), s.cache[0x21])        // H
}

func TestPosition(t *testing.T) {
	_, s := servo(map[int]byte{
		0x24: byte(1), // L
		0x25: 0x00,    // H
	})

	val, err := s.Position()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestPresentSpeed(t *testing.T) {
	_, s := servo(map[int]byte{
		0x26: byte(0x01), // L
		0x27: byte(0x04), // H
	})

	val, err := s.PresentSpeed()
	assert.NoError(t, err)
	assert.Equal(t, 1025, val)
}

// presentLoad
func TestPresentLoad(t *testing.T) {
	_, s := servo(map[int]byte{
		0x28: 5, // L
		0x29: 4, // H
	})

	val, err := s.PresentLoad()
	assert.NoError(t, err)
	assert.Equal(t, 1029, val)
}

func TestPresentVoltage(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2a: 95,
	})

	val, err := s.PresentVoltage()
	assert.NoError(t, err)
	assert.Equal(t, 95, val)
}

// presentTemperature
func TestPresentTemperature(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2b: 0x55,
	})

	val, err := s.PresentTemperature()
	assert.NoError(t, err)
	assert.Equal(t, 85, val)
}

// registered
func TestRegistered(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2c: 1,
	})

	val, err := s.Registered()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

// moving
func TestMoving(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2e: 0,
	})

	val, err := s.Moving()
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
}

// -- High-level interface

func TestVoltage(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2A: 105,
	})

	val, err := s.Voltage()
	assert.NoError(t, err)
	assert.Equal(t, 10.5, val)
}

// -----------------------------------------------------------------------------

// MockNetwork provides a fake servo, with a control table which can be read
// from and written to like a real servo.
type mockNetwork struct {
	controlTable [50]byte
}

// servo returns a real Servo backed by a mock network, where the control table
// initially contains the given bytes.
func servo(b map[int]byte) (*mockNetwork, *DynamixelServo) {
	n := &mockNetwork{}

	for addr, val := range b {
		n.controlTable[addr] = val
	}

	s := NewServo(n, 1)
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
