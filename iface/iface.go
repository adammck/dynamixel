package iface

// TODO: Use an io.writer instead?
type Logger interface {
	Printf(format string, v ...interface{})
}

// Protocol provides an abstract interface to command servos. This exists so
// that our abstract Servo type can communicate with actual servos regardless
// which protocol version they speak.
//
// The interface must be the union of all protocol versions, but (so far) they
// all have roughly the same instructions, so this isn't a big deal.
type Protocol interface {
	Ping(ident int) error

	ReadData(ident int, address int, length int) ([]byte, error)

	// WriteData writes a slice of bytes to the control table of the given servo
	// ID.
	WriteData(ident int, address int, data []byte, expectResponse bool) error

	// RegWrite writes a slice of bytes to the control table of the given servo
	// ID, like WriteData, but the resulting instruction (e.g. set goal, torque)
	// is not executed until the Action method is called.
	RegWrite(ident int, address int, data []byte, expectResponse bool) error

	// Action causes writes buffered by the RegWrite method to be executed. This
	// is useful to update the state of many servos simultaneously.
	Action() error

	// FactoryReset() error
	// Reboot() error
	// SyncRead() error
	// SyncWrite() error
	// BulkRead() error
	// BulkWrite() error
}
