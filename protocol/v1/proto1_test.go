package v1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtoWriteInstruction(t *testing.T) {
	b := &bytes.Buffer{}
	p := New(b)

	err := p.writeInstruction(1, Ping, []byte{2, 3, 4})
	if assert.NoError(t, err) {

		//                     header----  id--  p+2-  inst  p---------------  chk-
		assert.Equal(t, []byte{0xff, 0xff, 0x01, 0x05, 0x01, 0x02, 0x03, 0x04, 0xef}, b.Bytes())
	}
}
