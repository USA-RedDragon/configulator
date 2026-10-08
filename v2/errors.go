//go:build goexperiment.jsonv2

package configulator

import "fmt"

// ParseError reports a value that could not be parsed into its field.
type ParseError struct {
	// Path is the dotted field path (tag names).
	Path string
	// Source names the input: an env var, a flag, or a file path.
	Source string
	// Value is the raw input, or "(redacted)" for fields tagged secret:.
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: cannot parse %q from %s: %v", e.Path, e.Value, e.Source, e.Err)
}
func (e *ParseError) Unwrap() error { return e.Err }

// MissingFileError reports that a config file named by --config or
// FileOptions.Explicit could not be read. It is returned regardless of
// FileOptions.RequireFound, and the search paths are not tried instead.
type MissingFileError struct {
	Path string
	// Searched lists the configured search paths. They were not tried.
	Searched []string
	Err      error
}

func (e *MissingFileError) Error() string {
	return fmt.Sprintf("config file %s: %v", e.Path, e.Err)
}
func (e *MissingFileError) Unwrap() error { return e.Err }

// SearchPathError reports a search path that exists but can't be read,
// such as a directory. A search path that doesn't exist is just skipped.
type SearchPathError struct {
	Path string
	Err  error
}

func (e *SearchPathError) Error() string {
	return fmt.Sprintf("config search path %s: %v", e.Path, e.Err)
}
func (e *SearchPathError) Unwrap() error { return e.Err }

// NoFileFoundError reports that RequireFound was set and no search path
// matched.
type NoFileFoundError struct {
	Searched []string
}

func (e *NoFileFoundError) Error() string {
	return fmt.Sprintf("no config file found; searched %v", e.Searched)
}

// RequiredError reports a field tagged required:"true" that no layer set.
type RequiredError struct {
	Path string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("%s: required but not set by any layer", e.Path)
}

// BadEnvOptionsError is returned by Load when EnvironmentVariableOptions has
// a prefix that isn't uppercase or a separator containing "-". Env names
// built from either could never match.
type BadEnvOptionsError struct {
	Field  string // "Prefix" or "Separator"
	Value  string
	Reason string
}

func (e *BadEnvOptionsError) Error() string {
	return fmt.Sprintf("EnvironmentVariableOptions.%s %q: %s", e.Field, e.Value, e.Reason)
}

// DecodeError wraps an error from the Unmarshal function for a config file.
type DecodeError struct {
	Path string
	Err  error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("decoding %s: %v", e.Path, e.Err)
}
func (e *DecodeError) Unwrap() error { return e.Err }

// OpaqueSpellingError reports a config file that wrote a text value (like a
// CIDR or duration) as a nested table/mapping instead of a string.
type OpaqueSpellingError struct {
	Path string
	Hint string // e.g. `"10.0.0.0/8"`
}

func (e *OpaqueSpellingError) Error() string {
	return fmt.Sprintf("%s: expected a text scalar (e.g. %s), got a nested table/mapping", e.Path, e.Hint)
}

// FlagConflictError is returned by Load when a flag adapter's Bind could
// not add a flag because its name or shorthand is already taken, by a flag
// already on the FlagSet or by another config field. pflag and flag panic
// on a duplicate, so Bind returns this through Load instead.
type FlagConflictError struct {
	// Flag is the name of the flag being added, without dashes.
	Flag string
	// Shorthand is set when the shorthand conflicts rather than the name.
	Shorthand string
	// Existing is the flag that already holds the name or shorthand.
	Existing string
}

func (e *FlagConflictError) Error() string {
	if e.Shorthand != "" {
		return fmt.Sprintf("flag %q: shorthand %q is already used by flag %q", e.Flag, e.Shorthand, e.Existing)
	}
	return fmt.Sprintf("flag %q is already defined; rename one of them (flag:\"name\" tag, FileOptions.FlagName) or skip the field with flag:\"-\"", e.Flag)
}
