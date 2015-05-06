package dynamixel

import (
	"errors"
	"fmt"
	"math"
)

const (

	// Control Table Addresses (EEPROM)
	addrID                byte = 0x03 // 1
	addrStatusReturnLevel byte = 0x10 // 1

	// Control Table Addresses (RAM, Read/Write)
	addrTorqueEnable      byte = 0x18 // 1
	addrLed               byte = 0x19 // 1
	addrGoalPosition      byte = 0x1E // 2
	addrMovingSpeed       byte = 0x20 // 2
	addrTorqueLimit       byte = 0x22 // 2

	// Control Table Addresses (RAM, Read Only)
	addrCurrentPosition byte = 0x24 // 2

	// Limits (from dxl_ax_actuator.htm)
	maxPos   uint16  = 1023
	maxSpeed uint16  = 1023
	maxAngle float64 = 300

	// Unit conversions
	positionToAngle float64 = maxAngle / float64(maxPos) // 0.293255132
	angleToPosition float64 = 1 / positionToAngle        // 3.41
)

type DynamixelServo struct {
	Network   *DynamixelNetwork
	Ident     uint8
	zeroAngle float64

	// Cache of control table values
	statusReturnLevel int
}

// http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
func NewServo(network *DynamixelNetwork, ident uint8) *DynamixelServo {
	s := &DynamixelServo{
		Network:           network,
		Ident:             ident,
		zeroAngle:         150,
		statusReturnLevel: 2,
	}

	return s
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

func (servo *DynamixelServo) readData(startAddress byte, length int) (uint16, error) {
	if servo.statusReturnLevel == 0 {
		return 0, errors.New("can't READ while Status Return Level is zero")
	}

	return servo.Network.ReadData(servo.Ident, startAddress, length)
}

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

func (servo *DynamixelServo) posToAngle(pos uint16) float64 {
	return (positionToAngle * float64(pos)) - servo.zeroAngle
}

func (servo *DynamixelServo) angleToPos(angle float64) uint16 {
	pos := uint16((servo.zeroAngle + angle) * angleToPosition)
	return pos
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
	pos := int(servo.angleToPos(normalizeAngle(angle)))
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

// Returns the current position.
func (servo *DynamixelServo) Position() (uint16, error) {
	return servo.readData(addrCurrentPosition, 2)
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
	servo.Network.Log(prefix + format, v...)
}
