package service

import "strings"

func containsSubstring(text, pattern string) bool {
	if text == "" || pattern == "" {
		return false
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}
