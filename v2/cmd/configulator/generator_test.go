package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureFile = "cfg.go"

// v2Path returns the local configulator/v2 tree; go test runs in this
// package's directory, two levels below it.
func v2Path(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func hermeticEnv(t *testing.T) []string {
	t.Helper()
	out, err := exec.CommandContext(t.Context(), "go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	cache := strings.TrimSpace(string(out)) + "/cache/download"
	return append(os.Environ(),
		"GOWORK=off", "GOFLAGS=-mod=mod", "GOPROXY=file://"+cache, "GOPRIVATE=*",
	)
}

// writeModule lays out a temp module whose go.mod replaces configulator/v2
// with the local tree, returning its dir.
func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	gomod := fmt.Sprintf(`module fixture

go 1.27

require github.com/USA-RedDragon/configulator/v2 v2.0.0

replace github.com/USA-RedDragon/configulator/v2 => %s
`, v2Path(t))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = hermeticEnv(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	return dir
}

// buildFixtureModel builds the model for type Cfg declared in src.
func buildFixtureModel(t *testing.T, src string, noValidate bool) (*Model, error) {
	t.Helper()
	dir := writeModule(t, map[string]string{fixtureFile: src})
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
	if err != nil {
		return nil, err
	}
	return buildModel(named, outPkg, noValidate)
}

const validateStub = "func (Cfg) Validate() error { return nil }\n"

func TestGenerateTimeErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		src      string
		contains string
	}{
		{"sibling-case-fold-collision",
			"package fixture\n\ntype Cfg struct {\n\tA string `name:\"foo\"`\n\tB string `name:\"FOO\"`\n}\n" + validateStub,
			"collide under env-name folding"},
		{"sibling-dash-underscore-collision",
			"package fixture\n\ntype Cfg struct {\n\tA string `name:\"a-b\"`\n\tB string `name:\"a_b\"`\n}\n" + validateStub,
			"collide under env-name folding"},
		{"unexported-with-name-tag",
			"package fixture\n\ntype Cfg struct {\n\tx string `name:\"x\"`\n}\n" + validateStub,
			"unexported field carries a name: tag"},
		{"unparseable-default",
			"package fixture\n\ntype Cfg struct {\n\tA int `name:\"a\" default:\"zzz\"`\n}\n" + validateStub,
			"default:"},
		{"default-on-map",
			"package fixture\n\ntype Cfg struct {\n\tA map[string]string `name:\"a\" default:\"x\"`\n}\n" + validateStub,
			"default: on a map is not supported"},
		{"env-optin-on-collection",
			"package fixture\n\ntype S struct {\n\tA string `name:\"a\"`\n}\n\ntype Cfg struct {\n\tXs []S `name:\"xs\" env:\"XS\"`\n}\n" + validateStub,
			"opt-in on a list of structs"},
		{"generic-config",
			"package fixture\n\ntype Cfg[T any] struct {\n\tA string `name:\"a\"`\n}\n\nfunc (Cfg[T]) Validate() error { return nil }\n",
			"generic config types"},
		{"missing-validate",
			"package fixture\n\ntype Cfg struct {\n\tA string `name:\"a\"`\n}\n",
			"no Validate() error method"},
		{"struct-kind-textunmarshaler",
			"package fixture\n\ntype Endpoint struct {\n\tHost string\n\tPort int\n}\n\nfunc (e *Endpoint) UnmarshalText(b []byte) error { return nil }\n\ntype Cfg struct {\n\tEp Endpoint `name:\"ep\"`\n}\n" + validateStub,
			"has no built-in wrapper"},
		{"bad-stdlib-default",
			"package fixture\n\nimport \"net\"\n\ntype Cfg struct {\n\tN net.IPNet `name:\"n\" default:\"nope\"`\n}\n" + validateStub,
			"invalid CIDR address"},
		{"bad-duration-default",
			"package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n\tD time.Duration `name:\"d\" default:\"30x\"`\n}\n" + validateStub,
			"unknown unit"},
		{"pointer-to-duration",
			"package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n\tD *time.Duration `name:\"d\"`\n}\n" + validateStub,
			"pointer to time.Duration is not supported"},
		{"anonymous-struct",
			"package fixture\n\ntype Cfg struct {\n\tHTTP struct {\n\t\tPort int `name:\"port\"`\n\t} `name:\"http\"`\n}\n" + validateStub,
			"anonymous struct types are not supported"},
		{"location-by-value",
			"package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n\tL time.Location `name:\"l\"`\n}\n" + validateStub,
			"use *time.Location"},
		{"map-non-string-keys",
			"package fixture\n\ntype Cfg struct {\n\tA map[int]string `name:\"a\"`\n}\n" + validateStub,
			"map keys must be strings"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := buildFixtureModel(t, tc.src, false)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.contains)
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("error %q does not contain %q", err, tc.contains)
			}
		})
	}
}

