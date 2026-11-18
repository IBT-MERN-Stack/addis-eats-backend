package order

import (
	"errors"
	"time"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/delivery"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
)

var (
	ErrEmptyOrder = errors.New("order must contain at least one item")
	ErrNotFound   = errors.New("order not found")
)

type Order struct {
	ID          string      `json:"id"`
	Area        string      `json:"area"`
	Phone       string      `json:"phone"`
	Items       []menu.Dish `json:"items"`
	Status      Status      `json:"status"`
	TotalSantim int         `json:"total_santim"`
	PlacedAt    time.Time   `json:"placed_at"`
}

func New(area, phone string, items []menu.Dish) Order {
	return Order{
		Area:   area,
		Phone:  phone,
		Items:  items,
		Status: Placed,
	}
}

func (o *Order) ComputeTotal() (int, error) {
	if len(o.Items) == 0 {
		return 0, ErrEmptyOrder
	}

	subtotal := 0
	for _, item := range o.Items {
		subtotal += item.PriceSantim
	}

	serviceFee := int(float64(subtotal) * 0.05)
	deliveryFee := 0

	if subtotal < 50000 {
		fee, err := delivery.FeeSantim(o.Area)
		if err != nil {
			return 0, err
		}
		deliveryFee = fee
	}

	o.TotalSantim = subtotal + serviceFee + deliveryFee
	return o.TotalSantim, nil
}