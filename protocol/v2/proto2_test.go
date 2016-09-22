package v2

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RW struct {
	io.Reader
	io.Writer
}

func TestProto2WriteData(t *testing.T) {

	// Helper to write a some test data.
	write := func(pp *Proto2, expectResponse bool) error {
		return pp.WriteData(1, 2, []byte{3, 4, 5}, expectResponse)
	}

	// Expected to be written every time the func above is run.
	expWrite := []byte{
		0xFF, // header
		0xFF, // "
		0xFD, // "
		0x0,  // reserved
		0x1,  // ident
		0x8,  // plen+3 (lsb)
		0x0,  // "      (msb)
		0x3,  // inst
		0x2,  // params | addr (lsb)
		0x0,  //        | "    (msb)
		0x3,  //        | data
		0x4,  //        | "
		0x5,  //        | "
		0x3D, // crc (msb)
		0x30} //     (lsb)

	// ReturnLevel < 2 (no response) -------------------------------------------

	r := bytes.NewReader(nil)
	w := &bytes.Buffer{}
	b := &RW{r, w}
	p := New(b)

	err := write(p, false)
	if assert.NoError(t, err) {
		assert.Equal(t, expWrite, w.Bytes())
	}

	// ReturnLevel == 2 (valid response) ---------------------------------------

	r = bytes.NewReader([]byte{0xFF, 0xFF, 0xFD, 0x00, 0x01, 0x00, 0x00, 0x55, 0x00, 0xFF, 0xFF})
	w = &bytes.Buffer{}
	b = &RW{r, w}
	p = New(b)

	err = write(p, true)
	assert.Equal(t, expWrite, w.Bytes())
	if assert.NoError(t, err) {
		assert.Equal(t, w.Bytes(), expWrite)
	}

	// ReturnLevel == 2 (invalid responses) ------------------------------------

	errExamples := []struct {
		buf []byte
		err string
	}{

		// Receiving the following erroneous responses from the network should
		// cause the given errors to be returned.
		{[]byte{}, "reading packet header: EOF"},
		{[]byte{0x00}, "reading packet header: expected 9 bytes, got 1"},
		{[]byte{0x1a, 0x2b, 0x3c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, "bad status packet header: 0x1A 0x2B 0x3C"},
		{[]byte{0xff, 0xff, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x99, 0x00}, "bad status packet instruction: 0x99"},
		{[]byte{0xff, 0xff, 0xfd, 0x00, 0x02, 0x00, 0x00, 0x55, 0x00}, "reading checksum: EOF"},
		{[]byte{0xff, 0xff, 0xfd, 0x00, 0x02, 0x00, 0x00, 0x55, 0x00, 0x00}, "reading checksum: expected 2 bytes, got 1"},
		{[]byte{0xff, 0xff, 0xfd, 0x00, 0x02, 0x00, 0x00, 0x55, 0x00, 0xFF, 0xFF}, "expected status packet for 1, but got 2"},
	}

	for _, eg := range errExamples {

		r = bytes.NewReader(eg.buf)
		w = &bytes.Buffer{}
		b = &RW{r, w}
		p = New(b)

		err = write(p, true)
		assert.Equal(t, expWrite, w.Bytes())
		if assert.Error(t, err) {
			assert.EqualError(t, err, eg.err)
		}
	}

}
