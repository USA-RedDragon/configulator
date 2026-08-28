//go:build goexperiment.jsonv2

package conformance

import (
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
)

// The corpus runner: enumerates every case under spec/cases and fails on
// any case it cannot execute unless the case is named in spec/skip-go.txt
// (which CI prints and which must be empty for a release).

func specDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", "..", "..", "spec"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(d, "SPEC.md")); err != nil {
		t.Fatalf("spec dir not found at %s: %v", d, err)
	}
	return d
}

type caseInput struct {
	dir     string
	shape   string
	config  []byte // nil if no config.json
	env     map[string]string
	argv    []string
	opts    caseOptions
	expect  map[string]any
	experr  *expectError
	origins map[string]expectOrigin
}

type caseOptions struct {
	Prefix        string   `json:"prefix"`
	EnvSeparator  string   `json:"env_separator"`
	FlagSeparator string   `json:"flag_separator"`
	ArraySep      string   `json:"array_separator"`
	Capabilities  []string `json:"capabilities"`
}

type expectError struct {
	Kind     string   `json:"kind"`
	Contains []string `json:"contains"`
}

type expectOrigin struct {
	Layer  string `json:"layer"`
	Detail string `json:"detail"`
}

func loadCase(t *testing.T, dir string) caseInput {
	t.Helper()
	tc := caseInput{dir: dir, opts: caseOptions{
		Prefix: "APP_", EnvSeparator: "_", FlagSeparator: ".", ArraySep: ",",
	}}
	shape, err := os.ReadFile(filepath.Join(dir, "shape"))
	if err != nil {
		t.Fatalf("case %s: no shape file: %v", dir, err)
	}
	tc.shape = strings.TrimSpace(string(shape))

	readJSON := func(name string, into any) bool {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return false
		}
		if err := jsonv2.Unmarshal(b, into); err != nil {
			t.Fatalf("case %s: bad %s: %v", dir, name, err)
		}
		return true
	}

	if b, err := os.ReadFile(filepath.Join(dir, "config.json")); err == nil {
		tc.config = b
	}
	readJSON("env.json", &tc.env)
	readJSON("argv.json", &tc.argv)
	readJSON("options.json", &tc.opts)
	if tc.opts.Prefix == "" {
		tc.opts.Prefix = "APP_"
	}
	if tc.opts.EnvSeparator == "" {
		tc.opts.EnvSeparator = "_"
	}
	if tc.opts.FlagSeparator == "" {
		tc.opts.FlagSeparator = "."
	}
	if tc.opts.ArraySep == "" {
		tc.opts.ArraySep = ","
	}
	readJSON("expect.json", &tc.expect)
	var ee expectError
	if readJSON("expect_errors.json", &ee) {
		tc.experr = &ee
	}
	readJSON("expect_origins.json", &tc.origins)
	if tc.expect == nil && tc.experr == nil {
		t.Fatalf("case %s: neither expect.json nor expect_errors.json", dir)
	}
	return tc
}

// runShape executes the full pipeline for one shape instantiation.
func runShape[C any](t *testing.T, tc caseInput, schema *configulator.Schema[C], hooks *cpflag.Hooks[C], toTree func(*C) any) {
	t.Helper()
	c := configulator.New(schema).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
			Prefix: tc.opts.Prefix, Separator: tc.opts.EnvSeparator,
		}).
		WithArraySeparator(tc.opts.ArraySep).
		WithEnviron(func(k string) (string, bool) { v, ok := tc.env[k]; return v, ok })

	if tc.config != nil {
		c.WithFile(&configulator.FileOptions{Search: []string{filepath.Join(tc.dir, "config.json")}})
	} else if len(tc.argv) > 0 {
		// --config cases need the file layer enabled even with no search file
		c.WithFile(&configulator.FileOptions{})
	}

	if len(tc.argv) > 0 {
		if hooks == nil {
			t.Fatalf("case %s: argv given but shape %q has no flag hooks", tc.dir, tc.shape)
		}
		fs := pflag.NewFlagSet("corpus", pflag.ContinueOnError)
		cpflag.Bind(c, fs, *hooks, &cpflag.Options{Separator: tc.opts.FlagSeparator})
		if err := fs.Parse(tc.argv); err != nil {
			t.Fatalf("case %s: parsing argv: %v", tc.dir, err)
		}
	}

	cfg, err := c.Load()

	if tc.experr != nil {
		if err == nil {
			t.Fatalf("case %s: expected error kind %s, got nil", tc.dir, tc.experr.Kind)
		}
		checkErrorKind(t, tc, err)
		for _, want := range tc.experr.Contains {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("case %s: error %q does not contain %q", tc.dir, err, want)
			}
		}
		return
	}
	if err != nil {
		t.Fatalf("case %s: unexpected error: %v", tc.dir, err)
	}

	actual := toTree(cfg)
	var normalized any
	b, merr := jsonv2.Marshal(actual)
	if merr != nil {
		t.Fatalf("case %s: marshaling actual: %v", tc.dir, merr)
	}
	if err := jsonv2.Unmarshal(b, &normalized); err != nil {
		t.Fatal(err)
	}
	compareTrees(t, tc.dir, "", any(tc.expect), normalized)

	rep := c.Report()
	for path, want := range tc.origins {
		got, ok := rep.Origin(path)
		if !ok {
			t.Errorf("case %s: no origin recorded for %q", tc.dir, path)
			continue
		}
		if got.Layer.String() != want.Layer {
			t.Errorf("case %s: origin %q layer = %s, want %s", tc.dir, path, got.Layer, want.Layer)
		}
		if want.Detail != "" && got.Detail != want.Detail {
			t.Errorf("case %s: origin %q detail = %q, want %q", tc.dir, path, got.Detail, want.Detail)
		}
	}
}

