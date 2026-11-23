package validator

const MaxOrderSantim = 5_000_00 // 5,000.00 ETB

type createDishRequest struct {
	PriceSantim int `json:"price_santim" validate:"required,min=1,max=500000,modulo=25"`
}

func (in createDishRequest) Validate(v *Validator) {
	v.Check(in.PriceSantim > 0,
		"price_santim", "must be greater than zero")
	v.Check(in.PriceSantim <= MaxOrderSantim,
		"price_santim", "must not exceed 5,000.00 ETB")
	v.Check(in.PriceSantim%25 == 0,
		"price_santim", "must be a multiple of 25 santim")
}