package dynamixel

import (
  "errors"
)

const (

  // Control Table Start Addresses
  addrLed          byte = 0x19 // 1
  addrGoalPosition byte = 0x1E // 2
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

func (servo *DynamixelServo) writeData(params ...byte) error {
  return servo.Network.WriteData(servo.Ident, params...)
}

// Enables or disables the LED.
func (servo *DynamixelServo) SetLed(status bool) error {
  return servo.writeData(addrLed, btoi(status))
}

// Sets the goal position.
func (servo *DynamixelServo) SetGoalPosition(pos int) error {
  if pos < 0 || pos > 1023 {
    return errors.New("goal position out of range")
  }
  return servo.writeData(addrGoalPosition, byte(pos & 0xFF), byte((pos >> 8) & 0xFF))
}
