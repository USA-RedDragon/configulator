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
	e.decl().Comment(n + "PFlagHooks returns the pflag hooks for " + n + ".")
	e.f.Func().Id(n+"PFlagHooks").Params().Qual(pkgPFlag, "Hooks").Index(Id(n)).Block(
		Return(Qual(pkgPFlag, "Hooks").Index(Id(n)).Values(Dict{
			Id("Register"): Id(lowerFirst(n) + "RegisterPFlags"),
			Id("Apply"):    Id(lowerFirst(n) + "ApplyPFlags"),
		})),
	)

	flags := flagFields(e.m.Fields, pflagOK)

	names := make([]Code, 0, len(flags))
	shorts := make([]Code, 0, len(flags))
	anyShort := false
	for _, ff := range flags {
		names = append(names, flagNameCode(ff.segs))
		shorts = append(shorts, Lit(ff.f.Short))
		anyShort = anyShort || ff.f.Short != ""
	}
	check := conflictChecks()
	var reg []Code
	if len(flags) > 0 {
		reg = append(reg, Id("names").Op(":=").Index().String().Custom(multiLine(), names...))
	}
	if anyShort {
		reg = append(reg, Id("shorts").Op(":=").Index().String().Values(shorts...))
		check = append(check, If(Id("s").Op(":=").Id("shorts").Index(Id("i")), Id("s").Op("!=").Lit("")).Block(
			If(Id("f").Op(":=").Id("fs").Dot("ShorthandLookup").Call(Id("s")), Id("f").Op("!=").Nil()).Block(
				Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{
					Id("Flag"): Id("name"), Id("Shorthand"): Id("s"), Id("Existing"): Id("f").Dot("Name"),
				})),
			),
		))
	}
	if len(flags) > 0 {
		reg = append(reg, For(List(Id("i"), Id("name")).Op(":=").Range().Id("names")).Block(check...))
	}
	for i, ff := range flags {
		reg = append(reg, registerFlag(ff.f, Id("names").Index(Lit(i)))...)
	}
	reg = append(reg, Return(Nil()))
	e.decl().Func().Id(lowerFirst(n)+"RegisterPFlags").Params(
		Id("fs").Op("*").Qual(pfl, "FlagSet"), Id("o").Op("*").Qual(pkgPFlag, "Options"),
	).Error().Block(reg...)

	app := make([]Code, 0, len(flags)+1)
	for _, ff := range flags {
		app = append(app, e.applyFlag(ff, flagNameCode(ff.segs)))
	}
	app = append(app, Return(Nil()))
	e.decl().Func().Id(lowerFirst(n)+"ApplyPFlags").Params(
		Id("cfg").Op("*").Id(n), Id("fs").Op("*").Qual(pfl, "FlagSet"),
		Id("o").Op("*").Qual(pkgPFlag, "Options"), Id("sep").String(), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(app...)
}

// multiLine renders a composite literal with one element per line.
func multiLine() Options {
	return Options{Open: "{", Close: "}", Separator: ",", Multi: true}
}

// flagNameCode renders a flag name: its segments joined by the separator
// from the adapter options.
func flagNameCode(segs []string) *Statement {
	s := Lit(segs[0])
	for _, seg := range segs[1:] {
		s = s.Op("+").Id("o").Dot("Separator").Op("+").Lit(seg)
	}
	return s
}

// conflictChecks returns the statements that reject the flag name at
// index i of names when the FlagSet or an earlier config field already has
// it. Existing is the name of the flag already holding it, which a pflag
// NormalizeFunc can make different from name.
func conflictChecks() []Code {
	return []Code{
		If(Id("f").Op(":=").Id("fs").Dot("Lookup").Call(Id("name")), Id("f").Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{Id("Flag"): Id("name"), Id("Existing"): Id("f").Dot("Name")})),
		),
		If(Qual("slices", "Contains").Call(Id("names").Index(Empty(), Id("i")), Id("name"))).Block(
			Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{Id("Flag"): Id("name"), Id("Existing"): Id("name")})),
		),
	}
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
		if f.Bits == 32 {
			return "Float32", "GetFloat32", Lit(float32(v)), true
		}
		return "Float64", "GetFloat64", Lit(v), true
	case KindDuration:
		return "Duration", "GetDuration", durDefault(f), true
	case KindStdSlot:
		return "String", "GetString", Lit(f.Default), true
	case KindInt:
		reg := intFlagType("Int", f.Bits)
		return reg, "Get" + reg, intLit(KindInt, f.Default), true
	case KindUint:
		reg := intFlagType("Uint", f.Bits)
		return reg, "Get" + reg, intLit(KindUint, f.Default), true
	case KindPointer:
		// no pflag default for *scalar, so an unset optional stays nil
		reg, get, _, ok := pflagTypeOps(&Field{Kind: f.Elem.Kind, Bits: f.Elem.Bits, Type: f.Elem.Type, SlotType: f.Elem.SlotType, Elem: f.Elem.Elem})
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

