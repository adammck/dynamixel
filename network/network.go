package network

import (
	"fmt"
	"io"
	"time"

	"github.com/adammck/dynamixel/iface"
)

const (

	// Send an instruction to all servos
	BroadcastIdent byte = 0xFE // 254
)

type Network struct {
	Serial io.ReadWriteCloser

	// The time to wait for a single read to complete before giving up.
	Timeout time.Duration

	// Optional Logger (which only implements Printf) to log network traffic. If
	// nil (the default), nothing is logged.
	Logger iface.Logger
}

func New(serial io.ReadWriteCloser) *Network {
	return &Network{
		Serial:  serial,
		Timeout: 10 * time.Millisecond,
		Logger:  nil,
	}
}

// read receives the next n bytes from the network, blocking if they're not
// immediately available. Returns a slice containing the bytes read. If the
// network timeout is reached, returns the bytes read so far (which might be
// none) and an error.
func (nw *Network) Read(n int) ([]byte, error) {
	start := time.Now()
	buf := make([]byte, n)
	retry := 1 * time.Millisecond
	m := 0

	for m < n {
		nn, err := nw.Serial.Read(buf[m:])
		m += nn

		nw.Logf("~~ n=%d, m=%d, nn=%d, err=%s\n", n, m, nn, err)

		// It's okay if we reached the end of the available bytes. They're
		// probably just not available yet. Other errors are fatal.
		if err != nil && err != io.EOF {
			return buf, err
		}

		// If the timeout has been exceeded, abort.
		if time.Since(start) >= nw.Timeout {
			return buf, fmt.Errorf("read timed out")
		}

		// If no bytes were read, back off exponentially. This is just to avoid
		// flooding the network with retries if a servo isn't responding.
		if nn == 0 {
			time.Sleep(retry)
			retry *= 2
		}
	}

	nw.Logf("<< %#v\n", buf)
	return buf, nil
}

func (nw *Network) Write(p []byte) (int, error) {
	nw.Logf(">> %#v\n", p)
	return nw.Serial.Write(p)
}

// Ping sends the PING instruction to the given Servo ID, and waits for the
// response. Returns an error if the ping fails, or nil if it succeeds.
func (nw *Network) Ping(ident uint8) error {
	return nil
}

// ReadData reads a slice of n bytes from the control table of the given servo
// ID. Use the bytesToInt function to convert the output to something more
// useful.
func (nw *Network) ReadData(ident int, addr int, count int) ([]byte, error) {
	return nil, nil
}

func (nw *Network) WriteData(ident int, addr int, params []byte, expectStausPacket bool) error {
	return nil
}

// Action broadcasts the ACTION instruction, which initiates any previously
// bufferred instructions. Doesn't wait for a status packet in response, because
// they are not sent in response to broadcast instructions.
func (nw *Network) Action() error {
	return nil
}

func (nw *Network) Flush() {
	buf := make([]byte, 128)
	var n int

	for {
		n, _ = nw.Serial.Read(buf)
		nw.Logf(".. %v\n", buf)
		if n == 0 {
			break
		}
	}
}

func (nw *Network) SetLogger(logger iface.Logger) {
	nw.Logger = logger
}

// Logf writes a message to the network logger, unless it's nil.
func (nw *Network) Logf(format string, v ...interface{}) {
	if nw.Logger != nil {
		nw.Logger.Printf(format, v...)
	}
}
