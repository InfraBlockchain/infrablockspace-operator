package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodingBase64(t *testing.T) {
	sentence := "aGVsbG8="
	t.Run("success base63 decoding", func(t *testing.T) {
		result := DecodingBase64([]byte(sentence))
		assert.Equal(t, "hello", result)
	})
}
