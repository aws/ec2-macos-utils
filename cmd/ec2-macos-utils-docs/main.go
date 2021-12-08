// +build docs

package main

import (
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/aws/ec2-macos-utils/internal/cmd"
)

func main() {
	outdir := "./docs"
	args := os.Args
	if len(args) >= 2 {
		outdir = args[1]
	}

	os.MkdirAll(outdir, 0755)
	err := doc.GenMarkdownTree(cmd.MainCommand(), outdir)
	if err != nil {
		os.Exit(1)
	}
}
