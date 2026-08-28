package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const v2Path = "/home/reddragon/projects/configulator/v2"

func hermeticEnv() []string {
	cache := os.Getenv("HOME") + "/go/pkg/mod/cache/download"
	return append(os.Environ(),
		"GOWORK=off", "GOFLAGS=-mod=mod", "GOPROXY=file://"+cache, "GOPRIVATE=*",
	)
}

// writeModule lays out a temp module whose go.mod replaces configulator/v2
// with the local tree, returning its dir.
func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	files["go.mod"] = fmt.Sprintf(`module fixture

go 1.27

require github.com/USA-RedDragon/configulator/v2 v2.0.0

replace github.com/USA-RedDragon/configulator/v2 => %s
`, v2Path)
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// resolve go.sum for the replace'd dep graph
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = hermeticEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	return dir
}

func buildFixtureModel(t *testing.T, files map[string]string, typeName string, noValidate bool) (*Model, error) {
	t.Helper()
	dir := writeModule(t, files)
	named, outPkg, err := loadPackage(dir, typeName, hermeticEnv())
	if err != nil {
		return nil, err
	}
	return buildModel(named, outPkg, noValidate)
}

const validateStub = "func (Cfg) Validate() error { return nil }\n"

func TestGenerateTimeErrors(t *testing.T) {
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
			"needs a sentinel slot"},
		{"map-non-string-keys",
			"package fixture\n\ntype Cfg struct {\n\tA map[int]string `name:\"a\"`\n}\n" + validateStub,
			"map keys must be strings"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildFixtureModel(t, map[string]string{"cfg.go": tc.src}, "Cfg", false)
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
	// A config inside a tree may reference its own internal packages;
	// generated output lands in the same package and shares that right.
	// (The cross-tree case cannot exist: user code importing an internal
	// package illegally never compiles to begin with.)
	dir := writeModule(t, map[string]string{
		"liba/internal/secret/secret.go": "package secret\n\ntype Options struct {\n\tMode string `name:\"mode\"`\n}\n",
		"liba/cfg.go":                    "package liba\n\nimport \"fixture/liba/internal/secret\"\n\ntype Cfg struct {\n\tOpts secret.Options `name:\"opts\"`\n}\n\nfunc (Cfg) Validate() error { return nil }\n",
	})
	named, outPkg, err := loadPackage(filepath.Join(dir, "liba"), "Cfg", hermeticEnv())
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, "none")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "liba", "cfg_configulator.go"), out, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = hermeticEnv()
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("same-tree internal output does not compile: %v\n%s", err, o)
	}
}

func TestNoValidateFlag(t *testing.T) {
	m, err := buildFixtureModel(t, map[string]string{
		"cfg.go": "package fixture\n\ntype Cfg struct {\n\tA string `name:\"a\"`\n}\n",
	}, "Cfg", true)
	if err != nil {
		t.Fatalf("-no-validate should permit a Validate-less config: %v", err)
	}
	if m.HasValidate {
		t.Fatal("HasValidate should be false")
	}
}

func TestURLBecomesSlot(t *testing.T) {
	m, err := buildFixtureModel(t, map[string]string{
		"cfg.go": "package fixture\n\nimport \"net/url\"\n\ntype Cfg struct {\n\tU url.URL `name:\"u\"`\n}\n" + validateStub,
	}, "Cfg", false)
	if err != nil {
		t.Fatalf("url.URL should map to the URL sentinel slot: %v", err)
	}
	if m.Fields[0].Kind != KindStdSlot || m.Fields[0].SlotType != "URL" {
		t.Fatalf("url.URL classified as %v/%s, want StdSlot/URL", m.Fields[0].Kind, m.Fields[0].SlotType)
	}
}

// TestStdFlagsOutputCompiles is the -flags=std smoke: generate into a
// hermetic module and build it.
func TestStdFlagsOutputCompiles(t *testing.T) {
	src := "package fixture\n\nimport \"time\"\n\ntype Cfg struct {\n\tName string `name:\"name\" default:\"x\"`\n\tPort uint16 `name:\"port\" default:\"8080\"`\n\tWait time.Duration `name:\"wait\" default:\"5s\"`\n\tOpt *int64 `name:\"opt\"`\n\tTags []string `name:\"tags\"`\n}\n" + validateStub
	dir := writeModule(t, map[string]string{"cfg.go": src})
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv())
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, "std")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = hermeticEnv()
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("std output does not compile: %v\n%s", err, o)
	}
	// short: under std is a generate-time error
	if p := findShortTag([]*Field{{Tag: "s", Short: "s"}}, ""); p == "" {
		t.Fatal("findShortTag missed a short: tag")
	}
}

