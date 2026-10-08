package main

import (
	"strings"
	"testing"
)

const slotsFixture = `package fixture

import (
	"net"
	"net/netip"
	"time"
)

type Cfg struct {
	Wait   *time.Duration 'name:"wait"'
	Grace  *time.Duration 'name:"grace" default:"5s"'
	Z      *complex128    'name:"z"'
	Net    *net.IPNet     'name:"net"'
	Month  *time.Month    'name:"month" default:"March"'
	Addr   netip.Addr     'name:"addr" opaque:"true"'
	Peer   *netip.Addr    'name:"peer" opaque:"true"'
	Allow  []netip.Prefix 'name:"allow" opaque:"true"'
	Unset  *time.Duration 'name:"unset"'
}

func (Cfg) Validate() error { return nil }
`

const slotsMain = `package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{os.Getenv("CFG_FILE")}, Decoders: configulator.Decoders{".json": configulator.StrictJSON, ".yaml": yaml.Unmarshal}})
	cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println("parse:", err)
		return
	}
	cfg, err := c.Load()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if cfg.Wait == nil || cfg.Z == nil || cfg.Net == nil || cfg.Peer == nil {
		fmt.Println("unset", cfg.Wait, cfg.Z, cfg.Net, cfg.Peer)
		return
	}
	fmt.Println(*cfg.Wait, *cfg.Grace, *cfg.Z, cfg.Net.String(), *cfg.Month, cfg.Addr, *cfg.Peer, cfg.Allow, cfg.Unset == nil)
}
`

// TestOptionalSlotsAndOpaque loads pointers to slot types (durations,
// complex numbers, stdlib types) and opaque:"true" fields from each layer.
func TestOptionalSlotsAndOpaque(t *testing.T) {
	t.Parallel()
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(slotsMain)
	files := map[string]string{
		fixtureFile:    bt(slotsFixture),
		cfgJSON:        `{"wait": "1m", "z": 2, "net": "10.0.0.0/8", "month": "July", "addr": "1.2.3.4", "peer": "::1", "allow": ["10.0.0.0/8"]}`,
		cfgYAML:        "wait: 1m\nz: 2\nnet: 10.0.0.0/8\nmonth: July\naddr: 1.2.3.4\npeer: '::1'\nallow: [10.0.0.0/8]\n",
		"mapping.yaml": "peer:\n  ip: 1.2.3.4\n",
	}
	const want = "1m0s 5s (2+0i) 10.0.0.0/8 July 1.2.3.4 ::1 [10.0.0.0/8] true"
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"json-file", []string{cfgFileJSON}, nil, want},
		{"yaml-file", []string{"CFG_FILE=cfg.yaml"}, nil, want},
		{"env-vars", []string{"FX_WAIT=1m", "FX_Z=2", "FX_NET=10.0.0.0/8", "FX_MONTH=7", "FX_ADDR=1.2.3.4", "FX_PEER=::1", "FX_ALLOW=10.0.0.0/8"}, nil, want},
		{"flag-args", nil, []string{"--wait=1m", "--z=2", "--net=10.0.0.0/8", "--month=July", "--addr=1.2.3.4", "--peer=::1"}, "1m0s 5s (2+0i) 10.0.0.0/8 July 1.2.3.4 ::1 [] true"},
		{"bad-env", []string{"FX_PEER=nope"}, nil, `error: peer: cannot parse "nope" from FX_PEER`},
		{"yaml-mapping", []string{"CFG_FILE=mapping.yaml"}, nil, "error: decoding mapping.yaml"},
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			main := slotsMain
			if mode == flagsStd {
				main = stdMain
			}
			dir, bin := buildFixture(t, files, mode, main)
			for _, tc := range cases {
				args := tc.args
				if mode == flagsStd {
					args = nil
					for _, a := range tc.args {
						args = append(args, strings.TrimPrefix(a, "-"))
					}
				}
				res := runBin(t, dir, bin, append(hermeticEnv(t), tc.env...), args...)
				if got := strings.TrimSpace(res.stdout); res.code != 0 || !strings.HasPrefix(got, tc.want) {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
				}
			}
		})
	}
}

func TestOpaqueGenerateErrors(t *testing.T) {
	t.Parallel()
	for name, field := range map[string]string{
		"default":       "A netip.Addr `name:\"a\" opaque:\"true\" default:\"1.2.3.4\"`",
		"no-unmarshal":  "A struct{ X int } `name:\"a\" opaque:\"true\"`",
		"named-no-text": "A Plain `name:\"a\" opaque:\"true\"`",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src := "package fixture\n\nimport \"net/netip\"\n\nvar _ netip.Addr\n\ntype Plain struct{ X int }\n\ntype Cfg struct {\n\t" + field + "\n}\n" + validateStub
			if _, err := buildFixtureModel(t, src, false); err == nil {
				t.Fatal("want a generate-time error")
			}
		})
	}
}
