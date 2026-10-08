package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

const platformIntFixture = `package fixture

type Cfg struct {
	N int            'name:"n"'
	U uint           'name:"u"'
	L []int          'name:"l"'
	M map[string]int 'name:"m"'
	P *int           'name:"p"'
	D []uint         'name:"d" default:"1,2"'
}

func (Cfg) Validate() error { return nil }
`

const platformIntMain = `package main

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
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Prefix: "FX_"}).
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
	fmt.Println(cfg.N, cfg.U, cfg.L, cfg.M, cfg.D)
	if cfg.P != nil {
		fmt.Println(*cfg.P)
	}
}
`

// TestPlatformIntOn32Bit builds a loader for GOARCH=386 and checks that a
// plain int or uint too big for 32 bits is an error from every layer
// instead of wrapping. It is skipped where the host can't run 386
// binaries.
func TestPlatformIntOn32Bit(t *testing.T) {
	t.Parallel()
	const big = "5000000000"
	stdMain := strings.NewReplacer(
		`cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"`, `cstd "github.com/USA-RedDragon/configulator/v2/flags/std"`,
		`"github.com/spf13/pflag"`, `"flag"`,
		"pflag.NewFlagSet(\"x\", pflag.ContinueOnError)", "flag.NewFlagSet(\"x\", flag.ContinueOnError)",
		"cpflag.Bind(c, fs, fixture.CfgPFlagHooks(), nil)", "cstd.Bind(c, fs, fixture.CfgStdFlagHooks(), nil)",
	).Replace(platformIntMain)
	files := map[string]string{
		fixtureFile: bt(platformIntFixture),
		"n.json":    `{"n": ` + big + `}`,
		"u.json":    `{"u": ` + big + `}`,
		"l.json":    `{"l": [1, ` + big + `]}`,
		"m.json":    `{"m": {"a": ` + big + `}}`,
		"ok.json":   `{"n": -7, "l": [1], "m": {"a": 2}}`,
	}
	cases := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{"in-range", []string{"CFG_FILE=ok.json", "FX_U=4294967295"}, nil, "-7 4294967295 [1] map[a:2] [1 2]"},
		{"env-int", []string{"FX_N=" + big}, nil, "error: n:"},
		{"env-uint", []string{"FX_U=" + big}, nil, "error: u:"},
		{"env-list", []string{"FX_L=1," + big}, nil, "error: l:"},
		{"env-pointer", []string{"FX_P=" + big}, nil, "error: p:"},
		{"file-int", []string{"CFG_FILE=n.json"}, nil, `error: n: cannot parse "5000000000" from n.json`},
		{"file-uint", []string{"CFG_FILE=u.json"}, nil, `error: u: cannot parse "5000000000" from u.json`},
		{"file-list", []string{"CFG_FILE=l.json"}, nil, `error: l[1]: cannot parse "5000000000" from l.json`},
		{"file-map", []string{"CFG_FILE=m.json"}, nil, `error: m.a: cannot parse "5000000000" from m.json`},
		{"flag-int", nil, []string{"--n=" + big}, wantFail},
		{"flag-uint", nil, []string{"--u=" + big}, wantFail},
		{"flag-pointer", nil, []string{"--p=" + big}, wantFail},
	}
	for _, mode := range []string{flagsPFlag, flagsStd} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			main := platformIntMain
			if mode == flagsStd {
				main = stdMain
			}
			dir, bin := build386(t, files, mode, main)
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
				if tc.want == wantFail {
					ok = strings.HasPrefix(got, "parse:") || strings.HasPrefix(got, "error:")
				}
				if res.code != 0 || !ok {
					t.Errorf("%s: got %q (exit %d), want %q\n%s", tc.name, got, res.code, tc.want, res.stderr)
				}
			}
		})
	}
}

// build386 is buildFixture for GOARCH=386. It skips the test when the host
// can't run the binary.
func build386(t *testing.T, files map[string]string, flagsMode, main string) (dir, bin string) {
	t.Helper()
	all := map[string]string{"cmd/run/main.go": main}
	for k, v := range files {
		all[k] = v
	}
	dir = writeModule(t, all)
	named, outPkg, err := loadPackage(dir, "Cfg", hermeticEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildModel(named, outPkg, false)
	if err != nil {
		t.Fatal(err)
	}
	out, err := emit(m, flagsMode)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cfg_configulator.go"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	bin = filepath.Join(dir, "run.bin")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", bin, "./cmd/run")
	cmd.Dir = dir
	cmd.Env = append(hermeticEnv(t), "GOARCH=386", "CGO_ENABLED=0")
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, o)
	}
	probe := exec.CommandContext(t.Context(), bin, "--help")
	probe.Dir = dir
	if err := probe.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || errors.Is(err, syscall.ENOEXEC) {
			t.Skipf("can't run 386 binaries here: %v", err)
		}
	}
	return dir, bin
}
