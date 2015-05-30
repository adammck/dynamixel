package dynamixel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockNetwork struct {
	remoteControlTable [50]byte
}

func (n *mockNetwork) Ping(ident uint8) error {
	return nil
}

func (n *mockNetwork) ReadData(ident uint8, addr byte, count int) ([]byte, error) {
	return n.remoteControlTable[int(addr) : int(addr)+count], nil
}

// TODO: Move this into Servo?
func (n *mockNetwork) ReadInt(ident uint8, addr byte, count int) (int, error) {
	return 0, nil
}

func (n *mockNetwork) WriteData(ident uint8, expectStausPacket bool, params ...byte) error {
	return nil
}

func (n *mockNetwork) Log(string, ...interface{}) {}

func TestUpdateCache(t *testing.T) {
	n := &mockNetwork{}
	n.remoteControlTable = [50]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
	}

	servo := NewServo(n, 1)
	err := servo.updateCache()

	assert.Nil(t, err)
	assert.Equal(t, servo.cache, n.remoteControlTable)
}