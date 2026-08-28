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
	named, outPkg, err := loadPackage(dir, *typeName)
	if err != nil {
		fatal(err)
	}
	model, err := buildModel(named, outPkg, *noValidate)
	if err != nil {
		fatal(err)
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