func TestInternalPackageSameTree(t *testing.T) {
	t.Parallel()
	// A config can use internal packages from its own tree, and the
	// generated file is in the same package so it can too.
	dir := writeModule(t, map[string]string{
		"liba/internal/secret/secret.go": "package secret\n\ntype Options struct {\n\tMode string `name:\"mode\"`\n}\n",
		"liba/cfg.go":                    "package liba\n\nimport \"fixture/liba/internal/secret\"\n\ntype Cfg struct {\n\tOpts secret.Options `name:\"opts\"`\n}\n\nfunc (Cfg) Validate() error { return nil }\n",
	})
	named, outPkg, err := loadPackage(filepath.Join(dir, "liba"), "Cfg", hermeticEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsNone)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "liba", "cfg_configulator.go"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = hermeticEnv(t)
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("same-tree internal output does not compile: %v\n%s", err, o)
	}
}

func TestNoValidateFlag(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n\tA string `name:\"a\"`\n}\n", true)
	if err != nil {
		t.Fatalf("-no-validate should permit a Validate-less config: %v", err)
	}
	if m.HasValidate {
		t.Fatal("HasValidate should be false")
	}
}

func TestURLBecomesSlot(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\nimport \"net/url\"\n\ntype Cfg struct {\n\tU url.URL `name:\"u\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatalf("url.URL should map to the URL sentinel slot: %v", err)
	}
	if m.Fields[0].Kind != KindStdSlot || m.Fields[0].SlotType != "URL" {
		t.Fatalf("url.URL classified as %v/%s, want StdSlot/URL", m.Fields[0].Kind, m.Fields[0].SlotType)
	}
}

// TestStdFlagsOutputCompiles generates -flags=std output into a hermetic
// module and builds it.
func TestStdFlagsOutputCompiles(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n\tName string `name:\"name\" default:\"x\"`\n\tPort uint16 `name:\"port\" default:\"8080\"`\n\tWait time.Duration `name:\"wait\" default:\"5s\"`\n\tOpt *int64 `name:\"opt\"`\n\tTags []string `name:\"tags\"`\n}\n" + validateStub
	dir := writeModule(t, map[string]string{fixtureFile: src})
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsStd)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = hermeticEnv(t)
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("std output does not compile: %v\n%s", err, o)
	}
	if p := findShortTag([]*Field{{Tag: "s", Short: "s"}}, ""); p == "" {
		t.Fatal("findShortTag missed a short: tag")
	}
}

