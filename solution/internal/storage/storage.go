package storage

import "errors"

var (
	ErrCountryNotFound = errors.New("country not found")
)

type Storage interface {
	Disconnect() error
}
