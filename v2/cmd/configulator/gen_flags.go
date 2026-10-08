package main

import (
	"fmt"
	"go/types"
	"strconv"
	"time"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitPFlagHooks() {
	n := e.m.TypeName
	e.f.Comment(n + "PFlagHooks returns the pflag hooks for " + n + ".")
	e.f.Func().Id(n+"PFlagHooks").Params().Qual(pkgPFlag, "Hooks").Index(Id(n)).Block(
		Return(Qual(pkgPFlag, "Hooks").Index(Id(n)).Values(Dict{
			Id("Register"): Id(lowerFirst(n) + "RegisterPFlags"),
			Id("Apply"):    Id(lowerFirst(n) + "ApplyPFlags"),
		})),
	)

	flags := flagFields(e.m.Fields, pflagOK)

	flagName := func(segs []string) *Statement {
		lits := make([]Code, 0, len(segs))
		for _, s := range segs {
			lits = append(lits, Lit(s))
		}
		return Qual("strings", "Join").Call(Index().String().Values(lits...), Id("o").Dot("Separator"))
	}

	names := make([]Code, 0, len(flags))
	shorts := make([]Code, 0, len(flags))
	anyShort := false
	for _, ff := range flags {
		names = append(names, flagName(ff.segs))
		shorts = append(shorts, Lit(ff.f.Short))
		anyShort = anyShort || ff.f.Short != ""
	}
	check := []Code{
		If(Id("fs").Dot("Lookup").Call(Id("name")).Op("!=").Nil().Op("||").
			Qual("slices", "Contains").Call(Id("names").Index(Empty(), Id("i")), Id("name"))).Block(
			Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{Id("Flag"): Id("name"), Id("Existing"): Id("name")})),
		),
	}
	var reg []Code
	if len(flags) > 0 {
		reg = append(reg, Id("names").Op(":=").Index().String().Values(names...))
	}
	if anyShort {
		reg = append(reg, Id("shorts").Op(":=").Index().String().Values(shorts...))
		check = append(check, If(
			Id("s").Op(":=").Id("shorts").Index(Id("i")), Id("s").Op("!=").Lit("").Op("&&").Id("fs").Dot("ShorthandLookup").Call(Id("s")).Op("!=").Nil(),
		).Block(
			Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{
				Id("Flag"): Id("name"), Id("Shorthand"): Id("s"),
				Id("Existing"): Id("fs").Dot("ShorthandLookup").Call(Id("s")).Dot("Name"),
			})),
		))
	}
	if len(flags) > 0 {
		reg = append(reg, For(List(Id("i"), Id("name")).Op(":=").Range().Id("names")).Block(check...))
	}
	for i, ff := range flags {
		reg = append(reg, registerFlag(ff.f, Id("names").Index(Lit(i)))...)
	}
	reg = append(reg, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"RegisterPFlags").Params(
		Id("fs").Op("*").Qual(pfl, "FlagSet"), Id("o").Op("*").Qual(pkgPFlag, "Options"),
	).Error().Block(reg...)

	app := make([]Code, 0, len(flags)+1)
	for _, ff := range flags {
		app = append(app, e.applyFlag(ff, flagName(ff.segs)))
	}
	app = append(app, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyPFlags").Params(
		Id("cfg").Op("*").Id(n), Id("fs").Op("*").Qual(pfl, "FlagSet"),
		Id("o").Op("*").Qual(pkgPFlag, "Options"), Id("sep").String(), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(app...)
}

// flagFields returns the leaves an adapter registers as flags, in order:
// those ok accepts, skipping any field tagged flag:"-" with its subtree.
func flagFields(fields []*Field, ok func(*Field) bool) []leaf {
	var out []leaf
	for _, l := range leaves(fields, (*Field).flagSeg, func(f *Field) bool { return f.FlagSkip }) {
		if ok(l.f) {
			out = append(out, l)
		}
	}
	return out
}

func pflagOK(f *Field) bool {
	reg, _, _, ok := pflagTypeOps(f)
	return ok && reg != ""
}

func stdOK(f *Field) bool {
	_, ok := stdKind(f)
	return ok
}

