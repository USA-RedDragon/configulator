package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// bt turns every ' in s into a backquote, so fixture sources with struct
// tags can be raw strings. Fixtures using it must not need a real '.
func bt(s string) string { return strings.ReplaceAll(s, "'", "`") }

// buildFixture lays out a module with files and main as ./cmd/run,
// generates flagsMode output for type Cfg, builds the program and returns
// the module dir and the binary path.
func buildFixture(t *testing.T, files map[string]string, flagsMode, main string) (dir, bin string) {
	t.Helper()
	all := map[string]string{"cmd/run/main.go": main}
	for k, v := range files {
		all[k] = v
	}
	dir = writeModule(t, all)
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsMode)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	bin = filepath.Join(dir, "run.bin")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", bin, "./cmd/run")
	cmd.Dir = dir
	cmd.Env = hermeticEnv(t)
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, o)
	}
	return dir, bin
}

type runResult struct {
	stdout, stderr string
	code           int
}

func runBin(t *testing.T, dir, bin string, env []string, args ...string) runResult {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), bin, args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	res := runResult{stdout: stdout.String(), stderr: stderr.String()}
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		res.code = exit.ExitCode()
	case err != nil:
		t.Fatal(err)
	}
	return res
}

const (
	cfgJSON      = "cfg.json"
	cfgFileJSON  = "CFG_FILE=cfg.json"
	cfgYAML      = "cfg.yaml"
	overflowJSON = "overflow.json"
)

const secretFixture = `package fixture

type Inner struct {
	Pass *string 'name:"pass" secret:"true" default:"inner-pw" env:"-"'
	Key  string  'name:"key" secret:"true" default:"inner-key"'
	Host string  'name:"host" default:"localhost"'
}

type Cfg struct {
	Token  *string          'name:"token" secret:"true" default:"top-token"'
	Opt    *Inner           'name:"opt"'
	Items  []Inner          'name:"items"'
	ByName map[string]Inner 'name:"by-name"'
	Keys   []string         'name:"keys" secret:"true" default:"k1,k2"'
}

func (Cfg) Validate() error { return nil }
`

func TestSecretDefaultsRedacted(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, bt(secretFixture), false)
	if err != nil {
		t.Fatal(err)
	}
	j, err := emitSampleJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := emitJSONSchema(m)
	if err != nil {
		t.Fatal(err)
	}
	yamlSample := string(emitSample(m))
	outputs := map[string]string{
		"yaml":     yamlSample,
		"json":     string(j),
		"toml":     string(emitSampleTOML(m)),
		"markdown": string(emitMarkdown(m, ".", "APP_", "_", false)),
		"schema":   string(schema),
	}
	for name, out := range outputs {
		for _, secret := range []string{"top-token", "inner-pw", "inner-key", "k1"} {
			if strings.Contains(out, secret) {
				t.Errorf("%s output leaks %q:\n%s", name, secret, out)
			}
		}
	}
	for _, want := range []string{`token: "(secret)"`, `#   pass: "(secret)"`, `#     pass: "(secret)"`, `#   host: "localhost"`} {
		if !strings.Contains(yamlSample, want) {
			t.Errorf("yaml sample missing %q:\n%s", want, yamlSample)
		}
	}
}

func TestSecretsHiddenAtRuntime(t *testing.T) {
	t.Parallel()
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{"cfg.json"}})
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)
	fs.SetOutput(os.Stdout)
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fs.PrintDefaults()
		return
	}
	cfg, err := c.Load()
	if err != nil {
		panic(err)
	}
	fmt.Print(cfg.PrintConfig())
}
`
	files := map[string]string{
		fixtureFile: bt(secretFixture),
		cfgJSON:     `{"opt": {"pass": "file-pw", "key": "file-key"}, "items": [{"pass": "item-pw", "key": "item-key"}], "by-name": {"a": {"key": "map-key"}}}`,
	}
	dir, bin := buildFixture(t, files, flagsPFlag, main)
	env := hermeticEnv(t)
	help := runBin(t, dir, bin, env, "help")
	printed := runBin(t, dir, bin, env)
	for name, res := range map[string]runResult{"help": help, "PrintConfig": printed} {
		if res.code != 0 {
			t.Fatalf("%s: exit %d\n%s", name, res.code, res.stderr)
		}
		for _, secret := range []string{"top-token", "inner-pw", "inner-key", "k1", "file-pw", "file-key", "item-pw", "item-key", "map-key"} {
			if strings.Contains(res.stdout, secret) {
				t.Errorf("%s leaks %q:\n%s", name, secret, res.stdout)
			}
		}
	}
	if !strings.Contains(printed.stdout, "opt.host = localhost") {
		t.Errorf("PrintConfig should still show non-secret fields of an optional struct:\n%s", printed.stdout)
	}
}

func TestSampleFormatsLoad(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Rule struct {
	From int  'name:"from"'
	On   bool 'name:"on" default:"true"'
}

type Peer struct {
	Name  string          'name:"name"'
	Rules []Rule          'name:"rules"'
	Tags  map[string]Rule 'name:"tags"'
	Opt   *Rule           'name:"opt"'
}

type Group struct {
	Labels map[string]string 'name:"labels"'
	Peers  []Peer            'name:"peers"'
	Ports  []string          'name:"ports" default:"1,2"'
}

type Cfg struct {
	Name   string          'name:"name" default:"svc"'
	Peers  []Peer          'name:"peers"'
	ByName map[string]Peer 'name:"by-name"'
	Labels map[string]string 'name:"labels"'
	Sub    Group           'name:"sub"'
	Opt    *Rule           'name:"opt"'
	Nick   *string         'name:"nick"'
}

func (Cfg) Validate() error { return nil }
`)
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	j, err := emitSampleJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/pelletier/go-toml/v2"
	fixture "fixture"
)

