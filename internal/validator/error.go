package validator

import "fmt"

// Error allows internal domain packages to return field-level validation errors
// without importing net/http or coupling to presentation layers [18, 19].
type Error struct {
	Fields map[string]string
}

func (e *Error) Error() string {
	return fmt.Sprintf("validation failed on %d field(s)", len(e.Fields))
}