// checkFlagTags rejects flag tags that pflag would panic on or that would
// take over help: a shorthand that isn't one ASCII character, a shorthand
// used twice, short:"h", and a top-level flag named help.
func checkFlagTags(flags []leaf) error {
	shorts := map[string]string{}
	for _, ff := range flags {
		if len(ff.segs) == 1 && ff.segs[0] == "help" {
			return fmt.Errorf("%s: flag --help is reserved for help; rename it with flag:\"name\" or skip it with flag:\"-\"", ff.path)
		}
		s := ff.f.Short
		switch {
		case s == "":
			continue
		case len(s) != 1 || s[0] > 127:
			return fmt.Errorf("%s: short:%q must be a single ASCII character", ff.path, s)
		case s == "h":
			return fmt.Errorf(`%s: short:"h" is reserved for help`, ff.path)
		}
		if prev, ok := shorts[s]; ok {
			return fmt.Errorf("%s: short:%q is also used by %s", ff.path, s, prev)
		}
		shorts[s] = ff.path
	}
	return nil
}

func pflagTypeOps(f *Field) (reg, get string, def *Statement, ok bool) {
	switch f.Kind {
	case KindString:
		d := f.Default
		return "String", "GetString", Lit(d), true
	case KindBool:
		v, _ := strconv.ParseBool(f.Default)
		return "Bool", "GetBool", Lit(v), true
	case KindFloat:
		v, _ := strconv.ParseFloat(f.Default, 64)
		return "Float64", "GetFloat64", Lit(v), true
	case KindDuration:
		return "Duration", "GetDuration", durDefault(f), true
	case KindStdSlot:
		return "String", "GetString", Lit(f.Default), true
	case KindInt:
		v, _ := strconv.ParseInt(f.Default, 10, 64)
		switch f.Bits {
		case 8:
			return "Int8", "GetInt8", Id("int8").Call(Lit(int(v))), true
		case 16:
			return "Int16", "GetInt16", Id("int16").Call(Lit(int(v))), true
		case 32:
			return "Int32", "GetInt32", Id("int32").Call(Lit(int(v))), true
		case 64:
			return "Int64", "GetInt64", Lit(v), true
		}
		return "Int", "GetInt", Lit(int(v)), true
	case KindUint:
		v, _ := strconv.ParseUint(f.Default, 10, 64)
		switch f.Bits {
		case 8:
			return "Uint8", "GetUint8", Id("uint8").Call(intLit(KindUint, f.Default)), true
		case 16:
			return "Uint16", "GetUint16", Id("uint16").Call(intLit(KindUint, f.Default)), true
		case 32:
			return "Uint32", "GetUint32", Id("uint32").Call(intLit(KindUint, f.Default)), true
		case 64:
			return "Uint64", "GetUint64", Lit(v), true
		}
		return "Uint", "GetUint", Lit(uint(v)), true
	case KindPointer:
		// no pflag default for *scalar, so an unset optional stays nil
		reg, get, _, ok := pflagTypeOps(&Field{Kind: f.Elem.Kind, Bits: f.Elem.Bits, Type: f.Elem.Type, Elem: f.Elem.Elem})
		if !ok || f.Elem.Kind == KindStruct {
			return "", "", nil, false
		}
		return reg, get, zeroDefault(f.Elem), true
	case KindSliceScalar:
		if native := pflagNativeSlice(f.Elem); native != "" {
			return native + "Slice", "Get" + native + "Slice", Nil(), true
		}
		return "StringSlice", "GetStringSlice", Nil(), listElemOK(f.Elem)
	default:
		return "", "", nil, false
	}
}

// pflagNativeSlice names the pflag slice type for elem, like "Int" for
// IntSlice, or returns "" when the list is a StringSlice parsed by the
// generated code.
func pflagNativeSlice(elem *Field) string {
	if isNamed(elem.Type, "time", "Duration") {
		return "Duration"
	}
	b, ok := types.Unalias(elem.Type).(*types.Basic)
	if !ok {
		return ""
	}
	switch b.Kind() {
	case types.String:
		return "String"
	case types.Bool:
		return "Bool"
	case types.Int:
		return "Int"
	case types.Int32:
		return "Int32"
	case types.Int64:
		return "Int64"
	case types.Uint:
		return "Uint"
	case types.Float32:
		return "Float32"
	case types.Float64:
		return "Float64"
	default:
		return ""
	}
}