func TestShortUnderStdRejected(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Cfg struct {\n\tA string `name:\"a\" short:\"a\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	if p := findShortTag(m.Fields, ""); p != "a" {
		t.Fatalf("findShortTag = %q, want \"a\"", p)
	}
}

func TestSampleFormats(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Inner struct {\n\tHost string `name:\"host\" default:\"localhost\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Inner  `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}

	j, err := emitSampleJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"port": 8080`, `"host": "localhost"`, `"tags": [`, `"a",`} {
		if !strings.Contains(string(j), want) {
			t.Errorf("json sample missing %q:\n%s", want, j)
		}
	}
	if strings.Contains(string(j), `"key"`) {
		t.Errorf("json sample must leave out the secret key:\n%s", j)
	}

	tm := string(emitSampleTOML(m))
	for _, want := range []string{"port = 8080", "[sub]", `host = "localhost"`, `tags = ["a", "b"]`} {
		if !strings.Contains(tm, want) {
			t.Errorf("toml sample missing %q:\n%s", want, tm)
		}
	}
	if strings.Contains(tm, "key =") {
		t.Errorf("toml sample must leave out the secret key:\n%s", tm)
	}
	if strings.Index(tm, "[sub]") < strings.Index(tm, "tags =") {
		t.Errorf("toml scalars must precede tables:\n%s", tm)
	}
}

func TestMarkdown(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Inner struct {\n\tHost string `name:\"host\" default:\"localhost\" description:\"bind host\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\" required:\"true\" description:\"listen port\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Inner  `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	md := string(emitMarkdown(m, ".", "APP_", "_", true))
	// Cells are padded, so collapse runs of spaces before checking content.
	lines := strings.Split(md, "\n")
	squeezed := make([]string, 0, len(lines))
	for _, line := range lines {
		squeezed = append(squeezed, strings.Join(strings.Fields(line), " "))
	}
	md = strings.Join(squeezed, "\n")
	for _, want := range []string{
		"| Key | Type | Default | Environment | Flag | Description |",
		"| `port` | integer | `8080` | `APP_PORT` | `--port` | listen port (required) |",
		"| `key` | string | | `APP_KEY` | `--key` | secret |",
		"| `sub.host` | string | `localhost` | `APP_SUB_HOST` | `--sub.host` | bind host |",
		"| `tags` | list of string | `a,b` | `APP_TAGS` | `--tags` | |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}
}

func TestEnvFlagOverrides(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Cfg struct {\n" +
		"\tHost     string `name:\"host\" env:\"HOSTNAME_OVERRIDE\" flag:\"hostname\"`\n" +
		"\tInternal string `name:\"internal\" env:\"-\" flag:\"-\"`\n" +
		"\tPort     uint16 `name:\"port\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsPFlag)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	if !strings.Contains(code, `"HOSTNAME_OVERRIDE"`) {
		t.Errorf("env override segment not emitted:\n%s", code)
	}
	if strings.Contains(code, `"HOST"`) {
		t.Errorf("derived env segment emitted despite override:\n%s", code)
	}
	if !strings.Contains(code, `"hostname"`) {
		t.Errorf("flag override segment not emitted:\n%s", code)
	}
	if strings.Contains(code, `"INTERNAL"`) {
		t.Errorf("env:\"-\" field still read from env:\n%s", code)
	}
}

func TestBadEnvOverrideRejected(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Cfg struct {\n" +
		"\tHost string `name:\"host\" env:\"lower-case\"`\n}\n" + validateStub
	_, err := buildFixtureModel(t, src, false)
	if err == nil || !strings.Contains(err.Error(), "uppercase") {
		t.Fatalf("want uppercase-override error, got %v", err)
	}
}

func TestSchemaAndSample(t *testing.T) {
	t.Parallel()
	src := "package fixture\n\ntype Inner struct {\n\tHost string `name:\"host\" default:\"localhost\" description:\"bind host\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\" required:\"true\" description:\"listen port\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Inner  `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, src, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := emitJSONSchema(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"required"`, `"port"`, `"listen port"`, `"additionalProperties": false`, `"default": 8080`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("schema missing %q:\n%s", want, b)
		}
	}
	sm := string(emitSample(m))
	for _, want := range []string{"port: 8080", `key: "(secret)"`, "# bind host", `host: "localhost"`, "tags: [a,b]"} {
		if !strings.Contains(sm, want) {
			t.Errorf("sample missing %q:\n%s", want, sm)
		}
	}
}

// TestNamedScalarTypes covers `type LogLevel string` style fields (same
// package and imported), with defaults and nesting. The generator used to
// assign the decoded basic value directly to the named field, which
// doesn't compile.
func TestNamedScalarTypes(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"kinds/kinds.go": "package kinds\n\ntype Storage string\n\nconst (\n\tStoragePostgres Storage = \"postgres\"\n\tStorageSQLite Storage = \"sqlite\"\n)\n",
		fixtureFile: "package fixture\n\nimport \"fixture/kinds\"\n\n" +
			"type LogLevel string\n\ntype RetryCount int32\n\ntype Switch bool\n\ntype Fraction float64\n\n" +
			"type Cfg struct {\n" +
			"\tLevel   LogLevel      `name:\"level\" default:\"info\"`\n" +
			"\tRetries RetryCount    `name:\"retries\" default:\"3\"`\n" +
			"\tToggle  Switch        `name:\"toggle\"`\n" +
			"\tRatio   Fraction      `name:\"ratio\" default:\"0.5\"`\n" +
			"\tStore   StoreConfig   `name:\"store\"`\n" +
			"}\n\n" +
			"type StoreConfig struct {\n\tType kinds.Storage `name:\"type\" default:\"sqlite\"`\n}\n" + validateStub,
	}
	for _, mode := range []string{flagsPFlag, flagsStd, flagsNone} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir := writeModule(t, files)
			named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
			if err != nil {
				t.Fatal(err)
			}
			m, err := buildModel(named, outPkg, false)
			if err != nil {
				t.Fatal(err)
			}
			out, err := emit(m, mode)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
				t.Fatal(err)
			}
			main := "package main\n\nimport (\n\t\"fmt\"\n\n\tconfigulator \"github.com/USA-RedDragon/configulator/v2\"\n\tfixture \"fixture\"\n)\n\n" +
				"func main() {\n\tcfg, err := configulator.New(fixture.CfgSchema()).\n" +
				"\t\tWithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: \"FX_\", Separator: \"_\"}).\n" +
				"\t\tWithFile(&configulator.FileOptions{Search: []string{\"cfg.json\"}}).\n\t\tLoad()\n" +
				"\tif err != nil {\n\t\tpanic(err)\n\t}\n" +
				"\tfmt.Printf(\"%s %d %t %g %s\\n\", cfg.Level, cfg.Retries, cfg.Toggle, cfg.Ratio, cfg.Store.Type)\n}\n"
			if err := os.MkdirAll(filepath.Join(dir, "cmd", "run"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "cmd", "run", "main.go"), []byte(main), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "cfg.json"), []byte(`{"retries": 7, "store": {"type": "postgres"}}`), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(t.Context(), "go", "run", "./cmd/run")
			cmd.Dir = dir
			cmd.Env = append(hermeticEnv(t), "FX_LEVEL=debug", "FX_TOGGLE=true")
			var stderr strings.Builder
			cmd.Stderr = &stderr
			o, err := cmd.Output()
			if err != nil {
				t.Fatalf("named scalar output does not build/run under -flags=%s: %v\n%s", mode, err, stderr.String())
			}
			if got, want := strings.TrimSpace(string(o)), "debug 7 true 0.5 postgres"; got != want {
				t.Fatalf("-flags=%s: got %q, want %q", mode, got, want)
			}
		})
	}
}

