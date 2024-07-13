package erolog

import (
	"context"
	"math"
)

type ContextBuilder struct {
	ctx    *Context
	parent context.Context
}

func NewContextBuilder() *ContextBuilder {
	return BuilderFrom(context.Background())
}

func BuilderFrom(parent context.Context) *ContextBuilder {
	return &ContextBuilder{
		ctx:    NewContext(),
		parent: parent,
	}
}

func (b *ContextBuilder) Build() *Context {
	return b.ctx.ExtractOrThis(b.parent)
}

func (b *ContextBuilder) BuildContext() context.Context {
	return b.ctx.Enriched(b.parent)
}

func (b *ContextBuilder) WithParent(parent context.Context) *ContextBuilder {
	b.parent = parent
	return b
}

func (b *ContextBuilder) With(key string, value interface{}) *ContextBuilder {
	b.ctx.attrs[key] = value
	return b
}

func (b *ContextBuilder) WithMap(mp map[string]interface{}) *ContextBuilder {
	for k, v := range mp {
		b.ctx.attrs[k] = v
	}
	return b
}

func (b *ContextBuilder) WithSecret(key string, value string, encryptingPercentage int) *ContextBuilder {
	encryptingPercentage = int(math.Max(math.Min(float64(encryptingPercentage), 100.0), 0.0))
	secretVal := secretValue{
		value:                value,
		encryptingPercentage: encryptingPercentage,
	}
	return b.With(key, secretVal)
}

func (b *ContextBuilder) WithDomain(domain string) *ContextBuilder {
	b.ctx.domain = domain
	return b
}
