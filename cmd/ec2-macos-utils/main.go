package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/ec2-macos-utils/internal/cmd"
	"github.com/aws/ec2-macos-utils/internal/contextual"
	"github.com/aws/ec2-macos-utils/internal/system"
)

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
