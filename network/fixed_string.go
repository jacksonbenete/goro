package network

import (
	"strings"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func encodeROFixedString(value string, length int) []byte {
	out := make([]byte, length)
	value = strings.TrimSpace(value)
	if value == "" || length <= 0 {
		return out
	}
	encoded, _, err := transform.Bytes(korean.EUCKR.NewEncoder(), []byte(value))
	if err != nil {
		encoded = []byte(value)
	}
	if len(encoded) >= length {
		encoded = encoded[:length-1]
	}
	copy(out, encoded)
	return out
}
