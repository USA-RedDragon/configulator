package main

import (
	"strings"
	"testing"
)

const widthsFixture = `package fixture

type Cfg struct {
	I    int        'name:"i" default:"-1"'
	I8   int8       'name:"i8" default:"-8"'
	I16  int16      'name:"i16" default:"-16"'
	I32  int32      'name:"i32" default:"-32"'
	I64  int64      'name:"i64" default:"-64"'
	U    uint       'name:"u" default:"1"'
	U8   uint8      'name:"u8" default:"8"'
	U16  uint16     'name:"u16" default:"16"'
	U32  uint32     'name:"u32" default:"32"'
	U64  uint64     'name:"u64" default:"64"'
	UP   uintptr    'name:"up" default:"7"'
	F32  float32    'name:"f32" default:"3.5"'
	F64  float64    'name:"f64" default:"6.5"'
	C64  complex64  'name:"c64" default:"1+2i"'
	C128 complex128 'name:"c128" default:"3+4i"'
	PI8  *int8      'name:"pi8"'
	PU64 *uint64    'name:"pu64"'
	PF32 *float32   'name:"pf32"'
	PUP  *uintptr   'name:"pup"'
}

func (Cfg) Validate() error { return nil }
`

const widthsMain = `package main

import (
	"fmt"
	"os"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/spf13/pflag"
	fixture "fixture"
)

func main() {
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	c := configulator.New(fixture.CfgSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_", Separator: "_"}).
		WithFile(&configulator.FileOptions{Search: []string{os.Getenv("CFG_FILE")}})
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
	fmt.Println(cfg.I, cfg.I8, cfg.I16, cfg.I32, cfg.I64, cfg.U, cfg.U8, cfg.U16, cfg.U32, cfg.U64, cfg.UP, cfg.F32, cfg.F64, cfg.C64, cfg.C128)
	if cfg.PI8 != nil && cfg.PU64 != nil && cfg.PF32 != nil && cfg.PUP != nil {
		fmt.Println(*cfg.PI8, *cfg.PU64, *cfg.PF32, *cfg.PUP)
	}
}
`

// TestScalarWidths loads every scalar width, and pointers to some, from
// each layer with both flag adapters.
func TestScalarWidths(t *testing.T) {
	t.Parallel()
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(widthsMain)
	files := map[string]string{
		fixtureFile: bt(widthsFixture),
		cfgJSON: `{"i": 1, "i8": 2, "i16": 3, "i32": 4, "i64": 5, "u": 6, "u8": 7, "u16": 8, "u32": 9, "u64": 10, "up": 11,
			"f32": 1.25, "f64": 2.5, "c64": "1i", "c128": "2", "pi8": -1, "pu64": 1, "pf32": 0.5, "pup": 2}`,
	}
	const fromFile = "1 2 3 4 5 6 7 8 9 10 11 1.25 2.5 (0+1i) (2+0i)\n-1 1 0.5 2"
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"from-defaults", nil, nil, "-1 -8 -16 -32 -64 1 8 16 32 64 7 3.5 6.5 (1+2i) (3+4i)"},
		{"from-file", []string{"CFG_FILE=cfg.json"}, nil, fromFile},
		{"from-env", []string{"FX_I=1", "FX_I8=2", "FX_I16=3", "FX_I32=4", "FX_I64=5", "FX_U=6", "FX_U8=7", "FX_U16=8", "FX_U32=9", "FX_U64=10", "FX_UP=11",
			"FX_F32=1.25", "FX_F64=2.5", "FX_C64=1i", "FX_C128=2", "FX_PI8=-1", "FX_PU64=1", "FX_PF32=0.5", "FX_PUP=2"}, nil, fromFile},
		{"from-flags", nil, []string{"--i=1", "--i8=2", "--i16=3", "--i32=4", "--i64=5", "--u=6", "--u8=7", "--u16=8", "--u32=9", "--u64=10", "--up=11",
			"--f32=1.25", "--f64=2.5", "--c64=1i", "--c128=2", "--pi8=-1", "--pu64=1", "--pf32=0.5", "--pup=2"}, fromFile},
		{"env-f32-overflow", []string{"FX_F32=1e300"}, nil, `error: f32: cannot parse "1e300" from FX_F32`},
		{"flag-f32-overflow", nil, []string{"--f32=1e300"}, "fail"},
		{"env-i8-overflow", []string{"FX_I8=300"}, nil, `error: i8: cannot parse "300" from FX_I8`},
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			main := widthsMain
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
				got := strings.TrimSpace(res.stdout)
				ok := strings.HasPrefix(got, tc.want)
				if tc.want == "fail" {
					// pflag rejects the value while parsing, std flag at Load.
					ok = (strings.HasPrefix(got, "parse:") || strings.HasPrefix(got, "error:")) && strings.Contains(got, "f32")
				}
				if res.code != 0 || !ok {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
				}
			}
		})
	}
}
