package sensitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWord(t *testing.T) {
	w, err := NewWord(OTHER_FILE)
	assert.NoError(t, err)
	pass, str := w.Validate("第一夫人")
	assert.Equal(t, false, pass)
	assert.Equal(t, "第一夫人", str)

	str = w.Replace("你是协警", '！')
	assert.Equal(t, "你是！！", str)
}