func main() {
	c := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{
		Explicit: os.Args[1],
		Decoders: configulator.Decoders{".json": configulator.StrictJSON, ".toml": toml.Unmarshal},
	})
	cfg, err := c.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(cfg.Name, len(cfg.Peers), cfg.Sub.Ports)
}
`
	files := map[string]string{
		fixtureFile:   src,
		"sample.json": string(j),
		"sample.toml": string(emitSampleTOML(m)),
	}
	dir, bin := buildFixture(t, files, flagsNone, main)
	for _, f := range []string{"sample.json", "sample.toml"} {
		res := runBin(t, dir, bin, hermeticEnv(t), f)
		if res.code != 0 || strings.TrimSpace(res.stdout) != "svc 0 [1 2]" {
			t.Errorf("%s does not load: exit %d\n%s%s\n%s", f, res.code, res.stdout, res.stderr, files[f])
		}
	}
}

func TestRequiredComposite(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Store struct {
	URL string 'name:"url"'
}

type TLS struct {
	Cert string 'name:"cert" required:"true"'
	Min  uint16 'name:"min" default:"12"'
	Deep *Store 'name:"deep" env:"-" required:"true"'
}

type Cfg struct {
	DB    Store  'name:"db" required:"true"'
	Cache *Store 'name:"cache" required:"true"'
	TLS   *TLS   'name:"tls"'
}

func (Cfg) Validate() error { return nil }
`)
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	fixture "fixture"
)

func main() {
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{os.Getenv("CFG_FILE")}})
	cfg, err := c.Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("ok", cfg.TLS != nil)
}
`
	files := map[string]string{
		fixtureFile:      src,
		"empty-tls.json": `{"tls": {}}`,
		"deep.json":      `{"tls": {"cert": "c", "deep": {"url": "u"}}}`,
	}
	dir, bin := buildFixture(t, files, flagsNone, main)
	db, cache := "FX_DB_URL=d", "FX_CACHE_URL=c"
	cases := []struct {
		name string
		env  []string
		want string
	}{
		{"struct-unset", []string{cache}, "error: db: required but not set by any layer"},
		{"pointer-struct-unset", []string{db}, "error: cache: required but not set by any layer"},
		{"optional-unset", []string{db, cache}, "ok false"},
		{"optional-allocated-by-env", []string{db, cache, "FX_TLS_MIN=13"}, "error: tls.cert: required but not set by any layer"},
		{"optional-allocated-by-file", []string{db, cache, "CFG_FILE=empty-tls.json"}, "error: tls.cert: required but not set by any layer"},
		{"required-pointer-inside-optional", []string{db, cache, "FX_TLS_CERT=x"}, "error: tls.deep: required but not set by any layer"},
		{"optional-satisfied", []string{db, cache, "CFG_FILE=deep.json"}, "ok true"},
	}
	for _, tc := range cases {
		res := runBin(t, dir, bin, append(hermeticEnv(t), tc.env...))
		if got := strings.TrimSpace(res.stdout); res.code != 0 || got != tc.want {
			t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
		}
	}
}

func TestFlagConflictsAtGenerate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, fields, contains string
	}{
		{"duplicate-short", "\tA string 'name:\"a\" short:\"x\"'\n\tB string 'name:\"b\" short:\"x\"'\n", `short:"x" is also used by a`},
		{"short-h", "\tA string 'name:\"a\" short:\"h\"'\n", `short:"h" is reserved for help`},
		{"help-name", "\tHelp bool 'name:\"help\"'\n", `--help is reserved for help`},
		{"help-flag-override", "\tA bool 'name:\"a\" flag:\"help\"'\n", `--help is reserved for help`},
		{"long-short", "\tA string 'name:\"a\" short:\"ab\"'\n", "must be a single ASCII character"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, err := buildFixtureModel(t, bt("package fixture\n\ntype Cfg struct {\n"+tc.fields+"}\n")+validateStub, false)
			if err != nil {
				t.Fatal(err)
			}
			_, err = emit(m, flagsPFlag)
			if err == nil || !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("want error containing %q, got %v", tc.contains, err)
			}
		})
	}
	m, err := buildFixtureModel(t, bt("package fixture\n\ntype Cfg struct {\n\tHelp bool 'name:\"help\" flag:\"-\"'\n}\n")+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := emit(m, flagsPFlag); err != nil {
		t.Fatalf("flag:\"-\" on a help field should be allowed: %v", err)
	}
	if _, err := emit(m, flagsNone); err != nil {
		t.Fatalf("-flags=none should not check flag names: %v", err)
	}
}

const conflictMainPFlag = `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema())
	o := &cpflag.Options{}
	switch os.Args[1] {
	case "file":
		c.WithFile(&configulator.FileOptions{})
	case "user-name":
		fs.String("port", "", "")
	case "user-short":
		fs.BoolP("zz", "p", false, "")
	case "config-short-taken":
		fs.BoolP("zz", "c", false, "")
		c.WithFile(&configulator.FileOptions{})
	case "dash":
		o.Separator = "-"
	}
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), o)
	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Println("parse:", err)
		return
	}
	cfg, err := c.Load()
	if err != nil {
		fmt.Printf("%T: %v\n", err, err)
		return
	}
	fmt.Printf("%+v\n", *cfg)
}
`

const conflictMainStd = `package main

