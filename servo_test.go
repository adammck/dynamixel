package dynamixel

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	x, err := servo.getRegister(Register{0x00, 3, ro, true, 0, 1})
	assert.Error(t, err)
	assert.Equal(t, 0, x)

	// one byte (cached)
	a, err := servo.getRegister(Register{0x00, 1, ro, true, 0, 1})
	assert.Nil(t, err)
	assert.Equal(t, 0x99, a)

	// two bytes (cached)
	b, err := servo.getRegister(Register{byte(1), 2, ro, true, 0, 1})
	assert.Nil(t, err)
	assert.Equal(t, 0x2010, b) // 0x10(L) | 0x20(H)<<8

	// one byte (immediate)
	servo.cache[0x02] = 0x77
	n.controlTable[0x02] = 0x88
	c, err := servo.getRegister(Register{0x02, 1, ro, false, 0, 1})
	assert.Nil(t, err)
	assert.Equal(t, 0x88, c)
	assert.Equal(t, byte(0x88), servo.cache[0x02], "servo cache should have been updated")
}

func TestSetRegister(t *testing.T) {
	n, servo := servo(map[int]byte{})

	// read only register can't be set
	err := servo.setRegister(Register{0x00, 1, ro, true, 0, 1}, 1)
	assert.Equal(t, byte(0), n.controlTable[0])
	assert.Equal(t, byte(0), servo.cache[0])
	assert.Error(t, err)

	// read/write single byte
	err = servo.setRegister(Register{0x01, 1, rw, true, 0, 2}, 2)
	assert.NoError(t, err)
	assert.Equal(t, byte(2), n.controlTable[1], "control table should have been written")
	assert.Equal(t, byte(2), servo.cache[1], "servo cache should have been updated")

	// read/write two bytes
	err = servo.setRegister(Register{0x02, 2, rw, true, 0, 2048}, 1025)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), n.controlTable[2], "low byte of control table should have been written")
	assert.Equal(t, byte(0x04), n.controlTable[3], "high byte of control table should have been written")
	assert.Equal(t, byte(0x01), servo.cache[2], "low byte of servo cache should have been updated")
	assert.Equal(t, byte(0x04), servo.cache[3], "high byte of servo cache should have been updated")

	// write too-low value with one byte
	err = servo.setRegister(Register{0x04, 1, rw, true, 2, 3}, 1)
	assert.EqualError(t, err, "value too low: 1 (min=2)")
	assert.Equal(t, byte(0x00), n.controlTable[4], "control table should NOT have been written")
	assert.Equal(t, byte(0x00), servo.cache[4], "servo cache should NOT have been updated")

	// write too-high value with one byte
	err = servo.setRegister(Register{0x05, 1, rw, true, 2, 3}, 4)
	assert.EqualError(t, err, "value too high: 4 (max=3)")
	assert.Equal(t, byte(0x00), n.controlTable[5], "control table should NOT have been written")
	assert.Equal(t, byte(0x00), servo.cache[5], "servo cache should NOT have been updated")
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

func TestServoID(t *testing.T) {
	n, s := servo(map[int]byte{
		0x03: 0x1,
	})

	// read
	val, err := s.ServoID()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	// min
	err = s.SetServoID(-1)
	assert.Error(t, err)

	// max
	err = s.SetServoID(253)
	assert.Error(t, err)

	// write
	err = s.SetServoID(2)
	assert.NoError(t, err)
	assert.Equal(t, byte(2), n.controlTable[0x03])
	assert.Equal(t, byte(2), s.cache[0x03])

	// re-read
	val, err = s.ServoID()
	assert.NoError(t, err)
	assert.Equal(t, 2, val)
}

// BaudRate
// SetBaudRate

// ReturnDelayTime
// SetReturnDelayTime

// CWAngleLimit
// SetCWAngleLimit

// CCWAngleLimit
// SetCCWAngleLimit

// HighestLimitTemperature
// SetHighestLimitTemperature

// LowestLimitVoltage
// SetLowestLimitVoltage

// HighestLimitVoltage
// SetHighestLimitVoltage

// MaxTorque
// SetMaxTorque

// StatusReturnLevel
// SetStatusReturnLevel

// AlarmLed
// SetAlarmLed

// AlarmShutdown
// SetAlarmShutdown

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

