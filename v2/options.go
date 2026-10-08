//go:build goexperiment.jsonv2

package configulator

// Getenv looks up an environment variable. The default is os.LookupEnv.
type Getenv func(string) (string, bool)

// Unmarshal decodes a config file. json.Unmarshal, yaml.Unmarshal and
// toml.Unmarshal all fit.
type Unmarshal func([]byte, any) error

// Decoders maps lowercase file extensions (with the dot) to decoders.
// Register both ".yml" and ".yaml" if you want both. A nil Decoders means
// strict encoding/json/v2 for ".json" and an error for anything else.
type Decoders map[string]Unmarshal

// EnvironmentVariableOptions configures the env layer.
type EnvironmentVariableOptions struct {
	// Prefix is prepended as-is to every variable name and must be
	// uppercase.
	Prefix string
	// Separator joins nested levels and must not contain "-". Empty
	// means "_".
	Separator string
}

// FileOptions configures the file layer.
type FileOptions struct {
	// Search paths are tried in order and the first readable file wins.
	// Missing paths are skipped; a directory or unreadable path is a
	// SearchPathError.
	Search []string
	// Explicit, if set, must exist and be readable (MissingFileError
	// otherwise), and Search is ignored. --config works the same way.
	Explicit string
	// RequireFound returns NoFileFoundError when no search path matched.
	RequireFound bool
	// Decoders picks the decoder by file extension. See Decoders.
	Decoders Decoders
	// FlagName is the config-file flag name; "" means "config".
	FlagName string
	// Shorthand is the config-file flag shorthand; "" means "c" with the
	// pflag adapter. Setting it with the stdlib flag adapter is an error.
	Shorthand string
}
