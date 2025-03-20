package main

import (
	"context"
	"fmt"

	"github.com/USA-RedDragon/configulator"
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
		return fmt.Errorf("failed to get config from context")
	}

	cfg, err := c.Load()
	if err != nil {
		return err
	}

	fmt.Printf("Host: %s\n", cfg.HTTP.Host)
	fmt.Printf("Port: %d\n", cfg.HTTP.Port)
	fmt.Printf("Enable: %t\n", cfg.Enable)
	fmt.Printf("Stuff: %v\n", cfg.HTTP.Stuff)

	return nil
}

func main() {
	cmd := &cobra.Command{
		RunE: run,
	}

	c := configulator.New[Config]().
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
			Prefix:    "MYAPP_",
			Separator: "_",
		}).
		WithFile(&configulator.FileOptions{
			Paths: []string{"config.yaml"},
		}).
		WithPFlags(cmd.Flags(), nil).
		WithArraySeparator(",")

	cmd.SetContext(c.WithContext(context.TODO()))

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
