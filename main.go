package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/ec2-macos-utils/cmd"
	"github.com/aws/ec2-macos-utils/internal/contextual"
	"github.com/aws/ec2-macos-utils/pkg/system"
)

// TODO: this will become cmd/ec2-macos-utils/main.go
// TODO: see https://code.amazon.com/packages/JakeevShare/trees/heads/go/conceptual/context-data
func main() {
	sys, err := system.Scan()
	if err != nil {
		panic(fmt.Errorf("cannot identify system: %w", err))
	}
	p := sys.Product()
	if p == nil {
		panic("no product associated with identified system")
	}

	ctx := contextual.WithProduct(context.Background(), p)

	if err := cmd.MainCommand().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
