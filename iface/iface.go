package iface

// TODO: Use an io.writer instead?
type Logger interface {
	Printf(format string, v ...interface{})
}

// Networker provides an interface to the underlying servos' control tables by
// reading and writing to/from the network interface.
type Networker interface {
	Ping(uint8) error
	ReadData(uint8, byte, int) ([]byte, error)
	WriteData(uint8, bool, ...byte) error
}
