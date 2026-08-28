// configulator generates reflection-free configuration loaders.
//
// Usage (from a //go:generate directive in the config type's package):
//
//	configulator -type Config [-output Config_configulator.go]
//	             [-flags pflag|std|none] [-no-validate]
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	typeName := flag.String("type", "", "config root type (required)")
	output := flag.String("output", "", "output file (default <type>_configulator.go)")
	flagsMode := flag.String("flags", "pflag", "flag adapter: pflag | std | none")
	noValidate := flag.Bool("no-validate", false, "allow a config type without Validate() error")
	schema := flag.Bool("schema", false, "print a JSON Schema to stdout instead of generating")
	sample := flag.Bool("sample", false, "print a commented YAML sample config to stdout instead of generating")
	markdown := flag.Bool("markdown", false, "print a Markdown reference table of every key to stdout instead of generating")
	envPrefix := flag.String("env-prefix", "", "env var prefix shown in -markdown output (verbatim)")
	envSep := flag.String("env-separator", "_", "env var separator shown in -markdown output")
	flagSep := flag.String("flag-separator", ".", "flag separator shown in -markdown output")
	flag.Parse()

	if *typeName == "" {
		fmt.Fprintln(os.Stderr, "configulator: -type is required")
		os.Exit(2)
	}
	switch *flagsMode {
	case "pflag", "std", "none":
	default:
		fmt.Fprintf(os.Stderr, "configulator: -flags must be pflag, std, or none (got %q)\n", *flagsMode)
		os.Exit(2)
	}

	dir, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	named, outPkg, err := loadPackage(dir, *typeName, nil)
	if err != nil {
		fatal(err)
	}
	model, err := buildModel(named, outPkg, *noValidate)
	if err != nil {
		fatal(err)
	}
	if *flagsMode == "std" {
		if p := findShortTag(model.Fields, ""); p != "" {
			fatal(fmt.Errorf("%s: short: tag is a generate-time error under -flags=std — stdlib flag has no shorthand concept", p))
		}
	}

	// Emit modes: print exactly one document to stdout and generate
	// nothing, matching the Rust configulator-cli.
	modes := 0
	for _, b := range []*bool{schema, sample, markdown} {
		if *b {
			modes++
		}
	}
	if modes > 1 {
		fmt.Fprintln(os.Stderr, "configulator: pass at most one of -schema, -sample, -markdown; output goes to stdout, pipe it where you want it")
		os.Exit(2)
	}
	if modes == 1 {
		switch {
		case *schema:
			b, err := emitJSONSchema(model)
			if err != nil {
				fatal(err)
			}
			os.Stdout.Write(b)
		case *sample:
			os.Stdout.Write(emitSample(model))
		case *markdown:
			os.Stdout.Write(emitMarkdown(model, *flagSep, *envPrefix, *envSep))
		}
		return
	}

	out := *output
	if out == "" {
		out = strings.ToLower(*typeName) + "_configulator.go"
	}
	src, err := emit(model, *flagsMode)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(out, src, 0o644); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "configulator: wrote %s\n", out)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "configulator: %v\n", err)
	os.Exit(1)
}