func TestCWComplianceMargin(t *testing.T) {
	n, s := servo(map[int]byte{
		0x1a: 0x1,
	})

	// read
	val, err := s.CWComplianceMargin()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	// min
	err = s.SetCWComplianceMargin(-1)
	assert.Error(t, err)

	// max
	err = s.SetCWComplianceMargin(1024)
	assert.Error(t, err)

	// write
	err = s.SetCWComplianceMargin(2)
	assert.NoError(t, err)
	assert.Equal(t, byte(2), n.controlTable[0x1a])
	assert.Equal(t, byte(2), s.cache[0x1a])

	// re-read
	val, err = s.CWComplianceMargin()
	assert.NoError(t, err)
	assert.Equal(t, 2, val)
}

// CcwComplianceMargin
// SetCcwComplianceMargin
// CwComplianceSlope
// SetCwComplianceSlope
// CcwComplianceSlope
// SetCcwComplianceSlope

func TestGoalPosition(t *testing.T) {
	_, s := servo(map[int]byte{
		0x1e: 0xff, // L
		0x1f: 0x03, // H
	})

	val, err := s.GoalPosition()
	assert.NoError(t, err)
	assert.Equal(t, 1023, val)
}

func TestSetGoalPosition(t *testing.T) {
	n, s := servo(map[int]byte{})

	// valid
	err := s.SetGoalPosition(513)
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n.controlTable[0x1e]) // L
	assert.Equal(t, byte(2), n.controlTable[0x1f]) // H
	assert.Equal(t, byte(1), s.cache[0x1e])        // L
	assert.Equal(t, byte(2), s.cache[0x1f])        // H

	// too low
	err = s.SetGoalPosition(-1)
	assert.Error(t, err)

	// too high
	err = s.SetGoalPosition(1025)
	assert.Error(t, err)
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

func TestTorqueLimit(t *testing.T) {
	n, s := servo(map[int]byte{
		0x22: 0xff, // L
		0x23: 0x03, // H
	})

	// read
	val, err := s.TorqueLimit()
	assert.NoError(t, err)
	assert.Equal(t, 1023, val)

	// min
	err = s.SetTorqueLimit(-1)
	assert.Error(t, err)

	// max
	err = s.SetTorqueLimit(1024)
	assert.Error(t, err)

	// write
	err = s.SetTorqueLimit(513)
	assert.NoError(t, err)
	assert.Equal(t, byte(1), n.controlTable[0x22]) // L
	assert.Equal(t, byte(2), n.controlTable[0x23]) // H
	assert.Equal(t, byte(1), s.cache[0x22])        // L
	assert.Equal(t, byte(2), s.cache[0x23])        // H

	// re-read
	val, err = s.TorqueLimit()
	assert.NoError(t, err)
	assert.Equal(t, 513, val)
}

func TestPresentPosition(t *testing.T) {
	_, s := servo(map[int]byte{
		0x24: byte(1), // L
		0x25: 0x00,    // H
	})

	val, err := s.PresentPosition()
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

func TestPresentTemperature(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2b: 0x55,
	})

	val, err := s.PresentTemperature()
	assert.NoError(t, err)
	assert.Equal(t, 85, val)
}

func TestRegistered(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2c: 1,
	})

	val, err := s.Registered()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestMoving(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2e: 0,
	})

	val, err := s.Moving()
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestLock(t *testing.T) {

	// not locked
	_, s1 := servo(map[int]byte{
		0x2f: 0,
	})

	val, err := s1.Lock()
	assert.NoError(t, err)
	assert.Equal(t, 0, val)

	// locked
	_, s2 := servo(map[int]byte{
		0x2f: 1,
	})

	val, err = s2.Lock()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestSetLock(t *testing.T) {
	_, s := servo(map[int]byte{
		0x2f: 0,
	})

	// okay to re-unlock
	err := s.SetLock(0)
	assert.NoError(t, err)

	// lock
	err = s.SetLock(1)
	assert.NoError(t, err)

	// can't unlock when already locked
	err = s.SetLock(0)
	assert.Error(t, err)
}

// Punch
// SetPunch

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
// initially contains the given bytes. The control table is empty, except that
// the servo ID is 1, and the status return level is 2. (This is just to avoid
// having to specify the same values for every test.)
func servo(b map[int]byte) (*mockNetwork, *DynamixelServo) {
	n := &mockNetwork{}

	n.controlTable[0x03] = byte(0x01) // servoID
	n.controlTable[0x10] = byte(0x02) // statusReturnLevel

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
