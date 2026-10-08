//go:build goexperiment.jsonv2

package std_test

import (
	"flag"
	"testing"

	cstd "github.com/USA-RedDragon/configulator/v2/flags/std"
)

func TestGet(t *testing.T) {
	t.Parallel()
	fs := flag.NewFlagSet("x", flag.ContinueOnError)
	fs.Int64("n", 7, "")
	if got := cstd.Get[int64](fs, "n"); got != 7 {
		t.Fatalf("Get = %d, want 7", got)
	}
	for _, name := range []string{"n", "missing"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Get[string](%q) did not panic", name)
				}
			}()
			cstd.Get[string](fs, name)
		}()
	}
}
