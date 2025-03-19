package main

import (
	"fmt"

	"github.com/USA-RedDragon/configulator"
	"github.com/spf13/cobra"
)

type Config struct {
	HTTP   HTTPConfig `config:"name:http"`
	Enable bool       `config:"enable,description:enable the service"`
}

type HTTPConfig struct {
	Host  string   `config:"host,description:host to listen on,default:localhost"`
	Port  int      `config:"port,description:port to listen on,default:8080"`
	Sub   Sub      `config:"sub"`
	Stuff []string `config:"stuff,description:some stuff"`
}

type Sub struct {
	SubField string `config:"subfield"`
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

func wrapRun(c *configulator.Configulator[Config]) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return run(c, cmd, args)
	}
}

func run(c *configulator.Configulator[Config], _ *cobra.Command, _ []string) error {
	defCfg, err := c.Default()
	if err != nil {
		return err
	}

	fmt.Printf("Default Host: %s\n", defCfg.HTTP.Host)
	fmt.Printf("Default Port: %d\n", defCfg.HTTP.Port)
	fmt.Printf("Default Enable: %t\n", defCfg.Enable)
	fmt.Printf("Default SubField: %s\n", defCfg.HTTP.Sub.SubField)
	fmt.Printf("Default Stuff: %v\n", defCfg.HTTP.Stuff)

	cfg, err := c.Load()
	if err != nil {
		return err
	}

	fmt.Printf("Host: %s\n", cfg.HTTP.Host)
	fmt.Printf("Port: %d\n", cfg.HTTP.Port)
	fmt.Printf("Enable: %t\n", cfg.Enable)
	fmt.Printf("SubField: %s\n", cfg.HTTP.Sub.SubField)
	fmt.Printf("Stuff: %v\n", cfg.HTTP.Stuff)

	return nil
}

func main() {
	cmd := &cobra.Command{}
	c := configulator.NewConfigulator[Config]().
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
			Prefix:    "MYAPP_",
			Separator: "_",
		}).
		WithFile(&configulator.FileOptions{
			Paths:           []string{"config.yaml"},
			Format:          configulator.YAML,
			ErrorIfNotFound: true,
		}).
		WithPFlags(cmd.Flags(), nil).
		WithArraySeparator(",")

	cmd.RunE = wrapRun(c)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
