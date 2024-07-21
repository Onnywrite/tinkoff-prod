package handler

import (
	"context"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

// It exists because I'm lazy to wrap wrap wrap and wrap ero.Error again.
// Of course it's better to not have 2 same structs expecially
// when ero.Error with the same purpose exists
type ErrorMessage string

func (e ErrorMessage) Blob() []byte {
	return []byte(`{"Service":"` + ero.CurrentService + `","ErrorMessage":"` + string(e) + `"}`)
}

type UserByIdProvider interface {
	UserById(ctx context.Context, id uint64) (*models.User, ero.Error)
}

type UserProvider interface {
	UserByEmail(ctx context.Context, email string) (*models.User, ero.Error)
}
