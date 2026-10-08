// Package exampleconfig is the config type used by the package examples.
package exampleconfig

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config

import "time"

// Config is a small web service config.
type Config struct {
	LogLevel string        `name:"log-level" default:"info" description:"log level"`
	Timeout  time.Duration `name:"timeout" default:"30s" description:"request timeout"`
	HTTP     HTTP          `name:"http"`
	DB       DB            `name:"db"`
}

// HTTP is the listener config.
type HTTP struct {
	Host string `name:"host" default:"localhost" description:"listen address"`
	Port uint16 `name:"port" default:"8080" description:"listen port"`
}

// DB is the database config.
type DB struct {
	URL string `name:"url" required:"true" secret:"true" description:"database URL"`
}

// Validate is called by Load after every layer is applied.
func (c Config) Validate() error { return nil }
