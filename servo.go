package dynamixel

import (
	"errors"
	"fmt"
	"math"
)

const (

	// Control table size (in bytes)
	tableSize = 50

	// Control Table Addresses (EEPROM)
	addrID                byte = 0x03 // 1
	addrStatusReturnLevel byte = 0x10 // 1

	// Control Table Addresses (RAM, Read/Write)
	addrTorqueEnable byte = 0x18 // 1
	addrLed          byte = 0x19 // 1
	addrGoalPosition byte = 0x1E // 2
	addrMovingSpeed  byte = 0x20 // 2
	addrTorqueLimit  byte = 0x22 // 2

	// Control Table Addresses (RAM, Read Only)
	addrCurrentPosition byte = 0x24 // 2
	addrPresentVoltage  byte = 0x2A // 1

	// Limits (from dxl_ax_actuator.htm)
	maxPos   uint16  = 1023
	maxSpeed uint16  = 1023
	maxAngle float64 = 300

	// Unit conversions
	positionToAngle float64 = maxAngle / float64(maxPos) // 0.293255132
	angleToPosition float64 = 1 / positionToAngle        // 3.41
)

// Networker provides an interface to the underlying servos' control tables.
type Networker interface {
	Ping(uint8) error
	ReadData(uint8, byte, int) ([]byte, error)
	ReadInt(uint8, byte, int) (int, error)
	WriteData(uint8, bool, ...byte) error
	Log(string, ...interface{})
}

type DynamixelServo struct {
	Network   Networker
	Ident     uint8
	zeroAngle float64

	// Cache of control table values
	cache             [tableSize]byte
	statusReturnLevel int
}

// http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
func NewServo(network Networker, ident uint8) *DynamixelServo {
	s := &DynamixelServo{
		Network:           network,
		Ident:             ident,
		zeroAngle:         150,
		statusReturnLevel: 2,
	}

	return s
}

// updateCache reads the entire control table from the servo, and stores it in
// the cache.
func (servo *DynamixelServo) updateCache() error {
	b, err := servo.Network.ReadData(servo.Ident, 0x0, tableSize)
	if err != nil {
		return err
	}

	// Ensure that the returned slice is the right size.
	if len(b) != tableSize {
		return fmt.Errorf("invalid control table size: %d (expected %d)", len(b), tableSize)
	}

	// Copy each byte to the cache.
	// TODO: Surely there is a better way to do this.
	for i := 0; i < tableSize; i++ {
		servo.cache[i] = b[i]
	}

	return nil
}

// getRegister fetches the value of a register from the cache.
func (servo *DynamixelServo) getRegister(reg Register) (int, error) {
	if reg.length != 1 && reg.length != 2 {
		return 0, fmt.Errorf("invalid register length: %d", reg.length)
	}

	if reg.cacheable {
		v := int(servo.cache[reg.address])

		if reg.length == 2 {
			v |= int(servo.cache[reg.address+1]) << 8
		}

		return v, nil
	} else {
		if servo.statusReturnLevel == 0 {
			return 0, errors.New("can't READ while Status Return Level is zero")
		}

		b, err := servo.Network.ReadData(servo.Ident, reg.address, reg.length)
		if err != nil {
			return 0, err
		}

		switch len(b) {
		case 1:
			servo.cache[reg.address] = b[0]
			return int(b[0]), nil

		case 2:
			servo.cache[reg.address] = b[0]
			servo.cache[reg.address+1] = b[1]
			return int(b[0]) | int(b[1])<<8, nil

		default:
			return 0, fmt.Errorf("expected %d bytes, got %d", reg.length, len(b))

		}
	}
}

// setRegister writes a value to the given register. Returns an error if the
// register is read only or if the write failed.
func (servo *DynamixelServo) setRegister(reg Register, value int) error {
	if reg.access == ro {
		return fmt.Errorf("can't write to a read-only register")
	}

	switch reg.length {
	case 1:
		servo.writeData(reg.address, low(value))
		servo.cache[reg.address] = low(value)

	case 2:
		servo.writeData(reg.address, low(value), high(value))
		servo.cache[reg.address] = low(value)
		servo.cache[reg.address+1] = high(value)

	default:
		return fmt.Errorf("invalid register length: %d", reg.length)
	}

	return nil
}

// Ping sends the PING instruction to servo, and waits for the response. Returns
// nil if the ping succeeds, otherwise an error. It's optional, but a very good
// idea, to call this before sending any other instructions to the servo.
func (servo *DynamixelServo) Ping() error {
	return servo.Network.Ping(servo.Ident)
}

