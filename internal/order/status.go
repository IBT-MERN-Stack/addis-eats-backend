package order

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrIllegalTransition signals an invalid state machine progression.
var ErrIllegalTransition = errors.New("illegal status transition")

// Status uses Go's iota pattern to represent an enum set of state steps.
type Status int

const (
	StatusPlaced Status = iota
	StatusAccepted
	StatusCooking
	StatusOnTheWay
	StatusDelivered
)

// String satisfies fmt.Stringer to provide human-readable text representations.
func (s Status) String() string {
	switch s {
	case StatusPlaced:
		return "placed"
	case StatusAccepted:
		return "accepted"
	case StatusCooking:
		return "cooking"
	case StatusOnTheWay:
		return "on_the_way"
	case StatusDelivered:
		return "delivered"
	default:
		return "unknown"
	}
}

func ParseStatus(str string) (Status, error) {
	switch str {
	case "placed":
		return StatusPlaced, nil
	case "accepted":
		return StatusAccepted, nil
	case "cooking":
		return StatusCooking, nil
	case "on_the_way":
		return StatusOnTheWay, nil
	case "delivered":
		return StatusDelivered, nil
	default:
		return Status(-1), fmt.Errorf("invalid status string: %s", str)
	}
}

// CanTransitionTo enforces strict unidirectional state progression.
// TEACHING NOTE: Enforcing order workflow rules prevents illegal state jumps
// (e.g., moving an order from "delivered" back to "cooking") [8].
func (s Status) CanTransitionTo(next Status) bool {
	return next == s+1
}

// MarshalJSON satisfies json.Marshaler.
// TEACHING NOTE: Without this method, Go would serialize `Status` as an integer (`0`, `1`, `2`).
// Implementing `MarshalJSON` converts it to human-readable strings (`"placed"`, `"accepted"`) on the wire [7].
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON satisfies json.Unmarshaler to parse incoming JSON strings into `Status` enums.
func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	st, err := ParseStatus(str)
	if err != nil {
		return err
	}
	*s = st
	return nil
}