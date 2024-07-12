package storage

import (
	"errors"
)

var (
	ErrInternal = errors.New("an error occurred while executing a query")

	ErrNoRows               = errors.New("no rows selected")
	ErrTooManyRows          = errors.New("too many rows selected")
	ErrUniqueConstraint     = errors.New("unique constraint violation")
	ErrForeignKeyConstraint = errors.New("foreign key violation")
	ErrNotNullConstraint    = errors.New("not null constraint violation")
	ErrCheckConstraint      = errors.New("check constraint violation")
)
