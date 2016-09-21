package iface

// TODO: Use an io.writer instead?
type Logger interface {
	Printf(format string, v ...interface{})
}

// Protocol provides an abstract interface to command servos.
type Protocol interface {
	Ping(uint8) error

	ReadData(ident int, address int, length int) ([]byte, error)
	WriteData(ident int, address int, data []byte, expectResponse bool) error

	// TODO: Move this to Networker
	SetLogger(logger Logger)

	// TODO: Combine this with Logger
	Logf(format string, v ...interface{})
}
