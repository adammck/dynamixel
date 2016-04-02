package servo

import (
	"fmt"
	"github.com/adammck/dynamixel/network"
	reg "github.com/adammck/dynamixel/registers"
	"github.com/adammck/dynamixel/utils"
)

const (

	// Unit conversions
	maxPos          uint16  = 1023
	maxAngle        float64 = 300
	positionToAngle float64 = maxAngle / float64(maxPos) // 0.293255132
	angleToPosition float64 = 1 / positionToAngle        // 3.41
)

type Servo struct {
	Network network.Networker
	ID      int

	// The map of register names to locations in the control table. This
	// (unfortunately) varies between models, so can't be const.
	registers reg.Map

	// TODO: Remove this!
	zeroAngle float64
}

// New returns a new Servo.
func New(network network.Networker, ID int, registers reg.Map) *Servo {
	return &Servo{
		Network:   network,
		ID:        ID,
		registers: registers,
		zeroAngle: 150,
	}
}

// getRegister fetches the value of a register from the control table.
func (servo *Servo) getRegister(n reg.RegName) (int, error) {
	r, ok := servo.registers[n]
	if !ok {
		return 0, fmt.Errorf("can't read unsupported register: %v", n)
	}

	if r.Length != 1 && r.Length != 2 {
		return 0, fmt.Errorf("invalid register length: %d", r.Length)
	}

	// Abort if return level is zero.

	// rl, err := servo.StatusReturnLevel()
	// if err != nil {
	// 	return 0, err
	// }
	// if rl == 0 {
	// 	return 0, errors.New("can't READ while Status Return Level is zero")
	// }

	b, err := servo.Network.ReadData(uint8(servo.ID), r.Address, r.Length)
	if err != nil {
		return 0, err
	}

	if len(b) != r.Length {
		return 0, fmt.Errorf("expected %d bytes, got %d", r.Length, len(b))
	}

	return utils.BytesToInt(b)
}

// setRegister writes a value to the given register. Returns an error if the
// register is read only or if the write failed.
func (servo *Servo) setRegister(n reg.RegName, value int) error {
	r, ok := servo.registers[n]
	if !ok {
		return fmt.Errorf("can't write to unsupported register: %v", n)
	}

	if r.Access == reg.RO {
		return fmt.Errorf("can't write to a read-only register")
	}

	if value < r.Min {
		return fmt.Errorf("value too low: %d (min=%d)", value, r.Min)
	}

	if value > r.Max {
		return fmt.Errorf("value too high: %d (max=%d)", value, r.Max)
	}

	rl, err := servo.StatusReturnLevel()
	if err != nil {
		return err
	}

	// TODO: Add log message when setting a register.
	switch r.Length {
	case 1:
		servo.Network.WriteData(uint8(servo.ID), (rl == 2), r.Address, utils.Low(value))

	case 2:
		servo.Network.WriteData(uint8(servo.ID), (rl == 2), r.Address, utils.Low(value), utils.High(value))

	default:
		return fmt.Errorf("invalid register length: %d", r.Length)
	}

	return nil
}

// Ping sends the PING instruction to servo, and waits for the response. Returns
// nil if the ping succeeds, otherwise an error. It's optional, but a very good
// idea, to call this before sending any other instructions to the servo.
func (servo *Servo) Ping() error {
	return servo.Network.Ping(uint8(servo.ID))
}
