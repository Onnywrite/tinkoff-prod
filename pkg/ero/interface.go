package ero

import "context"

type Error interface {
	error
	Is(error) bool
	Unwrap() error

	Code() int
	Context(context.Context) context.Context
	Wrap(context.Context) Error
	WrapCode(context.Context, int) Error
}

type LogContext[T any] interface {
	Enriched(context.Context) context.Context
	ExtractOrThis(ctx context.Context) T
}

func Wrap(ctx context.Context, erro Error) Error {
	return WrapCode(ctx, CodeUnknownServer, erro)
}

func WrapCode(ctx context.Context, code int, erro Error) Error {
	if erro == nil {
		return nil
	}
	return erro.Wrap(ctx)
}
