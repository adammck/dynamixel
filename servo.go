package dynamixel

import (
  "errors"
)

const (

  // Control Table Addresses
  addrTorqueEnable byte = 0x18 // 1
  addrLed          byte = 0x19 // 1
  addrGoalPosition byte = 0x1E // 2
  addrMovingSpeed  byte = 0x20 // 2
)

type DynamixelServo struct {
  Network *DynamixelNetwork
  Ident uint8
}

// http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
func NewServo(network *DynamixelNetwork, ident uint8) *DynamixelServo {
  return &DynamixelServo{
    Network: network,
    Ident: ident,
  }
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

func (servo *DynamixelServo) writeData(params ...byte) error {
  return servo.Network.WriteData(servo.Ident, params...)
}

// Enables or disables torque.
func (servo *DynamixelServo) SetTorqueEnable(state bool) error {
  return servo.writeData(addrTorqueEnable, btoi(state))
}

// Enables or disables the LED.
func (servo *DynamixelServo) SetLed(state bool) error {
  return servo.writeData(addrLed, btoi(state))
}

// Sets the goal position.
func (servo *DynamixelServo) SetGoalPosition(pos int) error {
  if pos < 0 || pos > 1023 {
    return errors.New("goal position out of range")
  }
  return servo.writeData(addrGoalPosition, low(pos), high(pos))
}

// Sets the moving speed.
func (servo *DynamixelServo) SetMovingSpeed(speed int) error {
  if speed < 0 || speed > 1023 {
    return errors.New("moving speed out of range")
  }
  return servo.writeData(addrMovingSpeed, low(speed), high(speed))
}
