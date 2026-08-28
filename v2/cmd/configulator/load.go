package main

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// loadPackage type-checks the package in dir and returns the named config
// type plus the package it lives in (the output package). env overrides
// the environment for the load; nil inherits the process environment
// (tests pass a hermetic one: GOWORK=off + file:// proxy).
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
	// Tolerate two classes of type error that this generation exists to
	// fix — refusing to run would leave no way to generate out of them:
	//   1. any error in a previously generated *_configulator.go (a
	//      schema-signature change makes stale output a type error), and
	//   2. "undefined:" errors anywhere (a first-time consumer writes
	//      main.go calling ConfigSchema() before it exists).
	// Everything else in the user's own code still fails loudly; if an
	// undefined symbol was not one of ours, the compile error returns
	// immediately after generation.
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
