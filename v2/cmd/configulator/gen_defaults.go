package main

import (
	"go/types"
	"strconv"
	"strings"
	"time"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitApplyDefaults() {
	n := e.m.TypeName
	var body []Code
	var walk func(fields []*Field, pathPrefix []string)
	walk = func(fields []*Field, pathPrefix []string) {
		for _, f := range fields {
			path := strings.Join(append(append([]string{}, pathPrefix...), f.Tag), ".")
			if f.Kind == KindStruct {
				walk(f.Fields, append(pathPrefix, f.Tag))
				continue
			}
			if f.Default == "" {
				continue
			}
			body = append(body, e.defaultAssign(f, path)...)
		}
	}
	walk(e.m.Fields, nil)
	body = append(body, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyDefaults").Params(
		Id("cfg").Op("*").Id(n), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(body...)
}

func (e *emitter) defaultAssign(f *Field, path string) []Code {
	rec := Id("set").Call(Lit(path), Qual(pkgCfg, "LayerDefault"), Lit("default tag"))
	switch f.Kind {
	case KindString:
		return []Code{cfgSel("cfg", f).Op("=").Lit(f.Default), rec}
	case KindBool:
		v, _ := strconv.ParseBool(f.Default)
		return []Code{cfgSel("cfg", f).Op("=").Lit(v), rec}
	case KindInt:
		v, _ := strconv.ParseInt(f.Default, 10, 64)
		return []Code{cfgSel("cfg", f).Op("=").Add(castLit(f, int(v))), rec}
	case KindUint:
		v, _ := strconv.ParseUint(f.Default, 10, 64)
		return []Code{cfgSel("cfg", f).Op("=").Add(castLit(f, int(v))), rec}
	case KindFloat:
		v, _ := strconv.ParseFloat(f.Default, 64)
		return []Code{cfgSel("cfg", f).Op("=").Lit(v), rec}
	case KindDuration:
		d, _ := time.ParseDuration(f.Default)
		return []Code{
			cfgSel("cfg", f).Op("=").Qual("time", "Duration").Call(Lit(int64(d))).
				Commentf("// %s, parsed at generate time from %q", d, f.Default),
			rec,
		}
	case KindSliceScalar:
		elems := strings.Split(f.Default, ",")
		var lits []Code
		for _, el := range elems {
			lits = append(lits, Lit(el))
		}
		return []Code{cfgSel("cfg", f).Op("=").Add(fieldGoType(f.Type.Underlying())).Values(lits...), rec}
	case KindPointer:
		tmp := lowerFirst(goName(f)) + "Default"
		inner := *f.Elem
		inner.GoName = tmp
		assign := e.defaultAssign(&inner, path)
		// first statement assigns Id(tmp)... rebuild as var decl
		switch f.Elem.Kind {
		case KindString:
			return []Code{
				Id(tmp).Op(":=").Lit(f.Default),
				cfgSel("cfg", f).Op("=").Op("&").Id(tmp),
				rec,
			}
		}
		_ = assign
		panic("default on pointer to non-string not yet supported")
	}
	panic("defaultAssign: unhandled kind for " + path)
}

// castLit renders a possibly-cast integer literal for sized types.
func castLit(f *Field, v int) *Statement {
	t := f.Type.Underlying().(*types.Basic)
	if t.Kind() == types.Int {
		return Lit(v)
	}
	return Id(t.Name()).Call(Lit(v))
}