func durDefault(f *Field) *Statement {
	if f.Default == "" {
		return Lit(0)
	}
	d, _ := time.ParseDuration(f.Default)
	return Qual("time", "Duration").Call(Lit(int64(d)))
}

// zeroDefault renders the zero value for a pointer flag's element.
func zeroDefault(f *Field) *Statement {
	switch f.Kind {
	case KindString:
		return Lit("")
	case KindBool:
		return Lit(false)
	case KindFloat:
		return Lit(0.0)
	case KindInt, KindUint:
		t := f.Type.Underlying().String()
		if t == "int" || t == "uint" {
			return Lit(0)
		}
		return Id(t).Call(Lit(0))
	default:
		return Lit(0)
	}
}

// registerFlag registers f under name. A list flag shows its default as
// written in the tag, since the tag is split with the separator only at
// load. A secret field's default is left out of the flag so it doesn't show
// in --help; ApplyDefaults still sets it.
func registerFlag(f *Field, name *Statement) []Code {
	if f.Secret {
		noDefault := *f
		noDefault.Default = ""
		f = &noDefault
	}
	reg, _, def, ok := pflagTypeOps(f)
	if !ok {
		return nil
	}
	call := Id("fs").Dot(reg).Call(name.Clone(), def, Lit(f.Desc))
	if f.Short != "" {
		call = Id("fs").Dot(reg+"P").Call(name.Clone(), Lit(f.Short), def, Lit(f.Desc))
	}
	out := []Code{call}
	if f.Kind == KindSliceScalar && f.Default != "" {
		out = append(out, Id("fs").Dot("Lookup").Call(name.Clone()).Dot("DefValue").Op("=").Lit("["+f.Default+"]"))
	}
	return out
}

func (e *emitter) applyFlag(l leaf, name *Statement) Code {
	f, path := l.f, l.path
	_, get, _, ok := pflagTypeOps(f)
	if !ok {
		return Null()
	}
	source := func() *Statement { return Lit("--").Op("+").Id("n") }
	var prep []Code
	var val *Statement
	switch {
	case f.Kind == KindStdSlot:
		prep, val = slotParse(f, path, source(), "v", "sv"), convNamed(f.Type, Id("sv"))
	case f.Kind == KindSliceScalar && pflagNativeSlice(f.Elem) == "":
		prep = parseList(f, Id("v"), Lit(path), source(), Qual("strings", "Join").Call(Id("v"), Lit(",")))
		val = Id("lst")
	case f.Kind == KindPointer:
		prep, val = []Code{Id("pv").Op(":=").Add(convNamed(f.Elem.Type, Id("v")))}, Op("&").Id("pv")
	default:
		val = convNamed(f.Type, Id("v"))
	}
	body := []Code{
		List(Id("v"), Err()).Op(":=").Id("fs").Dot(get).Call(Id("n")),
		If(Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
				Id("Path"): Lit(path), Id("Source"): source(), Id("Err"): Err(),
			})),
		),
	}
	body = append(body, prep...)
	body = append(body, e.chainAssign(l, Id("sep"), func(t *Statement) []Code { return []Code{t.Op("=").Add(val)} })...)
	body = append(body, Id("set").Call(Lit(path), Qual(pkgCfg, "LayerCLI"), source()))
	return If(
		Id("n").Op(":=").Add(name), Id("fs").Dot("Changed").Call(Id("n")),
	).Block(body...)
}

// slotParse decodes the string variable in through f's wrapper type into a
// new variable out, returning a ParseError on failure.
func slotParse(f *Field, path string, source *Statement, in, out string) []Code {
	val := Id(in)
	if f.Secret {
		val = Lit("(redacted)")
	}
	return []Code{
		Var().Id("slot").Qual(pkgCfg, f.SlotType),
		If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(Id(in))), Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
				Id("Path"): Lit(path), Id("Source"): source, Id("Value"): val, Id("Err"): Err(),
			})),
		),
		List(Id(out), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
	}
}
