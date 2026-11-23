package validator

import (
	"regexp"
	"slices"
)

// PhoneRx compiles the Ethiopian mobile format (09 or 07 followed by 8 digits) [15].
// Pre-compiled at package startup to avoid per-request compilation overhead [9].
var PhoneRx = regexp.MustCompile(`^0[79][0-9]{8}$`)

// Matches tests if a string satisfies a compiled regular expression [14].
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
 
// In checks if a string exists within a set of allowed values [14].
func In(value string, list ...string) bool {
	return slices.Contains(list, value)
}

// Unique verifies that all elements in a slice are distinct [14].
func Unique(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, v := range values {
		if seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}