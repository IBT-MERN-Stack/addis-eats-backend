package delivery

import "errors"

var ErrUnservedArea = errors.New("unserved delivery area")

// ServedAreas acts as the single source of truth for valid delivery zones [20, 21].
var ServedAreas = []string{"Bole", "Kazanchis", "Piassa", "Megenagna"}

var feeTable = map[string]int{
	"Bole":      3000,
	"Kazanchis": 3500,
	"Piassa":    4000,
	"Megenagna": 4500,
}

func FeeSantim(area string) (int, error) {
	fee, ok := feeTable[area]
	if !ok {
		return 0, ErrUnservedArea
	}
	return fee, nil
}