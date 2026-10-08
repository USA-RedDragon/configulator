# Conformance shapes

Each shape is hand-written once per language, field-for-field identical.
`tag` is the configulator name (Go `name:` tag / Rust `#[configulator(name)]`).
Types are given as Go / Rust pairs. A missing default means none.

Comparison normalization (runner contract): JSON numbers compare by value;
a nil/empty slice or map compares equal to `[]` / `{}`; absent optionals
appear as `null` in `expect.json`; durations appear as Go duration strings
(`"30s"`); complex numbers appear as `[real, imag]` (`1+2i` is `[1, 2]`).

## scalars

| tag | type | default |
|---|---|---|
| `name` | string / String | `svc` |
| `count` | int64 / i64 | — |
| `port` | uint16 / u16 | `8080` |
| `ratio` | float64 / f64 | `1.5` |
| `verbose` | bool / bool | `false` |

## nested

| tag | type | default |
|---|---|---|
| `app-name` | string / String | `myapp` |
| `http.host` | string / String | `localhost` |
| `http.port` | uint16 / u16 | `8080` |
| `db.url` | string / String | `postgres://localhost/db` |
| `db.pool.size` | uint16 / u16 | `10` |

## collections

| tag | type | default |
|---|---|---|
| `tags` | []string / Vec<String> | `a,b` |
| `labels` | map[string]string / HashMap<String,String> | — |
| `servers` | []Server / Vec<Server> | — |
| `pools` | map[string]Pool / HashMap<String,Pool> | — |
| `log-level` | string / String | `info` |

Server: `addr` string (no default), `weight` uint16/u16 (default `1`).
Pool: `size` uint16/u16 (default `5`).
Element defaults apply per element (SPEC rule 3).

## nested-collections

| tag | type | default |
|---|---|---|
| `peers` | []Peer / Vec<Peer> | — |
| `by-name` | map[string]Peer / HashMap<String,Peer> | — |

Peer: `name` string (no default), `slots` int64/i64 (default `3`),
`rules` []Rule / Vec<Rule>, `tags` map[string]Rule / HashMap<String,Rule>,
`inner` Level, `opt` *Level / Option<Level>.
Rule: `from` int64/i64 (no default), `range` int64/i64 (default `1`),
`on` bool (default `true`).
Level: `level` int64/i64 (default `7`).
Element defaults apply per element at every depth (SPEC rule 3), and an
optional struct inside an element gets its defaults when allocated (rule 5).

## optionals

| tag | type | default |
|---|---|---|
| `port` | *uint16 / Option<u16> | — |
| `name` | *string / Option<String> | `opt-name` (⇒ allocated to Some) |
| `tls` | *Tls / Option<Tls> | — |

Tls: `cert` string (no default), `min-version` uint16/u16 (default `12`).
Defaults on fields inside an optional struct apply when any layer allocates
it — the same element-constructor mechanism as collection element defaults.

## durations

| tag | type | default |
|---|---|---|
| `timeout` | time.Duration / Duration wrapper | `30s` |
| `label` | string / String | — |

## complex

| tag | type | default |
|---|---|---|
| `z` | complex128 / Complex128 | `1+2i` |
| `exponent` | complex128 / Complex128 | — |
| `w` | complex64 / Complex64 | — |
| `zs` | []complex128 / Vec<Complex128> | — |

## required

| tag | type | default |
|---|---|---|
| `top` | string / String, required | — |
| `nested.leaf` | string / String, required | — |
| `opt` | *ROpt / Option<ROpt> | — |
| `items` | []ROpt / Vec<ROpt> | — |

ROpt: `leaf` string (required), `other` string.
Validation fails when `top` is `invalid`, with a message containing
`invalid`.

## attributes

| tag | type | attributes |
|---|---|---|
| `token` | int64 / i64 | secret |
| `renamed` | string / String | env `RN`, flag `rn` |
| `no-env` | string / String | env `-` |
| `no-flag` | string / String | flag `-` |
| `port` | uint16 / u16 | short `p` |
