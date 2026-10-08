// configulator generates reflection-free configuration loaders.
//
// Usage (from a //go:generate directive in the config type's package):
//
//	configulator -type Config [-output config_configulator.go]
//	             [-flags pflag|std|none] [-no-validate]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	typeName := flag.String("type", "", "config root type (required)")
	output := flag.String("output", "", "output file (default <type>_configulator.go)")
	flagsMode := flag.String("flags", flagsPFlag, "flag adapter: pflag | std | none")
	noValidate := flag.Bool("no-validate", false, "allow a config type without Validate() error")
	schema := flag.Bool("schema", false, "print a JSON Schema to stdout instead of generating")
	sample := flag.Bool("sample", false, "print a sample config to stdout instead of generating (see -format)")
	format := flag.String("format", "yaml", "sample format: yaml (commented) | json | toml")
	markdown := flag.Bool("markdown", false, "print a Markdown reference table of every key to stdout instead of generating")
	envPrefix := flag.String("env-prefix", "", "env var prefix shown in -markdown output (verbatim)")
	envSep := flag.String("env-separator", "_", "env var separator shown in -markdown output")
	flagSep := flag.String("flag-separator", ".", "flag separator shown in -markdown output")
	markdownFile := flag.String("markdown-file", "", "with -markdown, write the table between the "+markdownBegin+" and "+markdownEnd+" markers in this file instead of stdout")
	check := flag.Bool("check", false, "with -markdown-file, change nothing and exit 1 if the file is out of date")
	pkgDir := flag.String("dir", ".", "directory of the package that declares -type")
	flag.Parse()

	if *typeName == "" {
		fmt.Fprintln(os.Stderr, "configulator: -type is required")
		os.Exit(2)
	}
	switch *flagsMode {
	case flagsPFlag, flagsStd, flagsNone:
	default:
		fmt.Fprintf(os.Stderr, "configulator: -flags must be pflag, std, or none (got %q)\n", *flagsMode)
		os.Exit(2)
	}

	if *markdownFile != "" && !*markdown {
		fmt.Fprintln(os.Stderr, "configulator: -markdown-file needs -markdown")
		os.Exit(2)
	}
	if *check && *markdownFile == "" {
		fmt.Fprintln(os.Stderr, "configulator: -check needs -markdown-file")
		os.Exit(2)
	}

	dir, err := filepath.Abs(*pkgDir)
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
	if *flagsMode == flagsStd {
		if p := findShortTag(model.Fields, ""); p != "" {
			fatal(fmt.Errorf("%s: short: tag is not supported with -flags=std (stdlib flag has no shorthands)", p))
		}
	}

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
	if *format != "yaml" && !*sample {
		fmt.Fprintln(os.Stderr, "configulator: -format only applies to -sample")
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
			writeSample(model, *format)
		case *markdown:
			table := emitMarkdown(model, *flagSep, *envPrefix, *envSep, *markdownFile == "")
			if *markdownFile == "" {
				os.Stdout.Write(table)
				return
			}
			changed, err := updateMarkdownFile(*markdownFile, table, *check)
			if err != nil {
				fatal(err)
			}
			if changed {
				fmt.Fprintf(os.Stderr, "configulator: updated %s\n", *markdownFile)
			}
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
	path := out
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	if err := os.WriteFile(path, src, 0o644); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "configulator: wrote %s\n", out)
}

// writeSample prints a sample config in format to stdout.
func writeSample(model *Model, format string) {
	switch format {
	case "yaml":
		os.Stdout.Write(emitSample(model))
	case "json":
		b, err := emitSampleJSON(model)
		if err != nil {
			fatal(err)
		}
		os.Stdout.Write(b)
	case "toml":
		os.Stdout.Write(emitSampleTOML(model))
	default:
		fmt.Fprintf(os.Stderr, "configulator: unknown -format %q: expected yaml, json, or toml\n", format)
		os.Exit(2)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "configulator: %v\n", err)
	os.Exit(1)
}