func TestStdlibTypes(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		fixtureFile: "package fixture\n\nimport (\n\t\"net\"\n\t\"net/url\"\n\t\"os\"\n\t\"time\"\n)\n\n" +
			"type Cfg struct {\n" +
			"\tTZ    *time.Location   `name:\"tz\" default:\"America/Chicago\"`\n" +
			"\tMonth time.Month       `name:\"month\" default:\"March\"`\n" +
			"\tNet   net.IPNet        `name:\"net\" default:\"10.0.0.0/8\"`\n" +
			"\tTCP   net.TCPAddr      `name:\"tcp\" default:\"127.0.0.1:80\"`\n" +
			"\tUDP   net.UDPAddr      `name:\"udp\"`\n" +
			"\tMAC   net.HardwareAddr `name:\"mac\" default:\"aa:bb:cc:dd:ee:ff\"`\n" +
			"\tURL   url.URL          `name:\"url\" default:\"https://a.example/x\"`\n" +
			"\tMode  os.FileMode      `name:\"mode\" default:\"0644\"`\n" +
			"\tPort  *int             `name:\"port\" default:\"5\"`\n" +
			"\tOn    *bool            `name:\"on\" default:\"true\"`\n" +
			"}\n" + validateStub,
	}
	binds := map[string][2]string{
		flagsPFlag: {"cpflag \"github.com/USA-RedDragon/configulator/v2/flags/pflag\"\n\t\"github.com/spf13/pflag\"",
			"fs := pflag.NewFlagSet(\"x\", pflag.ContinueOnError)\n\tcpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)\n\t_ = fs.Parse(os.Args[1:])"},
		flagsStd: {"cstd \"github.com/USA-RedDragon/configulator/v2/flags/std\"\n\t\"flag\"",
			"fs := flag.NewFlagSet(\"x\", flag.ContinueOnError)\n\tcstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)\n\t_ = fs.Parse(os.Args[1:])"},
		flagsNone: {"", ""},
	}
	for mode, bind := range binds {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir := writeModule(t, files)
			named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
			if err != nil {
				t.Fatal(err)
			}
			m, err := buildModel(named, outPkg, false)
			if err != nil {
				t.Fatal(err)
			}
			out, err := emit(m, mode)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
				t.Fatal(err)
			}
			main := "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\n\tconfigulator \"github.com/USA-RedDragon/configulator/v2\"\n\tfixture \"fixture\"\n\t" + bind[0] + "\n)\n\n" +
				"var _ = os.Args\n\n" +
				"func main() {\n\tc := configulator.New(fixture.CfgSchema()).\n" +
				"\t\tWithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: \"FX_\", Separator: \"_\"})\n\t" + bind[1] + "\n" +
				"\tcfg, err := c.Load()\n\tif err != nil {\n\t\tpanic(err)\n\t}\n" +
				"\tfmt.Println(cfg.TZ, cfg.Month, cfg.Net.String(), cfg.TCP.String(), cfg.UDP.String(), cfg.MAC, cfg.URL.String(), cfg.Mode, *cfg.Port, *cfg.On)\n}\n"
			if err := os.MkdirAll(filepath.Join(dir, "cmd", "run"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "cmd", "run", "main.go"), []byte(main), 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"run", "./cmd/run"}
			want := "Europe/Paris July 10.0.0.0/8 127.0.0.1:80 127.0.0.1:53 aa:bb:cc:dd:ee:ff https://a.example/x -rw-r--r-- 5 true"
			if mode != flagsNone {
				dash := "-"
				if mode == flagsPFlag {
					dash = "--"
				}
				args = append(args, dash+"tcp=127.0.0.1:443", dash+"mode=0600")
				want = "Europe/Paris July 10.0.0.0/8 127.0.0.1:443 127.0.0.1:53 aa:bb:cc:dd:ee:ff https://a.example/x -rw------- 5 true"
			}
			cmd := exec.CommandContext(t.Context(), "go", args...)
			cmd.Dir = dir
			cmd.Env = append(hermeticEnv(t), "FX_TZ=Europe/Paris", "FX_MONTH=7", "FX_UDP=127.0.0.1:53")
			var stderr strings.Builder
			cmd.Stderr = &stderr
			o, err := cmd.Output()
			if err != nil {
				t.Fatalf("-flags=%s: %v\n%s", mode, err, stderr.String())
			}
			if got := strings.TrimSpace(string(o)); got != want {
				t.Fatalf("-flags=%s:\n got %q\nwant %q", mode, got, want)
			}
		})
	}
}

