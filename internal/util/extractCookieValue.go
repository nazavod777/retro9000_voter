package util

import "strings"

func ExtractCookieValue(rawCookie string, cookieName string) string {
	if strings.HasPrefix(rawCookie, cookieName+"=") {
		rawCookie = strings.TrimPrefix(rawCookie, cookieName+"=")
		parts := strings.Split(rawCookie, ";")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}
