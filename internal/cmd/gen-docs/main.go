package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra/doc"

	"github.com/aws/ec2-macos-utils/internal/cmd"
)

func main() {
	outdir := "./docs"
	args := os.Args
	if len(args) >= 2 {
		outdir = args[1]
	}

	logrus.WithField("outdir", outdir).Info("generating docs")

	if err := os.MkdirAll(outdir, 0755); err != nil {
		panic(err)
	}
	err := doc.GenMarkdownTree(cmd.MainCommand(), outdir)
	if err != nil {
		panic(err)
	}

	logrus.WithField("outdir", outdir).Info("generated docs")
}
