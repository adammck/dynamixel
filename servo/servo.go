package servo

import (
	"errors"
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

	returnLevelValue int
	returnLevelKnown bool

	// The map of register names to locations in the control table. This
	// (unfortunately) varies between models, so can't be const.
	registers reg.Map

	// TODO: Remove this!
	zeroAngle float64
}

// New returns a new Servo.
func New(network network.Networker, registers reg.Map, ID int) *Servo {
	return &Servo{
		Network:   network,
		ID:        ID,
		registers: registers,
		zeroAngle: 150,
	}
}

// NewWithReturnLevel returns a servo with its Return Level preconfigured. It's
// better to use New and SetReturnLevel to be sure, but this can be useful when
// we're absolutely sure what the return level currently is.
func NewWithReturnLevel(network network.Networker, registers reg.Map, ID int, returnLevel int) *Servo {
	s := New(network, registers, ID)
	s.returnLevelValue = returnLevel
	s.returnLevelKnown = true
	return s
}

// SetReturnLevel sets the return level. Possible values are:
//
//   0 = Only respond to PING commands
//   1 = Only respond to PING and READ commands
//   2 = Respond to all commands
//
// The factory default setting is 2, but this register is persisted in EEPROM,
// so does not reset when power-cycled. To avoid waiting for a response from a
// servo which will never respond, or (worse) receiving unexpected responses,
// use this method to set the value explicitly immediately after connecting.
//
// See: dxl_ax_actuator.htm#Actuator_Address_10
func (s *Servo) SetReturnLevel(value int) error {
	reg := s.registers[reg.StatusReturnLevel]

	if value < reg.Min || value > reg.Max {
		return fmt.Errorf("invalid Status Return Level value: %d", value)
	}

	ident, err := s.ServoID()
	if err != nil {
		return err
	}

	// Call Network.WriteData directly, rather than via writeData, because the
	// return status level will depend upon the new level, rather than the
	// current level. We don't want to update that until we're sure that the write
	// was successful.
	err = s.Network.WriteData(uint8(ident), (value == 2), reg.Address, utils.Low(value))
	if err != nil {
		return err
	}

	return nil
}

// ReturnLevel returns the current return level of the servo, or an error if we
// don't know. This method will never actually read from the control table,
// because it's expected to be called by getters are setters.
func (servo *Servo) ReturnLevel() (int, error) {
	if !servo.returnLevelKnown {
		return 0, errors.New("current Return Level is unknown")
	}

	return servo.returnLevelValue, nil
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

	rl, err := servo.ReturnLevel()
	if err != nil {
		return 0, err
	}
	if rl == 0 {
		return 0, errors.New("can't READ while Return Level is zero")
	}

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

	// Refuse to write if we don't know the return level, because we can't know
	// whether to wait for a status packet or not.
	rl, err := servo.ReturnLevel()
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
