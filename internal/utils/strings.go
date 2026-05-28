// internal/utils/strings.go

package utils

import "strings"

func ContainsAnySubstrings(str string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

func ContainsAllSubstrings(str string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(str, sub) {
			return false
		}
	}
	return true
}
