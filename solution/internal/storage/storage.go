package storage

import (
	"errors"

	"solution/internal/models"
)

var (
	ErrCountryNotFound = errors.New("country not found")
)

type Storage interface {
	Countries(regions ...string) ([]models.Country, error)
	Disconnect() error
}
