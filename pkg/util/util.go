package util

import (
	"encoding/base64"
	"github.com/tae2089/bob-logging/logger"
	"strings"
)

func GenerateResourceName(names ...string) string {
	newNameList := make([]string, 0)
	for _, name := range names {
		if name != "" {
			newNameList = append(newNameList, name)
		}
	}
	return strings.Join(newNameList, "-")
}

func DecodingBase64(data []byte) string {
	if data == nil || len(data) == 0 {
		return ""
	}
	output, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		logger.Error(err)
		return ""
	}
	return string(output)
}
