package dynamixel

type RegName int
type Access int

const (

	// Access Levels specify whether a register is hard-coded into the servo
	// (e.g. the model number), or is a value which can be changed (e.g. the
	// identity).
	RO Access = iota
	RW
)

type Register struct {
	Address   byte
	Length    int
	Access    Access
	Cacheable bool

	// The range of values which this register can be set to. This only applies
	// is the register is RW.
	Min int
	Max int
}
