// Stdlib flag, no cobra, no pflag: -flags std generates hooks over
// flag.FlagSet. Stdlib flag has no shorthand, no slices, and no
// accumulation — the generator accounts for all three.
package main

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config -flags std

import (
	"flag"
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
)

type Config struct {
	Listen string `name:"listen" default:":8080" description:"listen address"`
	Debug  bool   `name:"debug" description:"debug mode"`
}

func (c Config) Validate() error { return nil }

func main() {
	fs := flag.NewFlagSet("stdflag", flag.ExitOnError)
	c := configulator.New(ConfigSchema()).
		WithFile(&configulator.FileOptions{Search: []string{"config.json"}}).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "APP_"})
	cstd.Bind(c, fs, ConfigStdFlagHooks(), nil)
	_ = fs.Parse(os.Args[1:])

	cfg, err := c.Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("listen=%s debug=%v\n", cfg.Listen, cfg.Debug)
}
