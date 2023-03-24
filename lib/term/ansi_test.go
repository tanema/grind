package term

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAnsiString(t *testing.T) {
	ansiStr := parseAnsiString("Hello World")
	assert.Equal(t, "Hello World", ansiStr.str)
	assert.Equal(t, []string{}, ansiStr.vals)
	assert.Equal(t, "Hello World", ansiStr.String())

	ansiStr = parseAnsiString("\033[31;4mHello World\033[m")
	assert.Equal(t, "Hello World", ansiStr.str)
	assert.Equal(t, []string{"31", "4"}, ansiStr.vals)
	assert.Equal(t, "\033[31;4mHello World\033[m", ansiStr.String())

	ansiStr = parseAnsiString("\033[mHello World\033[m")
	assert.Equal(t, "Hello World", ansiStr.str)
	assert.Equal(t, []string{}, ansiStr.vals)
	assert.Equal(t, "Hello World", ansiStr.String())
}

func TestAnsiStringAdd(t *testing.T) {
	ansiStr := parseAnsiString("\033[31;4mHello World\033[m")
	ansiStr.add("20")
	assert.Equal(t, "\033[31;4;20mHello World\033[m", ansiStr.String())
}

func TestAnsiStringReplace(t *testing.T) {
	ansiStr := parseAnsiString("\033[31;4mHello World\033[m")
	ansiStr.replace("3", "9")
	assert.Equal(t, "\033[91;4mHello World\033[m", ansiStr.String())

	ansiStr = parseAnsiString("\033[41;4mHello World\033[m")
	ansiStr.replace("4", "10")
	assert.Equal(t, "\033[101;4mHello World\033[m", ansiStr.String())
}

func TestAnsiStyler(t *testing.T) {
	replace := ansiStyler("3", "9")
	actual := replace("\033[31;4mHello World\033[m")
	assert.Equal(t, "\033[91;4mHello World\033[m", actual)

	add := ansiStyler("39")
	actual = add("\033[31;4mHello World\033[m")
	assert.Equal(t, "\033[31;4;39mHello World\033[m", actual)
}

func TestRemoveANSI(t *testing.T) {
	str := "\033[31;4mHello \033[1mWorld\033[m"
	out := removeANSI([]byte(str))
	assert.Equal(t, "Hello World", string(out))
}

func TestWrapAnsi(t *testing.T) {
	str := "\033[31;4mHello \033[1mWorld\033[m"
	out := wrapANSI(str, 10)
	assert.Equal(t, "\033[31;4mHello \033[1mWor\nld\033[m", out)
}
