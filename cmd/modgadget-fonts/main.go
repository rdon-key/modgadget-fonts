package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rdon-key/modgadget-fonts/internal/bdf"
	"github.com/rdon-key/modgadget-fonts/internal/generator"
)

const commandName = "modgadget-fonts"

type options struct {
	bdfPath      string
	subsetPath   string
	packageName  string
	variableName string
	outputPath   string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet(commandName, flag.ContinueOnError)
	flags.SetOutput(stderr)

	var opts options
	flags.StringVar(&opts.bdfPath, "bdf", "", "input BDF file")
	flags.StringVar(&opts.subsetPath, "subset", "", "UTF-8 subset text file")
	flags.StringVar(&opts.packageName, "package", "", "generated Go package name")
	flags.StringVar(&opts.variableName, "var", "", "generated Font variable name")
	flags.StringVar(&opts.outputPath, "o", "", "output Go file")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "%s: unexpected positional arguments: %v\n", commandName, flags.Args())
		return 2
	}
	for _, required := range []struct {
		name  string
		value string
	}{
		{"bdf", opts.bdfPath},
		{"subset", opts.subsetPath},
		{"package", opts.packageName},
		{"var", opts.variableName},
		{"o", opts.outputPath},
	} {
		if required.value == "" {
			fmt.Fprintf(stderr, "%s: missing required flag -%s\n", commandName, required.name)
			return 2
		}
	}

	if err := execute(opts); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", commandName, err)
		return 1
	}
	return 0
}

func execute(opts options) error {
	bdfFile, err := os.Open(opts.bdfPath)
	if err != nil {
		return fmt.Errorf("open BDF %q: %w", opts.bdfPath, err)
	}
	parsed, parseErr := bdf.Parse(opts.bdfPath, bdfFile)
	closeErr := bdfFile.Close()
	if parseErr != nil {
		return fmt.Errorf("parse BDF %q: %w", opts.bdfPath, parseErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close BDF %q: %w", opts.bdfPath, closeErr)
	}

	subsetText, err := os.ReadFile(opts.subsetPath)
	if err != nil {
		return fmt.Errorf("read subset %q: %w", opts.subsetPath, err)
	}
	subsetFont, err := bdf.Subset(parsed, subsetText)
	if err != nil {
		return fmt.Errorf("subset font: %w", err)
	}
	converted, err := bdf.Convert(subsetFont)
	if err != nil {
		return fmt.Errorf("convert font: %w", err)
	}
	source, err := generator.Generate(converted, generator.Options{
		PackageName:  opts.packageName,
		VariableName: opts.variableName,
	})
	if err != nil {
		return fmt.Errorf("generate source: %w", err)
	}
	if err := os.WriteFile(opts.outputPath, source, 0o644); err != nil {
		return fmt.Errorf("write output %q: %w", opts.outputPath, err)
	}
	return nil
}
