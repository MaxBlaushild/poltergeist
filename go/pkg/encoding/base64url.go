package encoding

import (
	"encoding/base64"
)

func DecodeBase64UrlEncodedString(str string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(str)
}

func BytesToBase64UrlEncoded(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