// Converts a bool to an int.
func btoi(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func low(i int) byte {
	return byte(i & 0xFF)
}

func high(i int) byte {
	return low(i >> 8)
}

func (servo *DynamixelServo) readInt(addr byte, n int) (int, error) {
	if servo.statusReturnLevel == 0 {
		return 0, errors.New("can't READ while Status Return Level is zero")
	}

	return servo.Network.ReadInt(servo.Ident, addr, n)
}

// TODO: Remove this in favor of setRegister?
func (servo *DynamixelServo) writeData(params ...byte) error {
	return servo.Network.WriteData(servo.Ident, (servo.statusReturnLevel == 2), params...)
}

func posDistance(a uint16, b uint16) uint16 {
	return uint16(math.Abs(float64(a) - float64(b)))
}

//
func normalizeAngle(d float64) float64 {
	if d > 180 {
		return normalizeAngle(d - 360)

	} else if d < -180 {
		return normalizeAngle(d + 360)

	} else {
		return d
	}
}

//
// -- High-level interface
//
//    These methods should provide as useful and friendly of an interface to the
//    servo as possible.

func (servo *DynamixelServo) posToAngle(pos int) float64 {
	return (positionToAngle * float64(pos)) - servo.zeroAngle
}

func (servo *DynamixelServo) angleToPos(angle float64) int {
	return int((servo.zeroAngle + angle) * angleToPosition)
}

// Sets the origin angle (in degrees).
func (servo *DynamixelServo) SetZero(offset float64) {
	servo.zeroAngle = offset
}

//
// Returns the current position of the servo, relative to the zero angle.
//
func (servo *DynamixelServo) Angle() (float64, error) {
	pos, err := servo.Position()

	if err != nil {
		return 0, err

	} else {
		return servo.posToAngle(pos), nil
	}
}

// MoveTo sets the goal position of the servo by angle (in degrees), where zero
// is the midpoint, 150 deg is max left (clockwise), and -150 deg is max right
// (counter-clockwise). This is generally preferable to calling SetGoalPosition,
// which uses the internal uint16 representation.
//
// If the angle is out of bounds
//
func (servo *DynamixelServo) MoveTo(angle float64) error {
	pos := servo.angleToPos(normalizeAngle(angle))
	return servo.SetGoalPosition(pos)
}

//
// -- Low-level interface
//
//    These methods should follow the Dynamixel protocol docs as closely as
//    possible, with no fancy stuff.
//

// Enables or disables torque.
func (servo *DynamixelServo) SetTorqueEnable(state bool) error {
	servo.logMethod("SetTorqueEnable(%t)", state)
	return servo.writeData(addrTorqueEnable, btoi(state))
}

// Enables or disables the LED.
func (servo *DynamixelServo) SetLed(state bool) error {
	servo.logMethod("SetLed(%t)", state)
	return servo.writeData(addrLed, btoi(state))
}

// Sets the goal position.
// See: http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm#Actuator_Address_1E
func (servo *DynamixelServo) SetGoalPosition(pos int) error {
	if pos < 0 || pos > int(maxPos) {
		return errors.New("goal position out of range")
	}
	return servo.writeData(addrGoalPosition, low(pos), high(pos))
}

// Sets the moving speed.
func (servo *DynamixelServo) SetMovingSpeed(speed int) error {
	if speed < 0 || speed > int(maxSpeed) {
		return errors.New("moving speed out of range")
	}
	return servo.writeData(addrMovingSpeed, low(speed), high(speed))
}

// Sets the torque limit.
func (servo *DynamixelServo) SetTorqueLimit(limit int) error {
	servo.logMethod("SetTorqueLimit(%d)", limit)

	if limit < 0 || limit > 1023 {
		return errors.New("torque limit out of range")
	}
	return servo.writeData(addrTorqueLimit, low(limit), high(limit))
}

// Sets the status return level. Possible values are:
//
// 0 = Only respond to PING commands
// 1 = Only respond to PING and READ commands
// 2 = Respond to all commands
//
// Servos default to 2, but retain the value so long as they're powered up. This
// makes it a very good idea to explicitly set the value after connecting, to
// avoid waiting for status packets which will never arrive.
//
// See: dxl_ax_actuator.htm#Actuator_Address_10
func (servo *DynamixelServo) SetStatusReturnLevel(value int) error {
	servo.logMethod("SetStatusReturnLevel(%d)", value)

	if value < 0 || value > 2 {
		return fmt.Errorf("invalid Status Return Level value: %d", value)
	}

	// Call Network.WriteData directly, rather than via servo.writeData, because
	// the return status level will depend upon the new level, rather than the
	// current level cache. We don't want to update that until we're sure that
	// the write was successful.
	err := servo.Network.WriteData(servo.Ident, (value == 2), addrStatusReturnLevel, low(value))
	if err != nil {
		return err
	}

	servo.statusReturnLevel = value
	return nil
}

// Voltage returns the current voltage supplied. Unlike the underlying Dynamixel
// interface, this is the actual voltage, not multiplied by ten.
func (servo *DynamixelServo) Voltage() (float64, error) {
	volts, err := servo.readInt(addrPresentVoltage, 1)
	if err != nil {
		return 0.0, err
	}

	// Convert the return value into actual volts.
	return (float64(volts) / 10), nil
}

// Returns the current position.
func (servo *DynamixelServo) Position() (int, error) {
	return servo.readInt(addrCurrentPosition, 2)
}

// Changes the identity of the servo.
// This is stored in EEPROM, so will persist between reboots.
func (servo *DynamixelServo) SetIdent(ident int) error {
	servo.logMethod("SetIdent(%d, %d)", ident)
	i := low(ident)

	if i < 0 || i > 252 {
		return fmt.Errorf("invalid ID (must be 0-252): %d", i)
	}

	err := servo.writeData(addrID, i)
	if err != nil {
		return err
	}

	servo.Ident = i
	return nil
}

func (servo *DynamixelServo) logMethod(format string, v ...interface{}) {
	prefix := fmt.Sprintf("servo[%d].", servo.Ident)
	servo.Network.Log(prefix+format, v...)
}
