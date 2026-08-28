# configulator behavioral specification

Version: 0.1.0 (see `SPEC_VERSION`)

This document governs two implementations:

- **Go**: `github.com/USA-RedDragon/configulator/v2` (this repository)
- **Rust**: `configulator-rs` (crate `configulator-rs`, lib `configulator`)

Both implementations MUST satisfy every **normative rule** below and MUST pass
the conformance corpus under `spec/cases/`. Behaviors listed in the
**capability tables** are owned by the underlying decoder or argument parser,
are recorded rather than mandated, and MAY differ per language and per format.

Canonical corpus format: **JSON**, decoded by `encoding/json/v2` (Go) and
`serde_json` (Rust). YAML and TOML conformance is exercised by per-repo format
tables, not by the shared corpus: a shared YAML fixture would test two
independent YAML parsers, not configulator.

## Terms

- **Layer**: one of *defaults*, *file*, *env*, *CLI*.
- **Shadow**: a generated mirror of the config type in which every field is
  presence-aware (Go: pointer; Rust: `Option<T>`).
- **Presence**: a layer *set* a field. File: shadow field non-nil/`Some`.
  Env: variable exists (`LookupEnv` / `env::var`). CLI: value came from the
  command line (pflag `Changed` / clap `ValueSource::CommandLine`).
- **Leaf**: a field decoded from a single scalar. **Collection**: a map or a
  list of structs. **Composite**: a nested struct.
- **Opaque leaf**: a struct-kind type decoded from a text scalar via
  `TextUnmarshaler`-equivalent machinery, carried in the shadow as a
  **sentinel slot** (unexported fields plus a set-flag) so that a decoder
  which walks fields instead of calling the unmarshaler produces a loud
  error, never a silently wrong value.
