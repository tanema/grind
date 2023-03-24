package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScreenBufRender(t *testing.T) {
	var buf bytes.Buffer
	screen := NewScreenBuf(&buf)
	err := screen.Render(`{{. | green}}
{{"world" | bold}}`, "hello")
	assert.Nil(t, err)
	assert.Equal(t, "\x1b[32mhello\x1b[m\n\x1b[1mworld\x1b[m\n", screen.buf.String())
	assert.Equal(t, "\x1b[32mhello\x1b[m\n\x1b[1mworld\x1b[m\n", buf.String())
}

func TestScreenBufReset(t *testing.T) {
	var buf bytes.Buffer
	screen := NewScreenBuf(&buf)
	screen.buf.WriteString(`This is
A Buffer full of
Lines that need to
Be Reset`)
	err := screen.Reset()
	assert.Nil(t, err)
	assert.Equal(t, strings.Repeat(clearLastLine, 3), screen.buf.String())
}

func TestScreenBufWrite(t *testing.T) {
	var buf bytes.Buffer
	screen := NewScreenBuf(&buf)
	err := screen.Write(`{{. | green}}
{{"world" | bold}}`, "hello")
	assert.Nil(t, err)
	assert.Equal(t, "\x1b[32mhello\x1b[m\n\x1b[1mworld\x1b[m\n", screen.buf.String())
}

func TestScreenBufFlush(t *testing.T) {
	var buf bytes.Buffer
	str := `This is
A Buffer full of
Lines that need to
Be Reset`
	screen := NewScreenBuf(&buf)
	screen.buf.WriteString(str)
	err := screen.Flush()
	assert.Nil(t, err)
	assert.Equal(t, str, buf.String())
}