import (
	"flag"
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
	fixture "fixture"
)

func main() {
	fs := flag.NewFlagSet("x", flag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema())
	o := &cstd.Options{}
	switch os.Args[1] {
	case "user-name":
		o.Separator = "-"
		fs.String("port", "", "")
	case "dash":
		o.Separator = "-"
	}
	cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), o)
	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Println("parse:", err)
		return
	}
	cfg, err := c.Load()
	if err != nil {
		fmt.Printf("%T: %v\n", err, err)
		return
	}
	fmt.Printf("%+v\n", *cfg)
}
`

const dupFixture = `package fixture

type Sub struct {
	B string 'name:"b"'
}

type Cfg struct {
	Port int    'name:"port"'
	A    Sub    'name:"a"'
	AB   string 'name:"a.b"'
}

func (Cfg) Validate() error { return nil }
`

func TestFlagConflictsAtLoad(t *testing.T) {
	t.Parallel()
	shortSrc := bt(`package fixture

type Cfg struct {
	Port    int  'name:"port" short:"p"'
	Verbose bool 'name:"verbose" short:"c"'
}

func (Cfg) Validate() error { return nil }
`)
	const conflict = "*configulator.FlagConflictError: "
	type scenario struct{ args, want string }
	fixtures := []struct {
		name, mode, src, main string
		scenarios             []scenario
	}{
		{"pflag-short", flagsPFlag, shortSrc, conflictMainPFlag, []scenario{
			{"plain -p 9 -c", "{Port:9 Verbose:true}"},
			{"file", conflict + `flag "verbose": shorthand "c" is already used by flag "config"`},
			{"user-name", conflict + `flag "port" is already defined`},
			{"user-short", conflict + `flag "port": shorthand "p" is already used by flag "zz"`},
			{"config-short-taken", conflict + `flag "config": shorthand "c" is already used by flag "zz"`},
		}},
		{"pflag-dup", flagsPFlag, bt(dupFixture), conflictMainPFlag, []scenario{
			{"plain", conflict + `flag "a.b" is already defined`},
			{"dash --a-b=x --a.b=y", "{Port:0 A:{B:x} AB:y}"},
		}},
		{"std-dup", flagsStd, bt(dupFixture), conflictMainStd, []scenario{
			{"plain", conflict + `flag "a.b" is already defined`},
			{"dash -a-b=x -a.b=y", "{Port:0 A:{B:x} AB:y}"},
			{"user-name", conflict + `flag "port" is already defined`},
		}},
	}
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			t.Parallel()
			dir, bin := buildFixture(t, map[string]string{fixtureFile: fx.src}, fx.mode, fx.main)
			for _, sc := range fx.scenarios {
				res := runBin(t, dir, bin, hermeticEnv(t), strings.Fields(sc.args)...)
				if got := strings.TrimSpace(res.stdout); res.code != 0 || !strings.HasPrefix(got, sc.want) {
					t.Errorf("%s: got %q (exit %d), want prefix %q\n%s", sc.args, got, res.code, sc.want, res.stderr)
				}
			}
		})
	}
}

func TestListDefaultsUseSeparator(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Elem struct {
	Name string   'name:"name"'
	Tags []string 'name:"tags" default:"a;b,c"'
	Nums []int    'name:"nums" default:"1;2"'
}

type Cfg struct {
	Tags  []string 'name:"tags" default:"a;b,c"'
	Nums  []int    'name:"nums" default:"1;2"'
	Elems []Elem   'name:"elems"'
	Opt   *Elem    'name:"opt"'
}

func (Cfg) Validate() error { return nil }
`)
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithArraySeparator(";").
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{"cfg.json"}})
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
		return
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
	cfg, err := c.Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	d, err := c.Default()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q %v %q %v %q %v %q\n", cfg.Tags, cfg.Nums, cfg.Elems[0].Tags, cfg.Elems[0].Nums, cfg.Opt.Tags, cfg.Opt.Nums, d.Tags)
}
`
	files := map[string]string{fixtureFile: src, cfgJSON: `{"elems": [{"name": "x"}]}`}
	dir, bin := buildFixture(t, files, flagsPFlag, main)
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"defaults", []string{"FX_OPT_NAME=o"}, nil, `["a" "b,c"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"]`},
		{"env", []string{"FX_TAGS=x;y,z", "FX_OPT_NAME=o"}, nil, `["x" "y,z"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"]`},
		{"flag-allocates-optional", nil, []string{"--opt.name=o"}, `["a" "b,c"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"] [1 2] ["a" "b,c"]`},
	}
	for _, tc := range cases {
		res := runBin(t, dir, bin, append(hermeticEnv(t), tc.env...), tc.args...)
		if got := strings.TrimSpace(res.stdout); res.code != 0 || got != tc.want {
			t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
		}
	}
	help := runBin(t, dir, bin, hermeticEnv(t), "help").stdout
	for _, want := range []string{"(default [a;b,c])", "(default [1;2])"} {
		if !strings.Contains(help, want) {
			t.Errorf("help should show the default as written, missing %q:\n%s", want, help)
		}
	}
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	if sample := string(emitSample(m)); !strings.Contains(sample, `tags: ["a;b", "c"]`) {
		t.Errorf("sample should split the default on commas and quote each element:\n%s", sample)
	}
}

