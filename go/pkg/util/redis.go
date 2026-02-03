package util

import (
	"net/url"
	"strings"
)

// NormalizeRedisAddr converts a redis URL into a host:port address.
// If the input is already a host:port pair or is invalid, it returns the input unchanged.
func NormalizeRedisAddr(raw string) string {
	if raw == "" {
		return raw
	}
	if strings.HasPrefix(raw, "redis://") || strings.HasPrefix(raw, "rediss://") {
		parsed, err := url.Parse(raw)
		if err == nil && parsed.Host != "" {
			return parsed.Host
		}
	}
	return raw
}
