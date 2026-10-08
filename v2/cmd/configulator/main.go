// configulator generates reflection-free configuration loaders.
//
// Usage (from a //go:generate directive in the config type's package):
//
//	configulator -type Config [-output config_configulator.go]
//	             [-flags pflag|std|none] [-no-validate]
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/types"
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
	format := flag.String("format", formatYAML, "sample format: yaml (commented) | json | toml")
	markdown := flag.Bool("markdown", false, "print a Markdown reference table of every key to stdout instead of generating")
	envPrefix := flag.String("env-prefix", "", "env var prefix shown in -markdown output (verbatim)")
	envSep := flag.String("env-separator", "_", "env var separator shown in -markdown output")
	flagSep := flag.String("flag-separator", ".", "flag separator shown in -markdown output")
	markdownFile := flag.String("markdown-file", "", "with -markdown, write the table between the "+markdownBegin+" and "+markdownEnd+" markers in this file instead of stdout")
	sampleFile := flag.String("sample-file", "", "with -sample, write the sample to this file instead of stdout")
	check := flag.Bool("check", false, "with -markdown-file or -sample-file, change nothing and exit 1 if the file is out of date")
	pkgDir := flag.String("dir", ".", "directory of the package that declares -type")
	flag.Parse()

	u := usage{
		typeName: *typeName, flagsMode: *flagsMode, format: *format,
		schema: *schema, sample: *sample, markdown: *markdown, check: *check,
		markdownFile: *markdownFile, sampleFile: *sampleFile,
	}
	if msg := u.problem(); msg != "" {
		fmt.Fprintln(os.Stderr, "configulator: "+msg)
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
	warnUntagged(named)
	if *flagsMode == flagsStd {
		if p := findShortTag(model.Fields, ""); p != "" {
			fatal(fmt.Errorf("%s: short: tag is not supported with -flags=std (stdlib flag has no shorthands)", p))
		}
	}

	if *schema || *sample || *markdown {
		switch {
		case *schema:
			b, err := emitJSONSchema(model)
			if err != nil {
				fatal(err)
			}
			os.Stdout.Write(b)
		case *sample:
			b, err := sampleBytes(model, *format)
			if err != nil {
				fatal(err)
			}
			writeOutput(*sampleFile, b, *check, updateSampleFile)
		case *markdown:
			table := emitMarkdown(model, *flagSep, *envPrefix, *envSep, *markdownFile == "")
			writeOutput(*markdownFile, table, *check, updateMarkdownFile)
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

// writeOutput prints b, or with a path, updates that file using update.
func writeOutput(path string, b []byte, check bool, update func(string, []byte, bool) (bool, error)) {
	if path == "" {
		os.Stdout.Write(b)
		return
	}
	changed, err := update(path, b, check)
	if errors.Is(err, errStale) {
		fatal(fmt.Errorf("%w; run %s", err, rerunCommand(os.Args[1:])))
	}
	if err != nil {
		fatal(err)
	}
	if changed {
		fmt.Fprintf(os.Stderr, "configulator: updated %s\n", path)
	}
}

// sampleBytes renders a sample config in format.
func sampleBytes(model *Model, format string) ([]byte, error) {
	switch format {
	case formatYAML:
		return emitSample(model), nil
	case formatJSON:
		return emitSampleJSON(model)
	case formatTOML:
		return emitSampleTOML(model), nil
	default:
		return nil, fmt.Errorf("unknown -format %q: expected yaml, json, or toml", format)
	}
}

const (
	formatYAML = "yaml"
	formatJSON = "json"
	formatTOML = "toml"
)

// usage holds the flags checked before the package loads.
type usage struct {
	typeName, flagsMode, format     string
	schema, sample, markdown, check bool
	markdownFile, sampleFile        string
}

// problem describes the first usage error, or returns "".
func (u usage) problem() string {
	modes := 0
	for _, b := range []bool{u.schema, u.sample, u.markdown} {
		if b {
			modes++
		}
	}
	switch {
	case u.typeName == "":
		return "-type is required"
	case u.flagsMode != flagsPFlag && u.flagsMode != flagsStd && u.flagsMode != flagsNone:
		return fmt.Sprintf("-flags must be pflag, std, or none (got %q)", u.flagsMode)
	case u.markdownFile != "" && !u.markdown:
		return "-markdown-file needs -markdown"
	case u.sampleFile != "" && !u.sample:
		return "-sample-file needs -sample"
	case u.check && u.markdownFile == "" && u.sampleFile == "":
		return "-check needs -markdown-file or -sample-file"
	case modes > 1:
		return "pass at most one of -schema, -sample, -markdown; output goes to stdout, pipe it where you want it"
	case u.format != formatYAML && !u.sample:
		return "-format only applies to -sample"
	case u.format != formatYAML && u.format != formatJSON && u.format != formatTOML:
		return fmt.Sprintf("-format must be yaml, json or toml (got %q)", u.format)
	}
	return ""
}

// warnUntagged prints a warning for each exported field the generator skips
// because it has no name, json or yaml tag.
func warnUntagged(named *types.Named) {
	for _, p := range untaggedFields(named) {
		fmt.Fprintf(os.Stderr, "configulator: warning: %s has no name: tag and is skipped\n", p)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "configulator: %v\n", err)
	os.Exit(1)
}