const optionalFixture = `package fixture

import "time"

type Depth struct {
	Level int64 'name:"level" default:"7"'
}

type TLS struct {
	Cert  string        'name:"cert"'
	On    bool          'name:"on" default:"true"'
	Ratio float64       'name:"ratio" default:"0.5"'
	Wait  time.Duration 'name:"wait" default:"5s"'
	Tags  []string      'name:"tags" default:"a,b"'
	Ports []uint16      'name:"ports"'
	Inner Depth         'name:"inner"'
	Deep  *Depth        'name:"deep"'
	Nick  *string       'name:"nick"'
}

type Cfg struct {
	Name string 'name:"name"'
	TLS  *TLS   'name:"tls"'
}

func (Cfg) Validate() error { return nil }
`

const optionalMainBody = `
	cfg, err := c.Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if cfg.TLS == nil {
		fmt.Println("tls unset")
		return
	}
	s := cfg.TLS
	deep, nick := "-", "-"
	if s.Deep != nil {
		deep = fmt.Sprint(s.Deep.Level)
	}
	if s.Nick != nil {
		nick = *s.Nick
	}
	o, _ := c.Report().Origin("tls.wait")
	fmt.Println(s.Cert, s.On, s.Ratio, s.Wait, s.Tags, s.Ports, s.Inner.Level, deep, nick, o.Layer, o.Detail)
}
`

func TestOptionalStructAllLayers(t *testing.T) {
	t.Parallel()
	pflagMain := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"})
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println("parse:", err)
		return
	}
