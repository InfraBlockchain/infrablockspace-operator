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

func TestEncodingBase64(t *testing.T) {
	data := []string{"test", "an3", "b"}
	t.Run("generating name test 1", func(t *testing.T) {
		result := GenerateResourceName(data...)
		assert.Equal(t, "test-an3-b", result)
	})
	data = []string{"test", "", "b"}
	t.Run("generating name test 2", func(t *testing.T) {
		result := GenerateResourceName(data...)
		assert.Equal(t, "test-b", result)
	})
	t.Run("generating name test 3", func(t *testing.T) {
		result := GenerateResourceName(data...)
		result = result + "-service"
		assert.Equal(t, "test-b-service", result)
	})
}
