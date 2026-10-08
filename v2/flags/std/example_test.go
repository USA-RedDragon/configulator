package std_test

import (
	"flag"
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
	"github.com/USA-RedDragon/configulator/v2/internal/examplestd"
)

// Bind adds a flag for every config field. Generate the code with
// -flags std to get the ConfigStdFlagHooks function.
func ExampleBind() {
	c := configulator.New(examplestd.ConfigSchema()).
		WithEnviron(func(string) (string, bool) { return "", false })

	fs := flag.NewFlagSet("myapp", flag.ContinueOnError)
	cstd.Bind(c, fs, examplestd.ConfigStdFlagHooks(), nil)
	if err := fs.Parse([]string{"-port=9000"}); err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := c.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.Host, cfg.Port)
	// Output: localhost 9000
}