func TestNestedCollections(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		fixtureFile: "package fixture\n\n" +
			"type Rule struct {\n" +
			"\tFrom  int  `name:\"from\"`\n" +
			"\tRange int  `name:\"range\" default:\"1\"`\n" +
			"\tOn    bool `name:\"on\" default:\"true\"`\n" +
			"}\n\n" +
			"type Level struct {\n\tLevel int `name:\"level\" default:\"7\"`\n}\n\n" +
			"type Peer struct {\n" +
			"\tName  string          `name:\"name\"`\n" +
			"\tSlots int             `name:\"slots\" default:\"3\"`\n" +
			"\tRules []Rule          `name:\"rules\"`\n" +
			"\tTags  map[string]Rule `name:\"tags\"`\n" +
			"\tInner Level           `name:\"inner\"`\n" +
			"\tOpt   *Level          `name:\"opt\"`\n" +
			"}\n\n" +
			"type Cfg struct {\n" +
			"\tPeers  []Peer          `name:\"peers\"`\n" +
			"\tByName map[string]Peer `name:\"by-name\"`\n" +
			"}\n" + validateStub,
	}
	dir := writeModule(t, files)
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsNone)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	main := "package main\n\nimport (\n\t\"fmt\"\n\n\tconfigulator \"github.com/USA-RedDragon/configulator/v2\"\n\tfixture \"fixture\"\n)\n\n" +
		"func main() {\n\tc := configulator.New(fixture.CfgSchema()).WithFile(&configulator.FileOptions{Search: []string{\"cfg.json\"}})\n" +
		"\tcfg, err := c.Load()\n\tif err != nil {\n\t\tpanic(err)\n\t}\n" +
		"\tp := cfg.Peers[0]\n" +
		"\tfmt.Println(p.Name, p.Slots, p.Rules, p.Tags[\"x\"], p.Inner.Level, p.Opt.Level, cfg.ByName[\"b\"].Slots, cfg.ByName[\"b\"].Rules)\n" +
		"\tfor _, path := range []string{\"peers[0].rules[0].range\", \"peers[0].rules[1].range\", \"by-name.b.rules[0].on\", \"peers[0].opt.level\"} {\n" +
		"\t\to, _ := c.Report().Origin(path)\n\t\tfmt.Println(path, o.Layer)\n\t}\n}\n"
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "run"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "run", "main.go"), []byte(main), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := `{"peers": [{"name": "a", "rules": [{"from": 1}, {"from": 2, "range": 5, "on": false}], "tags": {"x": {"from": 9}}, "opt": {}}],` +
		` "by-name": {"b": {"name": "b", "rules": [{"from": 4}]}}}`
	if err := os.WriteFile(filepath.Join(dir, "cfg.json"), []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "go", "run", "./cmd/run")
	cmd.Dir = dir
	cmd.Env = hermeticEnv(t)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	o, err := cmd.Output()
	if err != nil {
		t.Fatalf("nested collections do not build/run: %v\n%s", err, stderr.String())
	}
	want := "a 3 [{1 1 true} {2 5 false}] {9 1 true} 7 7 3 [{4 1 true}]\n" +
		"peers[0].rules[0].range default\n" +
		"peers[0].rules[1].range file\n" +
		"by-name.b.rules[0].on default\n" +
		"peers[0].opt.level default\n"
	if got := string(o); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestMarkdownSkippedSubtree(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\n"+
		"type Store struct {\n\tURL string `name:\"url\"`\n}\n\n"+
		"type Cfg struct {\n"+
		"\tDB    Store  `name:\"db\" env:\"-\"`\n"+
		"\tCache *Store `name:\"cache\" flag:\"-\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	var squeezed []string
	for line := range strings.SplitSeq(string(emitMarkdown(m, ".", "APP_", "_", false)), "\n") {
		squeezed = append(squeezed, strings.Join(strings.Fields(line), " "))
	}
	md := strings.Join(squeezed, "\n")
	for _, want := range []string{
		"| `db.url` | string | | \u2014 | `--db.url` | |",
		"| `cache.url` | string | | `APP_CACHE_URL` | \u2014 | |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}
}

func TestSampleCollectionExamples(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\n"+
		"type Rule struct {\n\tFrom int `name:\"from\"`\n\tOn bool `name:\"on\" default:\"true\"`\n}\n\n"+
		"type Peer struct {\n"+
		"\tName  string          `name:\"name\"`\n"+
		"\tRules []Rule          `name:\"rules\"`\n"+
		"\tTags  map[string]Rule `name:\"tags\"`\n"+
		"}\n\n"+
		"type Cfg struct {\n"+
		"\tPeers []Peer            `name:\"peers\"`\n"+
		"\tLabels map[string]string `name:\"labels\"`\n"+
		"\tOpt   *Rule             `name:\"opt\"`\n"+
		"\tNick  *string           `name:\"nick\" default:\"x\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Sample configuration for Cfg.\n" +
		"# peers:\n" +
		"#   - name: \"\"\n" +
		"#     rules:\n" +
		"#       - from: 0\n" +
		"#         on: true\n" +
		"#     tags:\n" +
		"#       example:\n" +
		"#         from: 0\n" +
		"#         on: true\n" +
		"# labels:\n" +
		"#   example: \"\"\n" +
		"# opt:\n" +
		"#   from: 0\n" +
		"#   on: true\n" +
		"nick: \"x\"\n"
	if got := string(emitSample(m)); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestMarkdownShorthand(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n\tAddr string `name:\"address\" short:\"a\"`\n\tPort int `name:\"port\"`\n}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	var squeezed []string
	for line := range strings.SplitSeq(string(emitMarkdown(m, ".", "", "_", false)), "\n") {
		squeezed = append(squeezed, strings.Join(strings.Fields(line), " "))
	}
	md := strings.Join(squeezed, "\n")
	for _, want := range []string{
		"| `address` | string | | `ADDRESS` | `-a`, `--address` | |",
		"| `port` | integer | | `PORT` | `--port` | |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}
}

func TestSampleSecretWithDefaultIsCommented(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n"+
		"\tDSN string `name:\"dsn\" secret:\"true\" default:\"file:app.db\"`\n"+
		"\tPtr *string `name:\"ptr\" secret:\"true\" default:\"x\"`\n"+
		"\tOn bool `name:\"on\" secret:\"true\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Sample configuration for Cfg.\n# dsn: \"(secret)\"\n# ptr: \"(secret)\"\n# on: \"(secret)\"\n"
	if got := string(emitSample(m)); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSampleListDefaultsQuoted(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n"+
		"\tOrigins []string `name:\"origins\" default:\"*,https://*\"`\n"+
		"\tPorts []int `name:\"ports\" default:\"80,443\"`\n"+
		"\tWaits []time.Duration `name:\"waits\" default:\"1s,2m\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Sample configuration for Cfg.\n" +
		"origins: [\"*\", \"https://*\"]\n" +
		"ports: [80, 443]\n" +
		"waits: [\"1s\", \"2m\"]\n"
	if got := string(emitSample(m)); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
