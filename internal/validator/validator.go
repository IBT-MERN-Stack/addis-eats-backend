package validator

// Validator accumulates field-level validation errors.
// It imports nothing outside the standard library and remains detached from HTTP logic [6, 7].
type Validator struct {
	Errors map[string]string `json:"fields"`
}

// New constructs an initialized Validator.
// Always use New() because a write to a nil map causes a runtime panic [7, 13].
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Valid returns true if no validation errors were recorded [7].
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError records the first validation failure per JSON field key [7, 13].
func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check evaluates a rule condition. If false, it registers the error message [7].
// Condition must evaluate to TRUE for valid input (e.g., v.Check(x != "", "field", "message")) [13].
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}