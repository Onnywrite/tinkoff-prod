package erolog

import "context"

type Context struct {
	domain string
	attrs  map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		attrs: make(map[string]interface{}),
	}
}

func (c *Context) Enriched(ctx context.Context) context.Context {
	return context.WithValue(ctx, logCtxKey, c)
}

func (c *Context) ExtractOrThis(ctx context.Context) *Context {
	newCtx := c
	if v, ok := getContext(ctx); ok {
		newCtx = v
	}
	return newCtx
}

type logKey struct{}

var logCtxKey = logKey{}

func getContext(ctx context.Context) (*Context, bool) {
	if v := ctx.Value(logCtxKey); v != nil {
		return v.(*Context), true
	}
	return nil, false
}

func Contextualize(ctx context.Context, logCtx *Context) context.Context {
	return context.WithValue(ctx, logCtxKey, logCtx)
}
