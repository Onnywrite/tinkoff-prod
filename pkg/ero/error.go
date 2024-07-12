package ero

import (
	"context"
	"errors"
)

var CurrentService string = "SomeService"

type TheBestError[TCtx LogContext[TCtx]] struct {
	ErrorMessage error
	c            TCtx
	code         int
}

func New[TCtx LogContext[TCtx]](ctx TCtx, code int, err error) Error {
	return &TheBestError[TCtx]{
		ErrorMessage: err,
		c:            ctx,
		code:         code,
	}
}

func (e *TheBestError[T]) Error() string {
	return `{"Service":"` + CurrentService + `","ErrorMessage":"` + e.ErrorMessage.Error() + `"}`
}

func (e *TheBestError[T]) Is(anotherErr error) bool {
	return errors.Is(e.Unwrap(), anotherErr)
}

func (e *TheBestError[T]) Unwrap() error {
	return e.ErrorMessage
}

func (e *TheBestError[T]) Code() int {
	return e.code
}

func (e *TheBestError[T]) Context(ctx context.Context) context.Context {
	return e.c.Enriched(ctx)
}

func (e *TheBestError[T]) Wrap(ctx context.Context) Error {
	return e.WrapCode(ctx, e.code)
}

func (e *TheBestError[T]) WrapCode(ctx context.Context, code int) Error {
	return &TheBestError[T]{
		ErrorMessage: e.ErrorMessage,
		c:            e.c.ExtractOrThis(ctx),
		code:         code,
	}
}
