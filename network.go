package dynamixel

import (
  "io"
  "bytes"
  "encoding/binary"
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

  // beginning of packet
  buf.Write([]byte{0xFF, 0xFF})

  // target Dynamixel ID
  binary.Write(buf, binary.LittleEndian, ident)

  // length (== number of parameters plus two)
  length := uint8(len(params) + 2)
  binary.Write(buf, binary.LittleEndian, length)

  // instruction
  buf.WriteByte(instruction)

  // parameter 0..n
  buf.Write(params)

  // checksum
  sum := ident + length + instruction
  for _, value := range params {
      sum += value
  }
  sum = (^sum) & 0xFF
  binary.Write(buf, binary.LittleEndian, sum)

  _, err := buf.WriteTo(n.Serial);

  if err != nil {
    return nil
  }

  return nil
}

func (n *DynamixelNetwork) WriteData(ident uint8, params ...byte) error {
  return n.WriteInstruction(ident, WriteData, params...)
}