// intFlagType returns the pflag type name for an integer of the given
// width: base itself for a plain int or uint, or base followed by bits.
func intFlagType(base string, bits int) string {
	if bits == 0 {
		return base
	}
	return base + strconv.Itoa(bits)
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
	case KindString, KindStdSlot:
		return Lit("")
	case KindBool:
		return Lit(false)
	case KindFloat:
		return Lit(0.0)
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
	switch {
	case secretText(f):
		reg, def = "String", Lit("")
	case f.Secret && f.Kind == KindSliceScalar:
		reg, def = "StringSlice", Nil()
	}
	call := Id("fs").Dot(reg).Call(name.Clone(), def, Lit(f.Desc))
	if f.Short != "" {
		call = Id("fs").Dot(reg+"P").Call(name.Clone(), Lit(f.Short), def, Lit(f.Desc))
	}
	// pflag's own int and uint flags wrap on 32-bit platforms.
	if reg == "Int" || reg == "Uint" {
		value := Qual(pkgImpl, "New"+reg).Call(def)
		call = Id("fs").Dot("Var").Call(value, name.Clone(), Lit(f.Desc))
		if f.Short != "" {
			call = Id("fs").Dot("VarP").Call(value, name.Clone(), Lit(f.Short), Lit(f.Desc))
		}
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
	case secretText(f):
		get = "GetString"
		prep, val = e.textValue(f, path, func() Code { return source() })
	case f.Kind == KindSliceScalar && f.Secret:
		get = "GetStringSlice"
		prep = parseList(f, Id("v"), Lit(path), source(), Qual("strings", "Join").Call(Id("v"), Lit(",")))
		val = Id("lst")
	case f.Kind == KindStdSlot:
		prep, val = slotParse(f, path, source(), "v", "sv"), convNamed(f.Type, Id("sv"))
	case f.Kind == KindSliceScalar && pflagNativeSlice(f.Elem) == "":
		prep = parseList(f, Id("v"), Lit(path), source(), Qual("strings", "Join").Call(Id("v"), Lit(",")))
		val = Id("lst")
	case f.Kind == KindPointer && f.Elem.Kind == KindStdSlot:
		elem := *f.Elem
		elem.Secret = f.Secret
		prep = append(slotParse(&elem, path, source(), "v", "sv"), Id("pv").Op(":=").Add(convNamed(f.Elem.Type, Id("sv"))))
		val = Op("&").Id("pv")
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

// secretText reports whether f is a secret field that a flag adapter takes
// as a string and parses itself, so the flag library's parse errors can't
// echo the value.
func secretText(f *Field) bool {
	if !f.Secret {
		return false
	}
	k := f.Kind
	if k == KindPointer {
		k = f.Elem.Kind
	}
	switch k {
	case KindInt, KindUint, KindFloat, KindDuration:
		return true
	default:
		return false
	}
}

// slotParse decodes the string variable in through f's wrapper type into a
// new variable out, returning a ParseError on failure.
func slotParse(f *Field, path string, source *Statement, in, out string) []Code {
	val, cause := Id(in), Err()
	if f.Secret {
		val, cause = Lit("(redacted)"), Qual(pkgImpl, "Redact").Call(Err(), Id(in))
	}
	return []Code{
		Var().Id("slot").Add(slotCode(f)),
		If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(Id(in))), Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
				Id("Path"): Lit(path), Id("Source"): source, Id("Value"): val, Id("Err"): cause,
			})),
		),
		List(Id(out), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
	}
}
