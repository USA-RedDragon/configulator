// The smallest possible consumer: defaults + env, nothing else.
package main

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config -flags none

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
)

type Config struct {
	Name string `name:"name" default:"minimal"`
	Port uint16 `name:"port" default:"8080"`
}

func (c Config) Validate() error { return nil }

func main() {
	cfg, err := configulator.New(ConfigSchema()).
		WithEnvironmentVariables(nil).
		Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s:%d\n", cfg.Name, cfg.Port)
}
