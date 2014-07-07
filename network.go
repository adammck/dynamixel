package dynamixel

import (
  "io"
  "bytes"
)

const (

  // Instruction Types
  Ping      byte = 0x01
  ReadData  byte = 0x02
  WriteData byte = 0x03
  RegWrite  byte = 0x04
  Action    byte = 0x05
  Reset     byte = 0x06
  SyncWrite byte = 0x83
)

type DynamixelNetwork struct {
  Serial io.ReadWriteCloser
}

func NewNetwork(serial io.ReadWriteCloser) *DynamixelNetwork {
  return &DynamixelNetwork{
    Serial: serial,
  }
}

//
// This stuff is generic to all Dynamixels. See:
//
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_packet.htm
// * http://support.robotis.com/en/product/dynamixel/communication/dxl_instruction.htm
//
func (n *DynamixelNetwork) WriteInstruction(ident uint8, instruction byte, params ...byte) error {

  buf := new(bytes.Buffer)
  paramsLength := byte(len(params) + 2)

  // build instruction packet

  buf.Write([]byte{
    0xFF, 0xFF,         // instruction header
    byte(ident),        // target Dynamixel ID
    byte(paramsLength), // len(params) + 2
    instruction,        // instruction type (read/write/etc)
  })

  buf.Write(params)

  // calculate checksum

  sum := ident + paramsLength + instruction

  for _, value := range params {
      sum += value
  }

  buf.WriteByte(byte((^sum) & 0xFF))

  // write to port

  _, err := buf.WriteTo(n.Serial)

  if err != nil {
    return err
  }

  return nil
}

func (n *DynamixelNetwork) WriteData(ident uint8, params ...byte) error {
  return n.WriteInstruction(ident, WriteData, params...)
}
