package envfile

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvFileParse(t *testing.T) {
	info := map[string]string{}
	assert.Nil(t, Parse(info, "./test/valid.env", ""))
	assert.Equal(t, "xyz123", info["SIMPLE"])
	assert.Equal(t, `Multiple\nLines and variable substitution: xyz123`, info["INTERPOLATED"])
	assert.Equal(t, "raw text without variable interpolation", info["NON_INTERPOLATED"])
	assert.Equal(t, `long text here,
e.g. a private SSH key`, info["MULTILINE"])
	assert.Equal(t, `long text here as well,
yes`, info["OTHERMULTILINE"])
	assert.Equal(t, `long text here as well,
yes`, info["HEREDOC"])

	err := Parse(info, "./test/notther.env")
	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}
