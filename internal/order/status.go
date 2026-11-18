package order

import "encoding/json"

type Status int

const (
	Placed Status = iota
	Accepted
	Cooking
	OnTheWay
	Delivered
)

func (s Status) String() string {
	switch s {
	case Placed:
		return "placed"
	case Accepted:
		return "accepted"
	case Cooking:
		return "cooking"
	case OnTheWay:
		return "on_the_way"
	case Delivered:
		return "delivered"
	default:
		return "unknown"
	}
}

// MarshalJSON satisfies json.Marshaler so status encodes as "placed" instead of integer 0.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}