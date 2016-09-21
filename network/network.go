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
func (nw *Network) Read(p []byte) (n int, err error) {
	start := time.Now()
	retry := 1 * time.Millisecond

	for n < len(p) {
		m, err := nw.Serial.Read(p[n:])
		n += m

		nw.Logf("~~ n=%d, m=%d, err=%v\n", n, m, err)

		// It's okay if we reached the end of the available bytes. They're
		// probably just not available yet. Other errors are fatal.
		if err != nil && err != io.EOF {
			return m, err
		}

		// If the timeout has been exceeded, abort.
		if time.Since(start) >= nw.Timeout {
			return n, fmt.Errorf("read timed out")
		}

		// If no bytes were read, back off exponentially. This is just to avoid
		// flooding the network with retries if a servo isn't responding.
		if m == 0 {
			time.Sleep(retry)
			retry *= 2
		}
	}

	nw.Logf("<< %#v\n", p)
	return n, nil
}

func (nw *Network) Write(p []byte) (int, error) {
	nw.Logf(">> %#v\n", p)
	return nw.Serial.Write(p)
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

// Logf writes a message to the network logger, unless it's nil.
func (nw *Network) Logf(format string, v ...interface{}) {
	if nw.Logger != nil {
		nw.Logger.Printf(format, v...)
	}
}