` + optionalMainBody
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(pflagMain)
	const allocated = "true 0.5 5s [a b] [] 7"
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"unset", nil, nil, "tls unset"},
		{"env-leaf", []string{"FX_TLS_CERT=x"}, nil, "x " + allocated + " - - default element default"},
		{"env-nested-optional", []string{"FX_TLS_DEEP_LEVEL=3", "FX_TLS_PORTS=1,2"}, nil, " true 0.5 5s [a b] [1 2] 7 3 - default element default"},
		{"env-bool-float-duration", []string{"FX_TLS_ON=false", "FX_TLS_RATIO=2", "FX_TLS_WAIT=1m"}, nil, " false 2 1m0s [a b] [] 7 - - env FX_TLS_WAIT"},
		{"flag-leaf", nil, []string{"--tls.cert=y", "--tls.nick=n"}, "y " + allocated + " - n default element default"},
		{"flag-nested-optional", nil, []string{"--tls.deep.level=4", "--tls.inner.level=8"}, " true 0.5 5s [a b] [] 8 4 - default element default"},
		{"env-then-flag", []string{"FX_TLS_INNER_LEVEL=9"}, []string{"--tls.cert=z", "--tls.wait=2s"}, "z true 0.5 2s [a b] [] 9 - - cli --tls.wait"},
		{"env-bad-value-allocates-nothing", []string{"FX_TLS_RATIO=nope"}, nil, `error: tls.ratio: cannot parse "nope" from FX_TLS_RATIO: strconv.ParseFloat: parsing "nope": invalid syntax`},
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			main := pflagMain
			if mode == flagsStd {
				main = stdMain
			}
			dir, bin := buildFixture(t, map[string]string{fixtureFile: bt(optionalFixture)}, mode, main)
			for _, tc := range cases {
				args := tc.args
				if mode == flagsStd {
					args = nil
					for _, a := range tc.args {
						args = append(args, strings.TrimPrefix(a, "-"))
					}
				}
				want := strings.ReplaceAll(tc.want, "--tls", map[string]string{flagsPFlag: "--tls", flagsStd: "-tls"}[mode])
				res := runBin(t, dir, bin, append(hermeticEnv(t), tc.env...), args...)
				if got := strings.TrimRight(res.stdout, "\n"); res.code != 0 || got != want {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, want, res.stderr)
				}
			}
		})
	}
}

func TestFlagSkipSubtree(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Store struct {
	URL string 'name:"url"'
}

type Cfg struct {
	DB    Store  'name:"db" flag:"-"'
	Cache *Store 'name:"cache" flag:"-"'
	Keep  Store  'name:"keep"'
}

func (Cfg) Validate() error { return nil }
`)
	main := map[string]string{
		flagsPFlag: `package main

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	cpflag.Bind(configulator.New(fixture.CfgSchema()), fs, fixture.CfgPFlagHooks(), nil)
	fs.VisitAll(func(f *pflag.Flag) { fmt.Println(f.Name) })
}
`,
		flagsStd: `package main

import (
	"flag"
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
	fixture "fixture"
)

func main() {
	fs := flag.NewFlagSet("x", flag.ContinueOnError)
	cstd.Bind(configulator.New(fixture.CfgSchema()), fs, fixture.CfgStdFlagHooks(), nil)
	fs.VisitAll(func(f *flag.Flag) { fmt.Println(f.Name) })
}
`,
	}
	for mode, m := range main {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, bin := buildFixture(t, map[string]string{fixtureFile: src}, mode, m)
			res := runBin(t, dir, bin, hermeticEnv(t))
			if got := strings.TrimSpace(res.stdout); res.code != 0 || got != "keep.url" {
				t.Errorf("got flags %q (exit %d), want only keep.url\n%s", got, res.code, res.stderr)
			}
		})
	}
}

const listFixture = `package fixture

import "time"

type Port uint16

type Level string

type Cfg struct {
	Ints   []int           'name:"ints" default:"1,2"'
	I8     []int8          'name:"i8"'
	Ports  []Port          'name:"ports" default:"80,443"'
	F64    []float64       'name:"f64"'
	F32    []float32       'name:"f32"'
	Bools  []bool          'name:"bools"'
	Durs   []time.Duration 'name:"durs" default:"1s,2m"'
	Levels []Level         'name:"levels"'
	Secret []int           'name:"secret" secret:"true"'
}

func (Cfg) Validate() error { return nil }
`

