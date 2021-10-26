package contextual

import (
	"context"

	"github.com/aws/ec2-macos-utils/pkg/system"
)

// productKey is used to set and retrieve context held values for Product.
var productKey = struct{}{}

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