func checkErrorKind(t *testing.T, tc caseInput, err error) {
	t.Helper()
	var ok bool
	switch tc.experr.Kind {
	case "ExplicitFileMissing":
		var e *configulator.MissingFileError
		ok = errors.As(err, &e)
	case "SearchPathUnreadable":
		var e *configulator.SearchPathError
		ok = errors.As(err, &e)
	case "ParseError", "UnknownKey":
		var pe *configulator.ParseError
		var de *configulator.DecodeError
		ok = errors.As(err, &pe) || errors.As(err, &de)
	case "ValidationError":
		ok = true // any non-nil error from Validate
	case "BadEnvOptions":
		var e *configulator.BadEnvOptionsError
		ok = errors.As(err, &e)
	default:
		t.Fatalf("case %s: unknown error kind %q in expect_errors.json", tc.dir, tc.experr.Kind)
	}
	if !ok {
		t.Errorf("case %s: error %T (%v) does not match kind %s", tc.dir, err, err, tc.experr.Kind)
	}
}

// compareTrees compares expected (from expect.json) against the actual
// value tree with SPEC normalization: numbers by value; null equals an
// empty list/map when the expectation says []/{}; absent optionals are
// null.
func compareTrees(t *testing.T, dir, path string, expected, actual any) {
	t.Helper()
	switch e := expected.(type) {
	case map[string]any:
		a, ok := actual.(map[string]any)
		if !ok {
			if actual == nil && len(e) == 0 {
				return // nil map == {}
			}
			t.Errorf("case %s: %s: expected object, got %T (%v)", dir, orRoot(path), actual, actual)
			return
		}
		for k, ev := range e {
			compareTrees(t, dir, path+"."+k, ev, a[k])
		}
		for k := range a {
			if _, ok := e[k]; !ok {
				t.Errorf("case %s: %s: unexpected key %q in actual", dir, orRoot(path), k)
			}
		}
	case []any:
		a, ok := actual.([]any)
		if !ok {
			if actual == nil && len(e) == 0 {
				return // nil slice == []
			}
			t.Errorf("case %s: %s: expected array, got %T (%v)", dir, orRoot(path), actual, actual)
			return
		}
		if len(a) != len(e) {
			t.Errorf("case %s: %s: length %d, want %d", dir, orRoot(path), len(a), len(e))
			return
		}
		for i := range e {
			compareTrees(t, dir, fmt.Sprintf("%s[%d]", path, i), e[i], a[i])
		}
	case float64:
		af, ok := toFloat(actual)
		if !ok || af != e {
			t.Errorf("case %s: %s: got %v, want %v", dir, orRoot(path), actual, e)
		}
	default:
		if fmt.Sprint(actual) != fmt.Sprint(expected) || (expected == nil) != (actual == nil) {
			t.Errorf("case %s: %s: got %#v, want %#v", dir, orRoot(path), actual, expected)
		}
	}
}

func orRoot(p string) string {
	if p == "" {
		return "(root)"
	}
	return strings.TrimPrefix(p, ".")
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}

func TestCorpus(t *testing.T) {
	spec := specDir(t)
	casesDir := filepath.Join(spec, "cases")
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}

	skip := map[string]bool{}
	if b, err := os.ReadFile(filepath.Join(spec, "skip-go.txt")); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				skip[line] = true
				t.Logf("SKIP-LIST: %s (must be empty for a release)", line)
			}
		}
	}

	ran := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			if skip[name] {
				t.Skipf("in skip-go.txt")
			}
			tc := loadCase(t, filepath.Join(casesDir, name))
			ran++
			switch tc.shape {
			case "scalars":
				h := ScalarsPFlagHooks()
				runShape(t, tc, ScalarsSchema(), &h, func(c *Scalars) any { return c })
			case "nested":
				runShape[Nested](t, tc, NestedSchema(), nil, func(c *Nested) any { return c })
			case "collections":
				runShape[Collections](t, tc, CollectionsSchema(), nil, func(c *Collections) any { return c })
			case "optionals":
				runShape[Optionals](t, tc, OptionalsSchema(), nil, func(c *Optionals) any { return c })
			case "durations":
				runShape[Durations](t, tc, DurationsSchema(), nil, func(c *Durations) any {
					// SPEC: durations serialize as Go duration strings.
					return map[string]any{"timeout": c.Timeout.String(), "label": c.Label}
				})
			default:
				t.Fatalf("case %s: unknown shape %q — enumerate-or-fail", name, tc.shape)
			}
		})
	}
	if ran == 0 && len(skip) == 0 {
		t.Fatal("corpus executed zero cases")
	}
}
