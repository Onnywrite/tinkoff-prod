package storage

import (
	"errors"
)

var (
	ErrInternal          = errors.New("an error occurred while executin a query")
	ErrCountryNotFound   = errors.New("country not found")
	ErrCountriesNotFound = errors.New("countries not found")
)
