package pflag_test

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/USA-RedDragon/configulator/v2/internal/exampleconfig"
	"github.com/spf13/pflag"
)

// Bind adds a flag for every config field. Flags override every other
// layer.
func ExampleBind() {
	c := configulator.New(exampleconfig.ConfigSchema()).
		WithEnviron(func(string) (string, bool) { return "", false })

	fs := pflag.NewFlagSet("myapp", pflag.ContinueOnError)
	cpflag.Bind(c, fs, exampleconfig.ConfigPFlagHooks(), nil)
	if err := fs.Parse([]string{"--http.port=9000", "--db.url=postgres://localhost/app"}); err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := c.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.HTTP.Host, cfg.HTTP.Port)
	// Output: localhost 9000
}
