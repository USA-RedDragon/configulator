// Package configulator loads a config struct from defaults, a config file,
// environment variables and command-line flags. Each layer overrides the
// one before it.
//
// There is no reflection at run time. The configulator command reads your
// struct and writes the loading code next to it.
//
// # Getting started
//
// Install the library and the generator:
//
//	go get github.com/USA-RedDragon/configulator/v2
//	go get -tool github.com/USA-RedDragon/configulator/v2/cmd/configulator
//
// Write the config struct, with a go:generate line and a Validate method:
//
//	//go:generate go tool configulator -type Config
//
//	type Config struct {
//		LogLevel string   `name:"log-level" default:"info" description:"log level"`
//		HTTP     Listener `name:"http"`
//	}
//
//	type Listener struct {
//		Host string `name:"host" default:"localhost" description:"listen address"`
//		Port uint16 `name:"port" default:"8080" description:"listen port"`
//	}
//
//	func (c Config) Validate() error { return nil }
//
// Run go generate. It writes config_configulator.go, which has
// ConfigSchema and ConfigPFlagHooks. Then load the config:
//
//	c := configulator.New(ConfigSchema()).
//		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "MYAPP_", Separator: "_"}).
//		WithFile(&configulator.FileOptions{
//			Search:   []string{"config.yaml"},
//			Decoders: configulator.Decoders{".yaml": yaml.Unmarshal},
//		})
//	cpflag.Bind(c, pflag.CommandLine, ConfigPFlagHooks(), nil)
//	pflag.Parse()
//	cfg, err := c.Load()
//
// Now http.port can be set in config.yaml, with MYAPP_HTTP_PORT=9000, or
// with --http.port=9000. --config picks a different file.
//
// # Struct tags
//
//   - name:"key" is the key in files, env and flags. It falls back to the
//     json or yaml tag.
//   - default:"value" is the default. The generator checks it.
//   - description:"text" is the flag help and the description in generated
//     docs.
//   - env:"NAME" and flag:"name" rename this field's part of the env var or
//     flag name. "-" skips that layer.
//   - short:"p" is a pflag shorthand.
//   - required:"true" makes Load fail if no layer sets the field.
//   - secret:"true" hides the value in PrintConfig, errors, generated docs
//     and flag help.
//   - opaque:"true" decodes the field with its UnmarshalText method.
//
// # Flags
//
// The flags/pflag package binds the config to a spf13/pflag FlagSet, and
// flags/std binds it to a standard library flag.FlagSet. Pick one with the
// generator's -flags option. With -flags none there are no flags.
//
// # Errors
//
// Load returns these error types, which errors.As can match:
// ParseError, DecodeError, UnknownKeyError, OpaqueSpellingError,
// MissingFileError, SearchPathError, NoFileFoundError, RequiredError,
// BadEnvOptionsError and FlagConflictError. An error from Validate is
// returned as it is. A bad value in a JSON file is a ParseError inside the
// DecodeError, with the full dotted path of the field.
//
// # Generated docs
//
// The generator can also print a JSON Schema (-schema), a sample config
// file (-sample) or a Markdown table of every option (-markdown). See the
// README for keeping a README table and config.example.yaml up to date.
package configulator
