package configulator_test

import (
	"errors"
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/internal/exampleconfig"
)

const configFile = "config.json"

// env and files stand in for the real environment and file system, so the
// examples give the same output everywhere.
func env(vars map[string]string) configulator.Getenv {
	return func(k string) (string, bool) {
		v, ok := vars[k]
		return v, ok
	}
}

func files(fs map[string]string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		if s, ok := fs[path]; ok {
			return []byte(s), nil
		}
		return nil, os.ErrNotExist
	}
}

// Load a config from defaults, a JSON file and environment variables. A
// later layer overrides an earlier one.
func Example() {
	c := configulator.New(exampleconfig.ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "MYAPP_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{configFile}}).
		WithEnviron(env(map[string]string{
			"MYAPP_HTTP_PORT": "9000",
			"MYAPP_DB_URL":    "postgres://localhost/app",
		})).
		WithReadFile(files(map[string]string{
			configFile: `{"log-level": "debug", "http": {"host": "0.0.0.0", "port": 8000}}`,
		}))

	cfg, err := c.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.LogLevel, cfg.Timeout, cfg.HTTP.Host, cfg.HTTP.Port)
	// Output: debug 30s 0.0.0.0 9000
}

// Report shows where each value came from.
func ExampleConfigulator_Report() {
	c := configulator.New(exampleconfig.ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "MYAPP_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{configFile}}).
		WithEnviron(env(map[string]string{"MYAPP_DB_URL": "postgres://localhost/app"})).
		WithReadFile(files(map[string]string{configFile: `{"http": {"port": 8000}}`}))

	if _, err := c.Load(); err != nil {
		fmt.Println(err)
		return
	}
	rep := c.Report()
	for _, path := range rep.Paths() {
		o, _ := rep.Origin(path)
		fmt.Printf("%s: %s (%s)\n", path, o.Detail, o.Layer)
	}
	// Output:
	// db.url: MYAPP_DB_URL (env)
	// http.host: default tag (default)
	// http.port: config.json (file)
	// log-level: default tag (default)
	// timeout: default tag (default)
}

// Load returns a typed error that errors.As can match. Here db.url is
// tagged required:"true" and nothing sets it.
func ExampleRequiredError() {
	c := configulator.New(exampleconfig.ConfigSchema()).
		WithEnviron(env(nil))

	_, err := c.Load()
	var req *configulator.RequiredError
	if errors.As(err, &req) {
		fmt.Println("missing:", req.Path)
	}
	// Output: missing: db.url
}

// Default applies only the default tags.
func ExampleConfigulator_Default() {
	c := configulator.New(exampleconfig.ConfigSchema())
	cfg, err := c.Default()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.HTTP.Host, cfg.HTTP.Port, cfg.Timeout)
	// Output: localhost 8080 30s
}