func TestScalarLists(t *testing.T) {
	t.Parallel()
	pflagMain := `package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{os.Getenv("CFG_FILE")}, Decoders: configulator.Decoders{".json": configulator.StrictJSON, ".yaml": yaml.Unmarshal}})
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println("parse:", err)
		return
	}
	cfg, err := c.Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.Ints, cfg.I8, cfg.Ports, cfg.F64, cfg.F32, cfg.Bools, cfg.Durs, cfg.Levels)
}
`
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(pflagMain)
	files := map[string]string{
		fixtureFile:  bt(listFixture),
		cfgJSON:      `{"i8": [1, -2], "f64": [1.5], "f32": [2.5], "bools": [true, false], "durs": ["3s"], "levels": ["a", "b"], "ports": [8080]}`,
		cfgYAML:      "i8: [1, -2]\nf64: [1.5]\nf32: [2.5]\nbools: [true, false]\ndurs: [3s]\nlevels: [a, b]\nports: [8080]\n",
		overflowJSON: `{"i8": [300]}`,
	}
	const fromFile = "[1 2] [1 -2] [8080] [1.5] [2.5] [true false] [3s] [a b]"
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"defaults", nil, nil, "[1 2] [] [80 443] [] [] [] [1s 2m0s] []"},
		{"json", []string{cfgFileJSON}, nil, fromFile},
		{"yaml", []string{"CFG_FILE=cfg.yaml"}, nil, fromFile},
		{"env", []string{"FX_INTS=3,4", "FX_I8=-1", "FX_PORTS=1", "FX_F64=0.5", "FX_F32=1.25", "FX_BOOLS=true", "FX_DURS=5s", "FX_LEVELS=x,y"}, nil,
			"[3 4] [-1] [1] [0.5] [1.25] [true] [5s] [x y]"},
		{"flags", nil, []string{"--ints=5,6", "--i8=7", "--ports=9", "--f64=2.5", "--f32=0.5", "--bools=false", "--durs=1h", "--levels=z"},
			"[5 6] [7] [9] [2.5] [0.5] [false] [1h0m0s] [z]"},
		{"env-bad-element", []string{"FX_INTS=1,x"}, nil, `error: ints: cannot parse "1,x" from FX_INTS: strconv.ParseInt: parsing "x": invalid syntax`},
		{"env-overflow", []string{"FX_I8=300"}, nil, `error: i8: cannot parse "300" from FX_I8: strconv.ParseInt: parsing "300": value out of range`},
		{"env-secret-redacted", []string{"FX_SECRET=1,hunter2"}, nil, `error: secret: cannot parse "(redacted)" from FX_SECRET: strconv.ParseInt: parsing "hunter2": invalid syntax`},
		{"file-overflow", []string{"CFG_FILE=overflow.json"}, nil, "error: decoding overflow.json"},
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			main := pflagMain
			if mode == flagsStd {
				main = stdMain
			}
			dir, bin := buildFixture(t, files, mode, main)
			for _, tc := range cases {
				args := tc.args
				if mode == flagsStd {
					args = nil
					for _, a := range tc.args {
						args = append(args, strings.TrimPrefix(a, "-"))
					}
				}
				res := runBin(t, dir, bin, append(hermeticEnv(t), tc.env...), args...)
				if got := strings.TrimSpace(res.stdout); res.code != 0 || !strings.HasPrefix(got, tc.want) {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
				}
			}
		})
	}
}

func TestRequiredInElements(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Rule struct {
	From int 'name:"from" required:"true"'
}

type TLS struct {
	Cert string 'name:"cert" required:"true"'
}

type Server struct {
	Addr   string 'name:"addr" required:"true"'
	Weight int    'name:"weight" default:"1" required:"true"'
	Rules  []Rule 'name:"rules"'
	TLS    *TLS   'name:"tls"'
}

type Cfg struct {
	Servers []Server          'name:"servers"'
	ByName  map[string]Server 'name:"by-name"'
}

func (Cfg) Validate() error { return nil }
`)
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	fixture "fixture"
)

func main() {
	_, err := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{Explicit: os.Args[1]}).Load()
	fmt.Println(err)
}
`
	const ok = "<nil>"
	docs := map[string][2]string{
		"none.json":     {`{}`, ok},
		"empty.json":    {`{"servers": [], "by-name": {}}`, ok},
		"ok.json":       {`{"servers": [{"addr": "a", "rules": [{"from": 1}]}], "by-name": {"x": {"addr": "b", "tls": {"cert": "c"}}}}`, ok},
		"list.json":     {`{"servers": [{"addr": "a"}, {"weight": 2}]}`, "servers[1].addr: required but not set by any layer"},
		"map.json":      {`{"by-name": {"b": {"addr": "x"}, "a": {}, "c": {}}}`, "by-name.a.addr: required but not set by any layer"},
		"nested.json":   {`{"servers": [{"addr": "a", "rules": [{"from": 1}, {}]}]}`, "servers[0].rules[1].from: required but not set by any layer"},
		"optional.json": {`{"by-name": {"x.y": {"addr": "b", "tls": {}}}}`, `by-name."x.y".tls.cert: required but not set by any layer`},
	}
	files := map[string]string{fixtureFile: src}
	for name, d := range docs {
		files[name] = d[0]
	}
	dir, bin := buildFixture(t, files, flagsNone, main)
	for name, d := range docs {
		res := runBin(t, dir, bin, hermeticEnv(t), name)
		if got := strings.TrimSpace(res.stdout); res.code != 0 || got != d[1] {
			t.Errorf("%s: got %q (exit %d), want %q\n%s", name, got, res.code, d[1], res.stderr)
		}
	}
}

