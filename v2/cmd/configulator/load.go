package main

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// loadPackage type-checks the package in dir and returns the named config
// type and its package, which is also the output package. env replaces the
// environment for the load; nil uses the process environment. Tests pass
// GOWORK=off and a file:// proxy.
func loadPackage(dir, typeName string, env []string) (*types.Named, *types.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax |
			packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Dir: dir,
		Env: env,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, nil, err
	}
	// Tolerate errors that generating fixes: stale *_configulator.go
	// output, and "undefined:" (main.go calling ConfigSchema() before the
	// first generate).
	fatal := 0
	for _, p := range pkgs {
		for _, e := range p.Errors {
			file := strings.SplitN(e.Pos, ":", 2)[0]
			if strings.HasSuffix(file, "_configulator.go") {
				continue
			}
			if strings.Contains(e.Msg, "undefined: ") {
				continue
			}
			fmt.Fprintln(os.Stderr, e)
			fatal++
		}
	}
	if fatal > 0 {
		return nil, nil, fmt.Errorf("package %s has type errors outside generated files", dir)
	}
	if len(pkgs) != 1 {
		return nil, nil, fmt.Errorf("expected one package in %s, got %d", dir, len(pkgs))
	}
	pkg := pkgs[0]
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, nil, fmt.Errorf("type %s not found in package %s", typeName, pkg.PkgPath)
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, nil, fmt.Errorf("%s is not a named type", typeName)
	}
	return named, pkg.Types, nil
}
