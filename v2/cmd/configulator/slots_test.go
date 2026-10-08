package main

import (
	"strings"
	"testing"
)

const slotsFixture = `package fixture

import (
	"fmt"
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
	Addr   netip.Addr     'name:"addr"'
	Peer   *netip.Addr    'name:"peer"'
	Allow  []netip.Prefix 'name:"allow"'
	IP     net.IP         'name:"ip"'
	Level  Verbosity         'name:"level" opaque:"true"'
	Unset  *time.Duration 'name:"unset"'
}

type Verbosity int

func (l *Verbosity) UnmarshalText(b []byte) error {
	if string(b) != "debug" {
		return fmt.Errorf("unknown level %q", b)
	}
	*l = -4
	return nil
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
	fmt.Println(*cfg.Wait, *cfg.Grace, *cfg.Z, cfg.Net.String(), *cfg.Month, cfg.Addr, *cfg.Peer, cfg.Allow, cfg.IP, cfg.Level, cfg.Unset == nil)
}
`

// TestOptionalSlotsAndOpaque loads pointers to slot types (durations,
// complex numbers, stdlib types) and fields decoded with their own
// UnmarshalText from each layer.
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
		cfgJSON:        `{"wait": "1m", "z": 2, "net": "10.0.0.0/8", "month": "July", "addr": "1.2.3.4", "peer": "::1", "allow": ["10.0.0.0/8"], "ip": "1.1.1.1", "level": "debug"}`,
		cfgYAML:        "wait: 1m\nz: 2\nnet: 10.0.0.0/8\nmonth: July\naddr: 1.2.3.4\npeer: '::1'\nallow: [10.0.0.0/8]\nip: 1.1.1.1\nlevel: debug\n",
		"mapping.yaml": "peer:\n  ip: 1.2.3.4\n",
	}
	const want = "1m0s 5s (2+0i) 10.0.0.0/8 July 1.2.3.4 ::1 [10.0.0.0/8] 1.1.1.1 -4 true"
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"json-file", []string{cfgFileJSON}, nil, want},
		{"yaml-file", []string{"CFG_FILE=cfg.yaml"}, nil, want},
		{"env-vars", []string{"FX_WAIT=1m", "FX_Z=2", "FX_NET=10.0.0.0/8", "FX_MONTH=7", "FX_ADDR=1.2.3.4", "FX_PEER=::1", "FX_ALLOW=10.0.0.0/8", "FX_IP=1.1.1.1", "FX_LEVEL=debug"}, nil, want},
		{"flag-args", nil, []string{"--wait=1m", "--z=2", "--net=10.0.0.0/8", "--month=July", "--addr=1.2.3.4", "--peer=::1", "--ip=1.1.1.1", "--level=debug"}, "1m0s 5s (2+0i) 10.0.0.0/8 July 1.2.3.4 ::1 [] 1.1.1.1 -4 true"},
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

const secretErrFixture = `package fixture

import (
	"net"
	"time"
)

type Cfg struct {
	Port  int           'name:"port" secret:"true"'
	Wait  time.Duration 'name:"wait" secret:"true"'
	Opt   *uint8        'name:"opt" secret:"true"'
	List  []int         'name:"list" secret:"true"'
	Net   net.IPNet     'name:"net" secret:"true"'
	Plain int           'name:"plain"'
}

func (Cfg) Validate() error { return nil }
`

// TestSecretValuesNotLeaked checks that no layer's parse error for a
// secret field quotes the value.
func TestSecretValuesNotLeaked(t *testing.T) {
	t.Parallel()
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(widthsMain)
	main := strings.Replace(widthsMain, `	fmt.Println(cfg.I, cfg.I8, cfg.I16, cfg.I32, cfg.I64, cfg.U, cfg.U8, cfg.U16, cfg.U32, cfg.U64, cfg.UP, cfg.F32, cfg.F64, cfg.C64, cfg.C128)
	if cfg.PI8 != nil && cfg.PU64 != nil && cfg.PF32 != nil && cfg.PUP != nil {
		fmt.Println(*cfg.PI8, *cfg.PU64, *cfg.PF32, *cfg.PUP)
	}
`, "\tfmt.Println(cfg.Port)\n", 1)
	stdMain = strings.Replace(stdMain, `	fmt.Println(cfg.I, cfg.I8, cfg.I16, cfg.I32, cfg.I64, cfg.U, cfg.U8, cfg.U16, cfg.U32, cfg.U64, cfg.UP, cfg.F32, cfg.F64, cfg.C64, cfg.C128)
	if cfg.PI8 != nil && cfg.PU64 != nil && cfg.PF32 != nil && cfg.PUP != nil {
		fmt.Println(*cfg.PI8, *cfg.PU64, *cfg.PF32, *cfg.PUP)
	}
`, "\tfmt.Println(cfg.Port)\n", 1)
	files := map[string]string{fixtureFile: bt(secretErrFixture)}
	secretFields := []string{"port", "wait", "opt", "list", "net"}
	for _, field := range secretFields {
		files[field+".json"] = `{"` + field + `": "hunter2"}`
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			m := main
			if mode == flagsStd {
				m = stdMain
			}
			dir, bin := buildFixture(t, files, mode, m)
			for _, field := range secretFields {
				dash := "--"
				if mode == flagsStd {
					dash = "-"
				}
				runs := map[string]struct {
					env  []string
					args []string
				}{
					"an env var": {env: []string{"FX_" + strings.ToUpper(field) + "=hunter2"}},
					"a flag":     {args: []string{dash + field + "=hunter2"}},
					"a file":     {env: []string{"CFG_FILE=" + field + ".json"}},
				}
				for layer, r := range runs {
					res := runBin(t, dir, bin, append(hermeticEnv(t), r.env...), r.args...)
					out := res.stdout + res.stderr
					if !strings.Contains(out, field) || !strings.Contains(out, "redacted") || strings.Contains(out, "hunter2") {
						t.Errorf("%s from %s: want an error naming the field without the value, got %q", field, layer, out)
					}
				}
			}
			res := runBin(t, dir, bin, append(hermeticEnv(t), "FX_PLAIN=oops"))
			if !strings.Contains(res.stdout, `"oops"`) {
				t.Errorf("a field that isn't secret should still quote its value: %q", res.stdout)
			}
		})
	}
}
