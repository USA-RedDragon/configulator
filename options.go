package configulator

import "github.com/spf13/pflag"

// EnvironmentVariableOptions contains the options for loading configuration from environment variables.
type EnvironmentVariableOptions struct {
	// Prefix is the prefix for the environment variables. If empty, no prefix is used.
	Prefix string
	// Separator is the separator for the environment variables if nested. If empty, "__" is used.
	Separator string
}

// WithEnvironmentVariables sets the options for loading configuration from environment variables.
func (c *Configulator[C]) WithEnvironmentVariables(opts *EnvironmentVariableOptions) *Configulator[C] {
	c.envOptions = opts
	return c
}

// WithArraySeparator sets the separator for array values
func (c *Configulator[C]) WithArraySeparator(sep string) *Configulator[C] {
	c.arraySeparator = sep
	return c
}

// FileFormat are the valid formats of the config file.
type FileFormat int

// FileOptions contains the options for loading configuration from files.
type FileOptions struct {
	// Paths is the list of paths to search for the configuration file.
	Paths []string
	// ErrorIfNotFound indicates if an error should be returned if no config file is found.
	ErrorIfNotFound bool
}

// WithFile sets the options for loading configuration from files.
func (c *Configulator[C]) WithFile(opts *FileOptions) *Configulator[C] {
	c.fileOptions = opts
	return c
}

// PFlagOptions contains the options for loading configuration from pflags.
type PFlagOptions struct {
	// Separator is the separator for the pflags if nested. If empty, "." is used.
	Separator string
}

// WithPFlags sets the options for loading configuration from pflags.
// If WithFile() is used first, a flag for the config file is added.
func (c *Configulator[C]) WithPFlags(flags *pflag.FlagSet, opts *PFlagOptions) *Configulator[C] {
	c.flags = flags
	if opts == nil {
		opts = &PFlagOptions{}
	}
	c.pflagOptions = opts
	err := c.registerFlags()
	if err != nil {
		panic(err)
	}
	return c
}