- **Explicit path**: a config-file path the operator named (`--config <path>`
  or the API's explicit-path option). **Search path**: a path from the
  configured search list.

## Normative rules

1. **Presence oracle.** The file layer decodes into the shadow; env and CLI
   use their native presence checks. No layer infers presence from a value
   comparing equal to a default or zero value.

2. **Precedence** is `defaults < file < env < CLI`. CLI is *parsed* before
   the file layer runs (it supplies `--config`) but *merged* last.

3. **Cross-layer merge.** Nested structs deep-merge: each layer assigns only
   the fields it set. Scalars and lists replace wholesale; a list set by a
   higher layer fully replaces a lower layer's list, never appends.
   Element defaults inside collections apply per element, emitted by the
   file layer's element constructor, and record origin
   `{default, "element default"}`.

4. **Explicit vs. search paths.** An explicit path MUST exist and be
   readable; any failure is a hard error naming the path and the underlying
   IO error, regardless of any not-found-tolerance option, with **no
   fallback** to search paths. Search paths miss softly; the first readable
   hit wins; a search path that exists but is a directory or unreadable is
   an error, not a soft miss. An empty file is a successful load of zero
   keys, distinct from not-found.

5. **Optionals are first class.** Go `*T`, Rust `Option<T>`, across all four
   layers. `null` in a config file is treated as **absent**. `*scalar` is
   env- and CLI-settable via allocate-on-set. Leaves inside a `*Struct` are
   env- and CLI-settable; the struct is allocated on first write. Allocation
   records no origin; only leaf writes do. A `*T` never written by any layer
   stays nil/`None`.

6. **Collections** are fully supported by the file layer and skipped, with a
   generator warning, for env and CLI. An explicit `env:"NAME"`/`flag:"name"`
   opt-in on a collection is a generate-time error.

7. **Env naming** is `PREFIX + upper(tag_name)` with nested levels joined by
   `SEPARATOR`, recursing with the **tag name** at every level. `-` → `_`
   applies **only to tag-derived segments**; prefix and separator are used
   verbatim. A prefix or separator that cannot survive this construction
   (lowercase prefix, separator containing `-`) is a typed runtime error at
   load, not a silent misconstruction.

8. **`--config` travels out of band**, never through the value tree under a
   sentinel key.

9. **Origin reporting.** Dotted paths map to `{layer, detail}` where detail
   is the file path, env var name, or flag name. Path grammar: tag names
   joined by `.`; list indices as `[i]`; map keys quoted when they contain
   `.` or `[`. Recording is **per element** inside collections. Fields never
   set and lacking a default have no entry. A report requested before load
   is empty, never an error or null.

10. **Zero values.** A field absent from every layer with no default keeps
    the language zero value.

11. **Load on error** yields the defaults plus every layer that *completed*;
    a layer that fails part-way contributes nothing (layers stage and commit,
    never mutate the result mid-layer). (Go returns this partial config
    alongside the error; Rust's `Result` cannot, a recorded divergence.)

## Capability tables (recorded, not mandated)

| Behavior | Owner | Go (pflag / decoder) | Rust (clap / decoder) |
|---|---|---|---|
| Repeated-flag accumulation | arg parser | `--x a --x b` → `[a,b]`; `--x a,b` also splits (CSV) | `--x a --x b` → `[a,b]`; `--x a,b` is one element |
| List separator on CLI | arg parser | comma, fixed | none (repeat the flag) |
| List separator on env/defaults | configulator | configurable | configurable (0.2) |
| File-key case sensitivity | decoder | json/v2: sensitive; goccy: sensitive; go-toml: **insensitive** | serde_json: sensitive; serde_yaml_ng: sensitive; toml: sensitive |
| Unknown-key rejection | decoder | json/v2: available (`RejectUnknownMembers`); goccy: opt-in; go-toml: decoder-level | `deny_unknown_fields`, derive-time, file layer only |
| Duplicate keys in one file | decoder | json/v2 rejects; goccy last-wins; go-toml errors | serde_json last-wins; yaml/toml per parser |
| Dual-interface leaf types (e.g. `TextUnmarshaler` + `UnmarshalJSON`) | decoder | dispatch differs per decoder; not silently allowed — generator warns | n/a (leaves use `FromStr`) |

Pinned versions: pflag v1.0.6+, goccy/go-yaml v1.19.2+, pelletier/go-toml/v2
v2.2.4+, clap 4, serde_yaml_ng 0.10, toml 1.1.

## Also specified

- Empty-string env values: the variable is **present**; it parses as the
  empty string (strings), an empty list (lists), or a parse error (numerics).
- Whitespace in list elements: not trimmed in Go; Rust trims and drops
  empties today — divergence recorded, converges on **no trimming** in 0.2.
- Separator escaping: unsupported, documented.
- `WithX` call ordering: none required.
- `--config` default value: the first search path, shown in help; using the
  default is not "explicit". If no file support is configured, no `--config`
  flag is registered and passing one is an unknown-flag error.
- Symlinks/relative paths: OS semantics, not normalized.
- Decoder-map keys are lowercased extensions matched literally — no
  aliasing; register both `.yml` and `.yaml` to accept both. An
  extension-less path is an error naming the path.

## Conformance corpus

Layout: `spec/cases/<case>/` containing:

- `config.json` — file-layer input (optional)
- `env.json` — env vars as a string map (optional)
- `argv.json` — CLI args as a string array (optional)
- `options.json` — `{"prefix", "env_separator", "flag_separator",
  "array_separator", "capabilities": [...]}` (optional; defaults
  `"APP_"`, `"_"`, `"."`, `","`, `[]`)
- `shape` — one of the shape names below
- `expect.json` — expected config as a nested JSON object, **or**
- `expect_errors.json` — `{"kind": "<logical kind>", "contains": [..]}`
- `expect_origins.json` — optional; dotted path → `{"layer", "detail"}`.
  Excluded from the release gate for collection-shaped cases until Rust
  origin support lands.

Error kinds: `ExplicitFileMissing`, `SearchPathUnreadable`, `ParseError`,
`UnknownKey`, `ValidationError`, `BadEnvOptions`.

**Shapes** are hand-written once per language, field-for-field identical:

- `scalars`: string/bool/i64-int/u16-int/f64 with defaults on some fields
- `nested`: two levels of nested struct, defaults at both levels
- `collections`: `[]string`, `[]struct` (element default), `map[string]string`, `map[string]struct`
- `optionals`: optional u16, optional string (with default), optional struct
- `durations`: duration leaf plus a plain string

Shape definitions live in `spec/shapes.md`; each has a `defaults-only` case
pinning every default and a boundary case pinning integer widths.

**Runner contract:** enumerate every directory under `spec/cases/`; any case
not executed is a failure unless named in the checked-in `spec/skip-<lang>.txt`,
which CI prints and which must be empty for a release. Numeric comparison is
normalized (JSON numbers compare by value, not representation). Cases whose
`options.json` lists a capability run only against decoders having it.

## Synchronization

`spec/SPEC_VERSION` holds this document's version. The Go repository
publishes `spec/` as a tarball on tags named `spec/vX.Y.Z`. The Rust
repository pins a version in `.spec-pin` and its CI fetches that tarball.
A change to `spec/` requires both repositories green before the pin
advances. Rules may be added in minor versions; changed or removed rules
require a major version and a migration note.
