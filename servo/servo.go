package servo

import (
	"errors"
	"fmt"
	"github.com/adammck/dynamixel"
	"github.com/adammck/dynamixel/network"
	"github.com/adammck/dynamixel/utils"
	"math"
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
	Network   network.Networker
	zeroAngle float64

	// Cache of control table values
	cache [tableSize]byte
}

// New returns a new Servo with its cache populated.
// TODO: Return a pointer, error tuple! We're currently ignoring the return
//       value of the populateCache call.
func New(network network.Networker, ident uint8) *Servo {
	s := &Servo{
		Network:   network,
		zeroAngle: 150,
	}

	_ = s.populateCache(ident)
	return s
}

// populateCache reads the entire control table from the servo, and stores it in
// the cache.
func (servo *Servo) populateCache(ident uint8) error {
	b, err := servo.Network.ReadData(ident, 0x0, tableSize)
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
func (servo *Servo) getRegister(reg dynamixel.Register) (int, error) {

	if reg.Length != 1 && reg.Length != 2 {
		return 0, fmt.Errorf("invalid register length: %d", reg.Length)
	}

	// If the register is cacheable, read the value from the cache and return
	// that. The populateCache method must have been called first.

	if reg.Cacheable {
		v := int(servo.cache[reg.Address])

		if reg.Length == 2 {
			v |= int(servo.cache[reg.Address+1]) << 8
		}

		return v, nil
	}

	// Abort if return level is zero.

	rl, err := servo.StatusReturnLevel()
	if err != nil {
		return 0, err
	}
	if rl == 0 {
		return 0, errors.New("can't READ while Status Return Level is zero")
	}

	// Fetch the servo ident from the cache.

	ident, err := servo.ServoID()
	if err != nil {
		return 0, err
	}

	// Read the single value from the control table.

	b, err := servo.Network.ReadData(uint8(ident), reg.Address, reg.Length)
	if err != nil {
		return 0, err
	}

	switch len(b) {
	case 1:
		servo.cache[reg.Address] = b[0]
		return int(b[0]), nil

	case 2:
		servo.cache[reg.Address] = b[0]
		servo.cache[reg.Address+1] = b[1]
		return int(b[0]) | int(b[1])<<8, nil

	default:
		return 0, fmt.Errorf("expected %d bytes, got %d", reg.Length, len(b))

	}
}

// setRegister writes a value to the given register. Returns an error if the
// register is read only or if the write failed.
func (servo *Servo) setRegister(reg dynamixel.Register, value int) error {
	if reg.Access == dynamixel.RO {
		return fmt.Errorf("can't write to a read-only register")
	}

	if value < reg.Min {
		return fmt.Errorf("value too low: %d (min=%d)", value, reg.Min)
	}

	if value > reg.Max {
		return fmt.Errorf("value too high: %d (max=%d)", value, reg.Max)
	}

	// TODO: Add log message when setting a register.
	switch reg.Length {
	case 1:
		servo.writeData(reg.Address, utils.Low(value))
		servo.cache[reg.Address] = utils.Low(value)

	case 2:
		servo.writeData(reg.Address, utils.Low(value), utils.High(value))
		servo.cache[reg.Address] = utils.Low(value)
		servo.cache[reg.Address+1] = utils.High(value)

	default:
		return fmt.Errorf("invalid register length: %d", reg.Length)
	}

	return nil
}

// Ping sends the PING instruction to servo, and waits for the response. Returns
// nil if the ping succeeds, otherwise an error. It's optional, but a very good
// idea, to call this before sending any other instructions to the servo.
func (servo *Servo) Ping() error {
	ident, err := servo.ServoID()
	if err != nil {
		return err
	}

	return servo.Network.Ping(uint8(ident))
}

// TODO: Remove this in favor of setRegister?
func (servo *Servo) writeData(params ...byte) error {
	ident, err := servo.ServoID()
	if err != nil {
		return err
	}

	rl, err := servo.StatusReturnLevel()
	if err != nil {
		return err
	}

	return servo.Network.WriteData(uint8(ident), (rl == 2), params...)
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

func (servo *Servo) logMethod(format string, v ...interface{}) {

	// Include the servo ID if possible, but log even if it's unknown.
	ident, err := servo.ServoID()
	if err != nil {
		ident = 0
	}

	prefix := fmt.Sprintf("servo[%s].", ident)
	servo.Network.Log(prefix+format, v...)
}
