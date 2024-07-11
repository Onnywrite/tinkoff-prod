package ero

import (
	"context"
	"encoding/json"
)

type ValidationError[T any, TCtx LogContext[TCtx]] struct {
	Service          string
	ValidationErrors []T
	c                TCtx
}

func NewValidation[T any, TCtx LogContext[TCtx]](ctx TCtx, validationErrors []T) Error {
	return &ValidationError[T, TCtx]{
		Service:          CurrentService,
		ValidationErrors: validationErrors,
		c:                ctx,
	}
}

func (e *ValidationError[T, TCtx]) Error() string {
	b, text := json.Marshal(e)
	if text != nil {
		panic(text)
	}
	return string(b)
}

func (e *ValidationError[T, TCtx]) Is(error) bool {
	panic("ValidationErrors cannot be compared")
}

func (e *ValidationError[T, TCtx]) Unwrap() error {
	panic("ValidationErrors cannot be unwrapped")
}

func (e *ValidationError[T, TCtx]) Code() int {
	return CodeBadRequest
}

func (e *ValidationError[T, TCtx]) Context(ctx context.Context) context.Context {
	return e.c.Enriched(ctx)
}

func (e *ValidationError[T, TCtx]) Wrap(context.Context) Error {
	panic("ValidationErrors cannot be wrapped")
}

func (e *ValidationError[T, TCtx]) WrapCode(context.Context, int) Error {
	panic("ValidationErrors cannot be wrapped")
}
