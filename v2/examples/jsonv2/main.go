// Zero third-party dependencies: strict encoding/json/v2 (the built-in
// default decoder) plus env. This is the binary-size gate example
// (-flags none keeps pflag entirely out of the link — Gate C).
package main

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config -flags none

import (
	"fmt"
	"time"

	configulator "github.com/USA-RedDragon/configulator/v2"
)

type Config struct {
	Listen  string        `name:"listen" default:":8080"`
	Timeout time.Duration `name:"timeout" default:"30s"`
}

func (c Config) Validate() error { return nil }

func main() {
	cfg, err := configulator.New(ConfigSchema()).
		WithFile(&configulator.FileOptions{Search: []string{"config.json"}}).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "APP_"}).
		Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("listen=%s timeout=%s\n", cfg.Listen, cfg.Timeout)
}
