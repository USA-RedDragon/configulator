package main

import (
	"strings"
	"testing"
)

const textDefaultTypes = `package fixture

import (
	"fmt"
	"net/netip"
)

type Level int

func (l *Level) UnmarshalText(b []byte) error {
	if string(b) != "debug" {
		return fmt.Errorf("unknown level %q", b)
	}
	*l = -4
	return nil
}

type Item struct {
	Addr netip.Addr 'name:"addr" default:"10.0.0.1"'
}

func (Cfg) Validate() error { return nil }
`

const textDefaultMain = `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	fixture "fixture"
)

func main() {
	cfg, err := configulator.New(fixture.CfgSchema()).
		WithFile(&configulator.FileOptions{Search: []string{"cfg.json"}}).
		Load()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(0)
	}
	fmt.Print(cfg.PrintConfig())
}
`

// TestTextDefaults checks that a default: on a type decoded with its own
// UnmarshalText is parsed at load, including element defaults, and that a
// bad one is a ParseError from the default tag that hides a secret.
func TestTextDefaults(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		fields, want, leak string
	}{
		"good": {
			fields: "\tA netip.Addr 'name:\"a\" default:\"1.2.3.4\"'\n\tP *netip.Addr 'name:\"p\" default:\"::1\"'\n" +
				"\tL Level 'name:\"l\" opaque:\"true\" default:\"debug\"'\n\tItems []Item 'name:\"items\"'\n",
			want: "a = 1.2.3.4\np = ::1\nl = -4\nitems = [{10.0.0.1}]",
		},
		"bad": {
			fields: "\tL Level 'name:\"l\" opaque:\"true\" default:\"loud\"'\n",
			want:   `error: l: cannot parse "loud" from default tag: unknown level "loud"`,
		},
		"bad-secret": {
			fields: "\tL Level 'name:\"l\" opaque:\"true\" secret:\"true\" default:\"loud\"'\n",
			want:   `error: l: cannot parse "(redacted)" from default tag`,
			leak:   "loud",
		},
		"bad-element": {
			fields: "\tItems []Item 'name:\"items\"'\n",
			want:   `error: items[0].addr: cannot parse "10.0.0.1x" from default tag`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src := textDefaultTypes + "\ntype Cfg struct {\n" + tc.fields + "}\n"
			if name == "bad-element" {
				src = strings.Replace(src, `default:"10.0.0.1"`, `default:"10.0.0.1x"`, 1)
			}
			files := map[string]string{fixtureFile: bt(src), cfgJSON: `{"items": [{}]}`}
			dir, bin := buildFixture(t, files, flagsNone, textDefaultMain)
			res := runBin(t, dir, bin, hermeticEnv(t))
			got := strings.TrimSpace(res.stdout)
			if res.code != 0 || !strings.HasPrefix(got, tc.want) || (tc.leak != "" && strings.Contains(got, tc.leak)) {
				t.Errorf("got %q (exit %d), want prefix %q\n%s", got, res.code, tc.want, res.stderr)
			}
		})
	}
}