func TestSchemaStrictnessMatchesLoaders(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Leaf struct {
	A string 'name:"a"'
}

type Cfg struct {
	Top   string          'name:"top"'
	Sub   Leaf            'name:"sub"'
	Opt   *Leaf           'name:"opt"'
	List  []Leaf          'name:"list"'
	ByKey map[string]Leaf 'name:"by-key"'
}

func (Cfg) Validate() error { return nil }
`)
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := emitJSONSchema(m)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(schema), `"additionalProperties": false`); n != 5 {
		t.Fatalf("want additionalProperties:false on all 5 objects, got %d:\n%s", n, schema)
	}
	main := `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
	fixture "fixture"
)

func main() {
	_, err := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{
		Explicit: os.Args[1],
		Decoders: configulator.Decoders{".json": configulator.StrictJSON, ".yaml": yaml.Unmarshal, ".toml": toml.Unmarshal},
	}).Load()
	fmt.Println(err == nil)
}
`
	files := map[string]string{fixtureFile: src}
	unknown := map[string][3]string{
		"top":    {`{"x": 1}`, "x: 1\n", "x = 1\n"},
		"sub":    {`{"sub": {"x": 1}}`, "sub:\n  x: 1\n", "[sub]\nx = 1\n"},
		"opt":    {`{"opt": {"x": 1}}`, "opt:\n  x: 1\n", "[opt]\nx = 1\n"},
		"list":   {`{"list": [{"x": 1}]}`, "list:\n  - x: 1\n", "[[list]]\nx = 1\n"},
		"by-key": {`{"by-key": {"k": {"x": 1}}}`, "by-key:\n  k:\n    x: 1\n", "[by-key.k]\nx = 1\n"},
	}
	for name, docs := range unknown {
		files[name+".json"], files[name+".yaml"], files[name+".toml"] = docs[0], docs[1], docs[2]
	}
	dir, bin := buildFixture(t, files, flagsNone, main)
	for name := range unknown {
		for ext, accepted := range map[string]string{".json": "false", ".yaml": "true", ".toml": "true"} {
			res := runBin(t, dir, bin, hermeticEnv(t), name+ext)
			if got := strings.TrimSpace(res.stdout); got != accepted {
				t.Errorf("unknown key in %s%s: loaded=%s, want %s\n%s", name, ext, got, accepted, res.stderr)
			}
		}
	}
}

func TestHelpIsNotALoadFailure(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

type Cfg struct {
	Port  int    'name:"port" default:"8080" description:"listen port"'
	Token string 'name:"token" secret:"true" default:"s3cret"'
}

func (Cfg) Validate() error { return nil }
`)
	cobraMain := `package main

import (
	"context"
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/cobra"
	fixture "fixture"
)

func main() {
	cmd := &cobra.Command{Use: "app", Version: "1.0.0", RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := configulator.FromContext[fixture.Cfg](cmd.Context())
		if err != nil {
			return err
		}
		cfg, err := c.Load()
		if err != nil {
			return err
		}
		fmt.Println("ran", cfg.Port)
		return nil
	}}
	c := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{Search: []string{"cfg.json"}})
	cpflag.Bind(c, cmd.Flags(), fixture.CfgPFlagHooks(), nil)
	cmd.SetContext(c.WithContext(context.Background()))
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}
}
`
	stdMain := `package main

import (
	"flag"
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
	fixture "fixture"
)

func main() {
	mode := flag.ExitOnError
	if os.Getenv("CONTINUE") != "" {
		mode = flag.ContinueOnError
	}
	fs := flag.NewFlagSet("app", mode)
	c := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{Search: []string{"cfg.json"}})
	cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)
	perr := fs.Parse(os.Args[1:])
	cfg, err := c.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}
	fmt.Println("parse:", perr, "load ok", cfg.Port)
}
`
	t.Run("cobra", func(t *testing.T) {
		t.Parallel()
		dir, bin := buildFixture(t, map[string]string{fixtureFile: src}, flagsPFlag, cobraMain)
		for _, arg := range []string{"--help", "-h", "--version"} {
			res := runBin(t, dir, bin, hermeticEnv(t), arg)
			if res.code != 0 || strings.Contains(res.stderr, "FATAL") || strings.Contains(res.stdout, "ran") {
				t.Errorf("%s: exit %d\n%s%s", arg, res.code, res.stdout, res.stderr)
			}
			if arg != "--version" && (!strings.Contains(res.stdout, "--port int") || !strings.Contains(res.stdout, "--config")) {
				t.Errorf("%s: help is missing the config flags:\n%s", arg, res.stdout)
			}
			if strings.Contains(res.stdout, "s3cret") {
				t.Errorf("%s: help shows a secret default:\n%s", arg, res.stdout)
			}
		}
	})
	t.Run("std", func(t *testing.T) {
		t.Parallel()
		dir, bin := buildFixture(t, map[string]string{fixtureFile: src}, flagsStd, stdMain)
		res := runBin(t, dir, bin, hermeticEnv(t), "-help")
		if res.code != 0 || !strings.Contains(res.stderr, "-port") || strings.Contains(res.stderr, "FATAL") {
			t.Errorf("-help with ExitOnError: exit %d\n%s%s", res.code, res.stdout, res.stderr)
		}
		if strings.Contains(res.stderr, "s3cret") {
			t.Errorf("-help shows a secret default:\n%s", res.stderr)
		}
		res = runBin(t, dir, bin, append(hermeticEnv(t), "CONTINUE=1"), "-help")
		if res.code != 0 || strings.TrimSpace(res.stdout) != "parse: flag: help requested load ok 8080" {
			t.Errorf("-help with ContinueOnError: exit %d\n%s%s", res.code, res.stdout, res.stderr)
		}
	})
}

