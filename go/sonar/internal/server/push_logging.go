package server

import "strings"

func tokenPreview(token string) string {
	t := strings.TrimSpace(token)
	if t == "" {
		return ""
	}
	if len(t) <= 12 {
		return t
	}
	return t[:6] + "..." + t[len(t)-4:]
}
