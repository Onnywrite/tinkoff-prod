package countries

import "errors"

var (
	ErrCountriesNotFound = errors.New("countries not found")
	ErrCountryNotFound   = errors.New("country not found")
	ErrBadAlpha2         = errors.New("code does not seem to be an alpha2")
	ErrInternal          = errors.New("internal error")
)