func TestScalarMaps(t *testing.T) {
	t.Parallel()
	src := bt(`package fixture

import (
	"net"
	"strings"
	"time"
)

type Level string

type Port uint16

type Code [2]byte

func (c *Code) UnmarshalText(b []byte) error {
	copy(c[:], strings.ToUpper(string(b)))
	return nil
}

type Cfg struct {
	Levels map[string]Level         'name:"levels"'
	Ports  map[string]Port          'name:"ports"'
	Waits  map[string]time.Duration 'name:"waits"'
	Nets   map[string]net.IPNet     'name:"nets"'
	Codes  map[string]Code          'name:"codes"'
	Small  map[string]int8          'name:"small"'
}

func (Cfg) Validate() error { return nil }
`)
	main := `package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	configulator "github.com/USA-RedDragon/configulator/v2"
	fixture "fixture"
)

func main() {
	cfg, err := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{
		Explicit: os.Args[1],
		Decoders: configulator.Decoders{".json": configulator.StrictJSON, ".yaml": yaml.Unmarshal},
	}).Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	n, c := cfg.Nets["a"], cfg.Codes["a"]
	fmt.Println(cfg.Levels, cfg.Ports, cfg.Waits, n.String(), string(c[:]), cfg.Small)
}
`
	files := map[string]string{
		fixtureFile:   src,
		"cfg.json":    `{"levels": {"a": "debug"}, "ports": {"a": 80}, "waits": {"a": "3s"}, "nets": {"a": "10.0.0.0/8"}, "codes": {"a": "ab"}, "small": {"a": -3}}`,
		cfgYAML:       "levels: {a: debug}\nports: {a: 80}\nwaits: {a: 3s}\nnets: {a: 10.0.0.0/8}\ncodes: {a: ab}\nsmall: {a: -3}\n",
		overflowJSON:  `{"small": {"a": 300}}`,
		"opaque.json": `{"waits": {"a": 3}}`,
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, bin := buildFixture(t, files, mode, main)
			want := map[string]string{
				"cfg.json":    "map[a:debug] map[a:80] map[a:3s] 10.0.0.0/8 AB map[a:-3]",
				cfgYAML:       "map[a:debug] map[a:80] map[a:3s] 10.0.0.0/8 AB map[a:-3]",
				overflowJSON:  "error: decoding overflow.json",
				"opaque.json": "error: decoding opaque.json",
			}
			for file, w := range want {
				res := runBin(t, dir, bin, hermeticEnv(t), file)
				if got := strings.TrimSpace(res.stdout); res.code != 0 || !strings.HasPrefix(got, w) {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", file, got, res.code, w, res.stderr)
				}
			}
		})
	}
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	if md := string(emitMarkdown(m, ".", "APP_", "_", false)); !strings.Contains(md, "map of string") || !strings.Contains(md, "map of integer") {
		t.Errorf("markdown:\n%s", md)
	}
	for _, unsupported := range []string{"map[string][]string", "map[string]*int", "map[string]map[string]string"} {
		_, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n\tM "+unsupported+" `name:\"m\"`\n}\n"+validateStub, false)
		if err == nil || !strings.Contains(err.Error(), "map of") || !strings.Contains(err.Error(), "is not supported") {
			t.Errorf("%s: want a generate-time error, got %v", unsupported, err)
		}
	}
}
