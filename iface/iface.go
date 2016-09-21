package iface

// TODO: Use an io.writer instead?
type Logger interface {
	Printf(format string, v ...interface{})
}

// Protocol provides an abstract interface to command servos.
type Protocol interface {
	Ping(ident int) error
	ReadData(ident int, address int, length int) ([]byte, error)
	WriteData(ident int, address int, data []byte, expectResponse bool) error

	// FactoryReset() error
	// Reboot() error
	// SyncRead() error
	// SyncWrite() error
	// BulkRead() error
	// BulkWrite() error
}
