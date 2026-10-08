// Package examplestd is the config type used by the flags/std examples.
package examplestd

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config -flags std

// Config is a small web service config.
type Config struct {
	Host string `name:"host" default:"localhost" description:"listen address"`
	Port uint16 `name:"port" default:"8080" description:"listen port"`
}

// Validate is called by Load after every layer is applied.
func (c Config) Validate() error { return nil }
