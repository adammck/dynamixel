package iface

// TODO: Use an io.writer instead?
type Logger interface {
	Printf(format string, v ...interface{})
}

// Networker provides an interface to the underlying servos' control tables by
// reading and writing to/from the network interface.
type Networker interface {
	Ping(uint8) error
	// ReadData(uint8, byte, int) ([]byte, error)
	// WriteData(uint8, bool, ...byte) error

	ReadData(ident int, address int, length int) ([]byte, error)
	WriteData(ident int, address int, data []byte, expectResponse bool) error

	SetLogger(logger Logger)

	// TODO: Combine this with Logger
	Logf(format string, v ...interface{})
}
