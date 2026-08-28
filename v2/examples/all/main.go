// The v1 flagship example, ported: YAML (goccy — zero-dependency, off the
// yaml.v2/v3 lineage), env vars, pflag via cobra, and origin reporting.
// The config struct is byte-identical to v1's.
package main

//go:generate go run github.com/USA-RedDragon/configulator/v2/cmd/configulator -type Config

import (
	"context"
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

type Config struct {
	HTTP   HTTPConfig `name:"http"`
	Enable bool       `name:"enable" description:"enable the service"`
}

type HTTPConfig struct {
	Host  string   `name:"host" description:"host to listen on" default:"localhost"`
	Port  int      `name:"port" description:"port to listen on" default:"8080"`
	Stuff []string `name:"stuff" description:"some stuff"`
}

func (c Config) Validate() error {
	if c.HTTP.Port < 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("invalid port")
	}
	if c.HTTP.Host == "" {
		return fmt.Errorf("invalid host")
	}
	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	c, err := configulator.FromContext[Config](cmd.Context())
	if err != nil {
		return err
	}
	cfg, err := c.Load() // a typo'd --config errors here; no os.Stat guard needed
	if err != nil {
		return err
	}
	fmt.Printf("Host: %s\nPort: %d\nEnable: %t\nStuff: %v\n",
		cfg.HTTP.Host, cfg.HTTP.Port, cfg.Enable, cfg.HTTP.Stuff)
	fmt.Println("--- origins ---")
	rep := c.Report()
	for _, p := range rep.Paths() {
		o, _ := rep.Origin(p)
		fmt.Printf("%-12s %-8s %s\n", p, o.Layer, o.Detail)
	}
	return nil
}

func main() {
	cmd := &cobra.Command{Use: "all", RunE: run, SilenceUsage: true}

	c := configulator.New(ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
			Prefix: "MYAPP_", Separator: "_",
		}).
		WithFile(&configulator.FileOptions{
			Search:   []string{"config.yaml"},
			Decoders: configulator.Decoders{".yaml": yaml.Unmarshal},
		})
	cpflag.Bind(c, cmd.Flags(), ConfigPFlagHooks(), nil)

	cmd.SetContext(c.WithContext(context.Background()))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
