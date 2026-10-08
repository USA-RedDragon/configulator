// TOML config files via pelletier/go-toml/v2.
package main

//go:generate go tool configulator -type Config -flags none

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DB DBConfig `name:"db"`
}

type DBConfig struct {
	URL  string `name:"url" default:"postgres://localhost/app"`
	Pool uint16 `name:"pool" default:"10"`
}

func (c Config) Validate() error { return nil }

func main() {
	cfg, err := configulator.New(ConfigSchema()).
		WithFile(&configulator.FileOptions{
			Search:   []string{"config.toml"},
			Decoders: configulator.Decoders{".toml": toml.Unmarshal},
		}).
		Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("db=%s pool=%d\n", cfg.DB.URL, cfg.DB.Pool)
}
