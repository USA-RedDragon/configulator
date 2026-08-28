//go:build goexperiment.jsonv2

package configulator

// Getenv is the environment lookup seam. The default is os.LookupEnv;
// tests may inject a map-backed implementation
type Getenv func(string) (string, bool)

// Unmarshal is the pluggable decoder seam: any function with the shape of
// json.Unmarshal / yaml.Unmarshal / toml.Unmarshal
type Unmarshal func([]byte, any) error

// Decoders maps lowercased file extensions (with dot) to decoders,
// matched literally: register both ".yml" and ".yaml" if you want both.
// A nil Decoders means strict encoding/json/v2 for ".json" and an error
// for anything else.
type Decoders map[string]Unmarshal

// EnvironmentVariableOptions configures the env layer.
type EnvironmentVariableOptions struct {
	// Prefix is prepended verbatim to every variable name and must be
	// uppercase (validated at Load; see BadEnvOptionsError).
	Prefix string
	// Separator joins nested levels, verbatim, and must not contain "-"
	// (validated at Load). Empty means "__".
	Separator string
}

// FileOptions configures the file layer.
type FileOptions struct {
	// Search paths are tried in order; the first readable file wins. A
	// missing search path is a soft miss; a directory or unreadable path
	// is a SearchPathError.
	Search []string
	// Explicit, if set, MUST exist and be readable (MissingFileError
	// otherwise) and Search is not consulted, the same rule as --config.
	Explicit string
	// RequireFound makes "no search path matched" an error
	// (NoFileFoundError). It does not affect Explicit, which is always
	// required.
	RequireFound bool
	// Decoders picks the decoder by file extension; nil means strict
	// json/v2 for ".json" only.
	Decoders Decoders
	// FlagName is the config-file flag name; "" means "config".
	FlagName string
	// Shorthand is the config-file flag shorthand; "" means "c" when a
	// flag adapter that supports shorthands is bound, and is an error
	// under adapters that do not (stdlib flag).
	Shorthand string
}
