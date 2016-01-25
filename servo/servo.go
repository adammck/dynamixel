package servo

import (
	"fmt"
	"github.com/adammck/dynamixel/network"
	reg "github.com/adammck/dynamixel/registers"
	"github.com/adammck/dynamixel/utils"
)

const (

	// Control table size (in bytes)
	// TODO: Instead of hard-coding this, maybe calculate the size by finding
	//       the highest register address and adding its length?
	tableSize = 50

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

	// Cache of control table values.
	cache [tableSize]byte
}

// New returns a new Servo with its cache populated.
// TODO: Return a pointer, error tuple! We're currently ignoring the return
//       value of the populateCache call.
func New(network network.Networker, ID int, registers reg.Map) *Servo {
	s := &Servo{
		Network:   network,
		ID:        ID,
		registers: registers,
		zeroAngle: 150,
	}

	_ = s.populateCache()
	return s
}

// populateCache reads the entire control table from the servo, and stores it in
// the cache.
func (servo *Servo) populateCache() error {
	b, err := servo.Network.ReadData(uint8(servo.ID), 0x0, tableSize)
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

// getRegister fetches the value of a register from the cache if possible,
// otherwise reads it from the control table (and caches it).
func (servo *Servo) getRegister(n reg.RegName) (int, error) {
	r, ok := servo.registers[n]
	if !ok {
		return 0, fmt.Errorf("can't read unsupported register: %s", n)
	}

	if r.Length != 1 && r.Length != 2 {
		return 0, fmt.Errorf("invalid register length: %d", r.Length)
	}

	// If the register is cacheable, read the value from the cache and return
	// that. The populateCache method must have been called first.

	if r.Cacheable {
		v := int(servo.cache[r.Address])

		if r.Length == 2 {
			v |= int(servo.cache[r.Address+1]) << 8
		}

		return v, nil
	}

	// Abort if return level is zero.

	// rl, err := servo.StatusReturnLevel()
	// if err != nil {
	// 	return 0, err
	// }
	// if rl == 0 {
	// 	return 0, errors.New("can't READ while Status Return Level is zero")
	// }

	// Read the single value from the control table.

	b, err := servo.Network.ReadData(uint8(servo.ID), r.Address, r.Length)
	if err != nil {
		return 0, err
	}

	switch len(b) {
	case 1:
		servo.cache[r.Address] = b[0]
		return int(b[0]), nil

	case 2:
		servo.cache[r.Address] = b[0]
		servo.cache[r.Address+1] = b[1]
		return int(b[0]) | int(b[1])<<8, nil

	default:
		return 0, fmt.Errorf("expected %d bytes, got %d", r.Length, len(b))

	}
}

// setRegister writes a value to the given register. Returns an error if the
// register is read only or if the write failed.
func (servo *Servo) setRegister(n reg.RegName, value int) error {
	r, ok := servo.registers[n]
	if !ok {
		return fmt.Errorf("can't write to unsupported register: %s", n)
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
		servo.cache[r.Address] = utils.Low(value)

	case 2:
		servo.Network.WriteData(uint8(servo.ID), (rl == 2), r.Address, utils.Low(value), utils.High(value))
		servo.cache[r.Address] = utils.Low(value)
		servo.cache[r.Address+1] = utils.High(value)

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
