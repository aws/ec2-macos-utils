package contextual

import (
	"context"

	"github.com/aws/ec2-macos-utils/internal/system"
)

type contextKey uint

const (
	// productKey is used to access current Product from context.
	productKey contextKey = iota + 1
)

// WithProduct extends the context to provide a Product.
func WithProduct(ctx context.Context, product *system.Product) context.Context {
	return context.WithValue(ctx, productKey, product)
}

// Product fetches the system's Product provided in ctx.
func Product(ctx context.Context) *system.Product {
	if val := ctx.Value(productKey); val != nil {
		if v, ok := val.(*system.Product); ok {
			return v
		}
		panic("incoherent context")
	}

	return nil
}
