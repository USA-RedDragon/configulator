package impl_test

import (
	"flag"
	"testing"

	"github.com/USA-RedDragon/configulator/v2/impl"
	"github.com/spf13/pflag"
)

// TestIntValues checks the int and uint flag values the generated code
// registers for plain int and uint fields.
func TestIntValues(t *testing.T) {
	t.Parallel()
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	fs.Var(impl.NewInt(-3), "i", "")
	fs.Var(impl.NewUint(3), "u", "")
	if err := fs.Parse([]string{"--i=0x10", "--u=7"}); err != nil {
		t.Fatal(err)
	}
	i, err := fs.GetInt("i")
	if err != nil || i != 16 {
		t.Errorf("GetInt = %d, %v; want 16", i, err)
	}
	u, err := fs.GetUint("u")
	if err != nil || u != 7 {
		t.Errorf("GetUint = %d, %v; want 7", u, err)
	}
	if fs.Lookup("i").DefValue != "-3" || fs.Lookup("u").DefValue != "3" {
		t.Errorf("defaults %q and %q, want -3 and 3", fs.Lookup("i").DefValue, fs.Lookup("u").DefValue)
	}
	for _, arg := range []string{"--i=x", "--u=-1"} {
		if err := fs.Parse([]string{arg}); err == nil {
			t.Errorf("%s: want a parse error", arg)
		}
	}
}

func TestFlagValue(t *testing.T) {
	t.Parallel()
	fs := flag.NewFlagSet("x", flag.ContinueOnError)
	fs.Int64("n", 7, "")
	if got := impl.FlagValue[int64](fs, "n"); got != 7 {
		t.Fatalf("FlagValue = %d, want 7", got)
	}
	for _, name := range []string{"n", "missing"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("FlagValue[string](%q) did not panic", name)
				}
			}()
			impl.FlagValue[string](fs, name)
		}()
	}
}