func TestShortUnderStdRejected(t *testing.T) {
	m, err := buildFixtureModel(t, map[string]string{
		"cfg.go": "package fixture\n\ntype Cfg struct {\n\tA string `name:\"a\" short:\"a\"`\n}\n" + validateStub,
	}, "Cfg", false)
	if err != nil {
		t.Fatal(err)
	}
	if p := findShortTag(m.Fields, ""); p != "a" {
		t.Fatalf("findShortTag = %q, want \"a\"", p)
	}
}

func TestSampleFormats(t *testing.T) {
	src := "package fixture\n\ntype Sub struct {\n\tHost string `name:\"host\" default:\"localhost\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Sub    `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, map[string]string{"cfg.go": src}, "Cfg", false)
	if err != nil {
		t.Fatal(err)
	}

	j, err := emitSampleJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"port": 8080`, `"key": "(secret)"`, `"host": "localhost"`, `"tags": [`, `"a",`} {
		if !strings.Contains(string(j), want) {
			t.Errorf("json sample missing %q:\n%s", want, j)
		}
	}

	tm := string(emitSampleTOML(m))
	for _, want := range []string{"port = 8080", `key = "(secret)"`, "[sub]", `host = "localhost"`, `tags = ["a", "b"]`} {
		if !strings.Contains(tm, want) {
			t.Errorf("toml sample missing %q:\n%s", want, tm)
		}
	}
	if strings.Index(tm, "[sub]") < strings.Index(tm, "tags =") {
		t.Errorf("toml scalars must precede tables:\n%s", tm)
	}
}

func TestMarkdown(t *testing.T) {
	src := "package fixture\n\ntype Sub struct {\n\tHost string `name:\"host\" default:\"localhost\" description:\"bind host\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\" required:\"true\" description:\"listen port\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Sub    `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, map[string]string{"cfg.go": src}, "Cfg", false)
	if err != nil {
		t.Fatal(err)
	}
	md := string(emitMarkdown(m, ".", "APP_", "_"))
	// Cells are padded for alignment; normalize runs of spaces per line
	// before asserting content. Mirrors the Rust CLI test.
	var squeezed []string
	for _, line := range strings.Split(md, "\n") {
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
	src := "package fixture\n\ntype Cfg struct {\n" +
		"\tHost     string `name:\"host\" env:\"HOSTNAME_OVERRIDE\" flag:\"hostname\"`\n" +
		"\tInternal string `name:\"internal\" env:\"-\" flag:\"-\"`\n" +
		"\tPort     uint16 `name:\"port\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, map[string]string{"cfg.go": src}, "Cfg", false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, "pflag")
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
	// flag:"-" must not register the flag; env:"-" must not read the var.
	if strings.Contains(code, `"INTERNAL"`) {
		t.Errorf("env:\"-\" field still read from env:\n%s", code)
	}
}

func TestBadEnvOverrideRejected(t *testing.T) {
	src := "package fixture\n\ntype Cfg struct {\n" +
		"\tHost string `name:\"host\" env:\"lower-case\"`\n}\n" + validateStub
	_, err := buildFixtureModel(t, map[string]string{"cfg.go": src}, "Cfg", false)
	if err == nil || !strings.Contains(err.Error(), "uppercase") {
		t.Fatalf("want uppercase-override error, got %v", err)
	}
}

func TestSchemaAndSample(t *testing.T) {
	src := "package fixture\n\ntype Sub struct {\n\tHost string `name:\"host\" default:\"localhost\" description:\"bind host\"`\n}\n\ntype Cfg struct {\n\tPort uint16 `name:\"port\" default:\"8080\" required:\"true\" description:\"listen port\"`\n\tKey  string `name:\"key\" secret:\"true\"`\n\tSub  Sub    `name:\"sub\"`\n\tTags []string `name:\"tags\" default:\"a,b\"`\n}\n" + validateStub
	m, err := buildFixtureModel(t, map[string]string{"cfg.go": src}, "Cfg", false)
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
