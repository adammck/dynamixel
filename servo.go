package dynamixel

const (

  // Control Table Addresses
  ModelNumber byte = 0x00
  Led         byte = 0x19
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
  return servo.writeData(Led, btoi(status))
}
