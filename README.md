# Configulator

[![go.mod version](https://img.shields.io/github/go-mod/go-version/USA-RedDragon/configulator.svg)](https://github.com/USA-RedDragon/configulator) [![codecov](https://codecov.io/gh/USA-RedDragon/configulator/graph/badge.svg?token=AhUJaQtw9R)](https://codecov.io/gh/USA-RedDragon/configulator) [![License](https://badgen.net/github/license/USA-RedDragon/configulator)](https://github.com/USA-RedDragon/configulator/blob/main/LICENSE) [![GitHub contributors](https://badgen.net/github/contributors/USA-RedDragon/configulator)](https://github.com/USA-RedDragon/configulator/graphs/contributors/) [![GoReportCard example](https://goreportcard.com/badge/github.com/USA-RedDragon/configulator)](https://goreportcard.com/report/github.com/USA-RedDragon/configulator)

A simple configuration manager for use in my apps. Config comes from
defaults, files, environment variables, and command-line flags, merged in
that order, with per-field origin reporting.

v2 lives in [v2/](v2/) and replaces v1's runtime reflection with generated
code (`go:generate`). A Rust sibling,
[configulator-rs](https://github.com/USA-RedDragon/configulator-rs), is kept
in feature parity against the shared [spec/](spec/).

## v2

### Features

- Zero reflection at runtime — a `go:generate` step emits the loading code
- Pluggable file decoders by extension; strict `encoding/json/v2` built in,
  bring your own YAML/TOML (e.g. `goccy/go-yaml`, `pelletier/go-toml/v2`)
- Environment variables and typed CLI flags (`spf13/pflag` or stdlib `flag`)
- A typo'd `--config` is a hard error, not a silent fall-through to defaults
- Full type support: nested structs, pointers/optionals, maps, slices of
  structs, `time.Duration` and friends via wrapper types
- Per-field origin report: which layer set each value, and from where
- `secret:"true"` redaction, `required:"true"` checks, generated
  `PrintConfig()`, JSON Schema and sample-config emission
- Flag-free builds: with `-flags=none`, no flag library is compiled in

### Quick start

```go
package main

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config

import (
	configulator "github.com/USA-RedDragon/configulator/v2"
)

type Config struct {
	Name string `name:"name" default:"myapp"`
	Port uint16 `name:"port" default:"8080"`
}

func (c Config) Validate() error { return nil }

func main() {
	cfg, err := configulator.New(ConfigSchema()).
		WithEnvironmentVariables(nil).
		Load()
	// ...
}
```

`go generate ./...` writes `config_configulator.go` next to your struct,
which provides `ConfigSchema()` (and `ConfigPFlagHooks()` unless
`-flags=none`). See [v2/examples/](v2/examples/) for the full spread:
[all](v2/examples/all/) (YAML + cobra/pflag + origins),
[minimal](v2/examples/minimal/), [jsonv2](v2/examples/jsonv2/) (flag-free),
[toml](v2/examples/toml/), and [stdflag](v2/examples/stdflag/).

### Files, env, and flags

```go
c := configulator.New(ConfigSchema()).
	WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
		Prefix: "MYAPP_", Separator: "_",
	}).
	WithFile(&configulator.FileOptions{
		Search:   []string{"config.yaml"},
		Decoders: configulator.Decoders{".yaml": yaml.Unmarshal},
	})
cpflag.Bind(c, cmd.Flags(), ConfigPFlagHooks(), nil) // registers flags, including --config
```

Precedence is defaults < file < env < flags. `--config` names a file that
must exist and parse; search paths miss softly (unless `RequireFound`).
After `Load()`, `c.Report()` says where every field came from:

```text
http.host    env      MYAPP_HTTP_HOST
http.port    file     config.yaml
enable       cli      --enable
```

### Struct tags

| Tag | Meaning |
|---|---|
| `name:"key"` | key in files, env, and flags; falls back to `json:`/`yaml:` tags |
| `default:"v"` | default value, parsed at generate time |
| `description:"…"` | flag help text |
| `env:"-"` / `env:"NAME"` | skip env, or override this field's env segment (uppercase `A-Z0-9_` only, used verbatim) |
| `flag:"-"` / `flag:"name"` | skip flags, or override this field's flag name segment |
| `short:"p"` | flag shorthand (pflag only) |
| `secret:"true"` | redacted in `PrintConfig()` and error messages |
| `required:"true"` | must be set by some layer; checked at `Load()` |
| `opaque:"true"` | treat the type as a text scalar (`encoding.TextUnmarshaler`) |

Nested env names use the tag name at every level: `http.host` becomes
`PREFIX + HTTP<sep>HOST`. (This differs from v1, which used the Go field
name — a footgun v2 removes.)

`time.Duration`, `url.URL`, `net.IPNet`, and similar stdlib types are not
directly decodable; the generator maps them to wrapper types in
[v2/stdtypes.go](v2/stdtypes.go) automatically — your struct keeps the real
type.

### Generator flags

| Flag | Meaning |
|---|---|
| `-type` | config root type (required) |
| `-output` | output file, default `<type>_configulator.go` |
| `-flags` | `pflag` (default) \| `std` \| `none` |
| `-tags` | shadow struct tags, default `json,yaml,toml` |
| `-schema` | print a JSON Schema to stdout (no code generated); pipe it where you want it |
| `-sample` | print a sample config to stdout (no code generated); `-format yaml\|json\|toml`, commented YAML by default |
| `-markdown` | print a Markdown reference table of every key to stdout (`-env-prefix`, `-env-separator`, `-flag-separator` shape the env/flag columns) |
| `-no-validate` | don't require a `Validate() error` method |

### Requirements

v2 requires Go 1.27+ (`encoding/json/v2`). If your build sets
`GOEXPERIMENT=nojsonv2`, compilation fails with a self-describing error.

## v1 (frozen)

v1 is the root module (`github.com/USA-RedDragon/configulator`). It is
reflection-based, frozen, and receives bug fixes only.

<details>
<summary>v1 documentation</summary>

### Features

- Supports configuration from:
  - YAML or JSON files
  - Environment variables
  - Command line arguments (`spf13/pflag`)

### Supported types

Working: all scalars except complex, structs, arrays of scalars.

Not working: maps, arrays of maps, arrays of structs, multi-dimensional
arrays, complex scalars.

### Usage

> [!NOTE]
> Because the configuration options are expressed as different cases (i.e. `http.host` in YAML would be `HTTP__HOST` in environment variables), this library cannot be used for configurations that contain the same field name in different cases.

v1 uses the `name`, `default`, and `description` tags. Fields without a
`name` tag are not utilized. The field name can be inferred from `json` or
`yaml` tags if present.

```go
// Field appears in config files, environment variables, and command line arguments as key "myName".
Field int `name:"myName"`

// Field has a description if seen in the CLI's --help
Field int `name:"myName" description:"this text appears in the help section of the CLI"`

// Field has a default value of 1
Field int `name:"myName" default:"1"`
```

Nested env names in v1 derive from the **Go field name**, not the tag name;
v2 fixes this.

</details>
