# configulator behavioral specification

Version: 0.2.0 (see `SPEC_VERSION`)

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
   verbatim. An empty separator means `_`. A prefix or separator that
   cannot survive this construction (lowercase prefix, separator containing
   `-`) is a typed runtime error at load, not a silent misconstruction.

8. **`--config` travels out of band**, never through the value tree under a
   sentinel key.

9. **Origin reporting.** Dotted paths map to `{layer, detail}` where detail
   is the file path, env var name, or flag name (`--name`). A `default` tag
   records detail `default tag`, and an element default `element default`.
   Path grammar: tag names joined by `.`; list indices as `[i]`; map keys
   quoted with `"` when they contain `.` or `[`, and inside the quotes `\`
   and `"` are escaped with a backslash (`pools."q\"x.y".size`). Recording
   is **per element** inside collections. Fields never set and lacking a
   default have no entry. A report requested before load is empty, never
   an error or null.

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
| List separator on env/defaults | configulator | configurable | configurable |
| Bool flag value | arg parser | `--x` or `--x=false`; in `--x false`, `false` is a positional argument | `--x`, `--x=false` and `--x false` |
| File-key case sensitivity | decoder | json/v2: sensitive; goccy: sensitive; go-toml: **insensitive** | serde_json: sensitive; serde_yaml_ng: sensitive; toml: sensitive |
| Unknown-key rejection | decoder | json/v2: available (`RejectUnknownMembers`); goccy: opt-in; go-toml: decoder-level | `deny_unknown_fields`, derive-time, file layer only |
| Duplicate keys in one file | decoder | json/v2 rejects; goccy last-wins; go-toml errors | serde_json last-wins; yaml/toml per parser |
| Dual-interface leaf types (e.g. `TextUnmarshaler` + `UnmarshalJSON`) | decoder | dispatch differs per decoder; not silently allowed — generator warns | n/a (leaves use `FromStr`) |
| Decoder selection | configulator | by lowercased file extension (`Decoders` map) | one `FileLoader` for every path |
| Negative durations | configulator | allowed (`-1h`) | parse error (`Duration` is unsigned) |

Pinned versions: pflag v1.0.6+, goccy/go-yaml v1.19.2+, pelletier/go-toml/v2
v2.2.4+, clap 4, serde_yaml_ng 0.10, toml 1.1.

## Also specified

- Empty-string env values: the variable is **present**; it parses as the
  empty string (strings), an empty list (lists), or a parse error (numerics).
- Whitespace in list elements: not trimmed, and empty elements are kept
  (`" a ,,b"` is `[" a ", "", "b"]`).
- Separator escaping: unsupported, documented.
- `WithX` call ordering: none required.
- `--config` default value: the first search path, shown in help; using the
  default is not "explicit". If no file support is configured, no `--config`
  flag is registered and passing one is an unknown-flag error.
- Symlinks/relative paths: OS semantics, not normalized.
- Complex numbers: a leaf decoded from text like `1+2i`, `(1+2i)`, `2i` or
  `3`, as Go's `strconv.ParseComplex` reads it. A file may also hold a plain
  number. `j` in place of `i` is a parse error. Go: `complex64` and
  `complex128`. Rust: `configulator::Complex64` and `Complex128`.
- Decoder-map keys (Go) are lowercased extensions matched literally — no
  aliasing; register both `.yml` and `.yaml` to accept both. An
  extension-less path is an error naming the path.
- File scalar types are strict, as in Go's `encoding/json/v2`: a number for
  a string field, a string for a number or bool field, and a fraction for an
  integer field are errors. An integer for a float field is fine.
- Bools in env vars, flag values and defaults use Go's `strconv.ParseBool`
  spellings: `1`, `t`, `T`, `TRUE`, `true`, `True`, `0`, `f`, `F`, `FALSE`,
  `false`, `False`. Anything else is a parse error.
- Durations use Go's `time.ParseDuration` syntax and must fit in an int64
  of nanoseconds (about 2562047h). They print as Go's `Duration.String()`.
- Secrets: no error message quotes the value of a field marked secret,
  whatever layer it came from. The value shows as `(redacted)`.
- Flag help shows each flag's default, except for secret fields. The
  `--config` flag's help is `config file`, with the first search path as
  its default.
- `PrintConfig` prints one `path = value` line per leaf, in field order,
  with nested structs flattened into dotted paths. Values print as Go's
  `%v` would: `[a b]` for a list, `map[a:x b:y]` for a map (keys sorted),
  `[{h:1 1}]` for a list of structs, `30s` for a duration. An unset optional
  prints `<unset>`, and an unset optional struct prints one `<unset>` line
  for itself. A secret prints `(redacted)`, and so does a whole list or map
  of structs that holds a secret anywhere.
- The generator CLIs exit 2 on a usage error, with a message starting
  `configulator:`.

## Conformance corpus

Layout: `spec/cases/<case>/` containing:

- `config.json` — file-layer input (optional)
- `env.json` — env vars as a string map (optional)
- `argv.json` — CLI args as a string array (optional)
- `options.json` — `{"prefix", "env_separator", "flag_separator",
  "array_separator", "capabilities": [...]}` (optional; defaults
  `"APP_"`, `"_"`, `"."`, `","`, `[]`). An `env_separator` of `""` is
  passed to the library as empty, so the library default applies.
- `shape` — one of the shape names below
- `expect.json` — expected config as a nested JSON object, **or**
- `expect_errors.json` — `{"kind": "<logical kind>", "contains": [..],
  "excludes": [..]}`. The message must contain every `contains` string and
  none of the `excludes` strings.
- `expect_origins.json` — optional; dotted path → `{"layer", "detail"}`.
  A file-layer detail is matched against the end of the path the runner
  passed, since that path depends on where the corpus lives.

`config.json` is loaded as a search path. It may be a directory, to test a
search path that can't be read.

Error kinds:

- `ExplicitFileMissing`: a `--config` or explicit path that can't be read
- `SearchPathUnreadable`: a search path that exists but can't be read
- `DecodeError`: the file decoder failed, including type mismatches and
  out-of-range numbers
- `UnknownKey`: the file has a key no field matches
- `ParseError`: an env var, flag value or default failed to parse
- `RequiredError`: a required field no layer set
- `ValidationError`: the config's validation hook failed
- `BadEnvOptions`: a lowercase prefix or a separator containing `-`
- `FlagError`: the argument parser rejected the command line, such as an
  unknown flag

**Shapes** are hand-written once per language, field-for-field identical:

- `scalars`: string/bool/i64-int/u16-int/f64 with defaults on some fields
- `nested`: two levels of nested struct, defaults at both levels
- `collections`: `[]string`, `[]struct` (element default), `map[string]string`, `map[string]struct`
- `nested-collections`: lists and maps of structs inside list and map elements, with defaults at each level
- `optionals`: optional u16, optional string (with default), optional struct
- `durations`: duration leaf plus a plain string
- `complex`: complex128 and complex64 leaves and a list of complex128
- `required`: required fields at the top level, in a nested struct, in an
  optional struct and in list elements, plus a validation hook
- `attributes`: a secret, env and flag renames, env and flag skips, and a
  short flag

Shape definitions live in `spec/shapes.md`; each has a `defaults-only` case
pinning every default and a boundary case pinning integer widths.

**Runner contract:** enumerate every directory under `spec/cases/`; any case
not executed is a failure unless named in the checked-in `spec/skip-<lang>.txt`,
which CI prints and which must be empty for a release. Numeric comparison is
normalized (JSON numbers compare by value, not representation). Cases whose
`options.json` lists a capability run only against decoders having it.

## Synchronization

`spec/SPEC_VERSION` holds this document's version. The Go repository runs
the corpus from its own `spec/` directory. The Rust CI checks out the
configulator repository's default branch, fails unless its
`spec/SPEC_VERSION` equals the `EXPECTED_SPEC_VERSION` pinned in the Rust
workflow, and runs the corpus from that checkout. A change to `spec/` bumps
`SPEC_VERSION`. Cases Rust can't pass yet go in `spec/skip-rust.txt`, and
the Rust pin moves to the new version once its corpus job passes. Rules may
be added in minor versions; changed or removed rules require a major
version and a migration note.
