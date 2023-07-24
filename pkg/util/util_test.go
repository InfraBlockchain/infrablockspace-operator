package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodingBase64(t *testing.T) {
	sentence := "YmFiZQ=="
	t.Run("success base64 decoding", func(t *testing.T) {
		result := DecodingBase64([]byte(sentence))
		assert.Equal(t, "babe", result)
	})

	t.Run("failed base64 decoding", func(t *testing.T) {
		result := DecodingBase64([]byte(""))
		assert.Equal(t, "", result)
	})
}
