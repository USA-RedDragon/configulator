package main

import (
	"fmt"
	"strconv"

	. "github.com/dave/jennifer/jen"
)

const pkgStdFlag = "github.com/USA-RedDragon/configulator/v2/flags/std"
const stdFlag = "flag"

// emitStdFlagHooks emits -flags=std hooks for the stdlib flag package.
// stdlib flag has no shorthands (short: is an error in this mode) and no
// slice types, so a list is a string flag split on ",", like pflag does.
// It has no Changed either, so Bind finds set flags with flag.Visit.
func (e *emitter) emitStdFlagHooks() {
	n := e.m.TypeName
	e.f.Comment(n + "StdFlagHooks returns the stdlib-flag hooks for " + n + ".")
	e.f.Func().Id(n+"StdFlagHooks").Params().Qual(pkgStdFlag, "Hooks").Index(Id(n)).Block(
		Return(Qual(pkgStdFlag, "Hooks").Index(Id(n)).Values(Dict{
			Id("Register"): Id(lowerFirst(n) + "RegisterStdFlags"),
			Id("Apply"):    Id(lowerFirst(n) + "ApplyStdFlags"),
		})),
	)

	flags := flagFields(e.m.Fields, stdOK)

	name := func(segs []string) *Statement {
		lits := make([]Code, 0, len(segs))
		for _, s := range segs {
			lits = append(lits, Lit(s))
		}
		return Qual("strings", "Join").Call(Index().String().Values(lits...), Id("o").Dot("Separator"))
	}

	names := make([]Code, 0, len(flags))
	for _, ff := range flags {
		names = append(names, name(ff.segs))
	}
	reg := make([]Code, 0, len(flags)+3)
	if len(flags) > 0 {
		reg = append(reg,
			Id("names").Op(":=").Index().String().Values(names...),
			For(List(Id("i"), Id("fn")).Op(":=").Range().Id("names")).Block(
				If(Id("fs").Dot("Lookup").Call(Id("fn")).Op("!=").Nil().Op("||").
					Qual("slices", "Contains").Call(Id("names").Index(Empty(), Id("i")), Id("fn"))).Block(
					Return(Op("&").Qual(pkgCfg, "FlagConflictError").Values(Dict{Id("Flag"): Id("fn"), Id("Existing"): Id("fn")})),
				),
			),
		)
	}
	for i, ff := range flags {
		reg = append(reg, stdRegister(ff.f, Id("names").Index(Lit(i))))
	}
	reg = append(reg, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"RegisterStdFlags").Params(
		Id("fs").Op("*").Qual(stdFlag, "FlagSet"), Id("o").Op("*").Qual(pkgStdFlag, "Options"),
	).Error().Block(reg...)

	app := make([]Code, 0, len(flags)+1)
	for _, ff := range flags {
		app = append(app, e.stdApply(ff, name(ff.segs))...)
	}
	app = append(app, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyStdFlags").Params(
		Id("cfg").Op("*").Id(n), Id("fs").Op("*").Qual(stdFlag, "FlagSet"),
		Id("o").Op("*").Qual(pkgStdFlag, "Options"), Id("isSet").Map(String()).Bool(),
		Id("sep").String(), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(app...)
}

// stdKind maps a field to its stdlib flag register method. Sized ints use
// Int64/Uint64 and are range checked in Apply.
func stdKind(f *Field) (reg string, ok bool) {
	k := f.Kind
	if k == KindPointer {
		k = f.Elem.Kind
	}
	switch k {
	case KindString, KindStdSlot:
		return "String", true
	case KindSliceScalar:
		return "String", listElemOK(f.Elem)
	case KindBool:
		return "Bool", true
	case KindFloat:
		return "Float64", true
	case KindInt, KindDuration:
		return "Int64", true
	case KindUint:
		return "Uint64", true
	default:
		return "", false
	}
}

// stdRegister registers f under name, leaving a secret field's default out
// of the flag so it doesn't show in -help.
func stdRegister(f *Field, name *Statement) Code {
	if f.Secret {
		noDefault := *f
		noDefault.Default = ""
		f = &noDefault
	}
	reg, ok := stdKind(f)
	if !ok {
		return Null()
	}
	var def *Statement
	target := f
	if f.Kind == KindPointer {
		target = f.Elem
	}
	switch target.Kind {
	case KindString, KindStdSlot, KindSliceScalar:
		def = Lit(f.Default)
	case KindBool:
		v, _ := strconv.ParseBool(f.Default)
		def = Lit(v)
	case KindFloat:
		v, _ := strconv.ParseFloat(f.Default, 64)
		def = Lit(v)
	case KindInt:
		v, _ := strconv.ParseInt(f.Default, 10, 64)
		def = Lit(v)
	case KindDuration:
		def = durDefault(f).Assert(Id("int64")) // placeholder; fixed below
	case KindUint:
		v, _ := strconv.ParseUint(f.Default, 10, 64)
		def = Lit(v)
	default:
	}
	if target.Kind == KindDuration {
		reg = "Duration"
		def = durDefault(f)
	}
	return Id("fs").Dot(reg).Call(name, def, Lit(f.Desc))
}

func (e *emitter) stdApply(l leaf, name *Statement) []Code {
	f, path := l.f, l.path
	_, ok := stdKind(f)
	if !ok {
		return nil
	}
	target := f
	if f.Kind == KindPointer {
		elem := *f.Elem
		elem.Secret = f.Secret
		target = &elem
	}
	conv := stdConv(target, path)
	val := convNamed(f.Type, Id("pv"))
	if f.Kind == KindPointer {
		conv = append(conv, Id("pp").Op(":=").Add(convNamed(f.Elem.Type, Id("pv"))))
		val = Op("&").Id("pp")
	}
	conv = append(conv, e.chainAssign(l, Id("sep"), func(t *Statement) []Code { return []Code{t.Op("=").Add(val)} })...)
	conv = append(conv, Id("set").Call(Lit(path), Qual(pkgCfg, "LayerCLI"), Lit("-").Op("+").Id("fn")))
	return []Code{If(
		Id("fn").Op(":=").Add(name), Id("isSet").Index(Id("fn")),
	).Block(conv...)}
}

// stdConv emits the statements that read target's flag value into pv.
func stdConv(target *Field, path string) []Code {
	get := Id("fs").Dot("Lookup").Call(Id("fn")).Dot("Value").Assert(Qual(stdFlag, "Getter")).Dot("Get").Call()
	t := target.Type.Underlying().String()
	switch target.Kind {
	case KindString:
		return []Code{Id("pv").Op(":=").Add(get).Assert(String())}
	case KindSliceScalar:
		code := []Code{Id("raw").Op(":=").Add(get).Assert(String())}
		code = append(code, parseList(target, Qual(pkgImpl, "SplitList").Call(Id("raw"), Lit(",")), Lit(path), Lit("-").Op("+").Id("fn"), Id("raw"))...)
		return append(code, Id("pv").Op(":=").Id("lst"))
	case KindBool:
		return []Code{Id("pv").Op(":=").Add(get).Assert(Bool())}
	case KindFloat:
		if t == "float64" {
			return []Code{Id("pv").Op(":=").Add(get).Assert(Float64())}
		}
		return []Code{
			Id("raw").Op(":=").Add(get).Assert(Float64()),
			If(Qual("math", "Abs").Call(Id("raw")).Op(">").Qual("math", "MaxFloat32").Op("&&").Op("!").Qual("math", "IsInf").Call(Id("raw"), Lit(0))).Block(
				Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
					Id("Path"): Lit(path), Id("Source"): Lit("-").Op("+").Id("fn"),
					Id("Err"): Qual("fmt", "Errorf").Call(Lit("%v overflows "+t), Id("raw")),
				}))),
			Id("pv").Op(":=").Id(t).Call(Id("raw")),
		}
	case KindDuration:
		return []Code{Id("pv").Op(":=").Add(get).Assert(Qual("time", "Duration"))}
	case KindStdSlot:
		conv := append([]Code{Id("raw").Op(":=").Add(get).Assert(String())},
			slotParse(target, path, Lit("-").Op("+").Id("fn"), "raw", "sv")...)
		return append(conv, Id("pv").Op(":=").Add(convNamed(target.Type, Id("sv"))))
	case KindInt:
		raw := Add(get).Assert(Int64())
		switch {
		case target.Bits != 0 && target.Bits != 64:
			return []Code{
				Id("raw").Op(":=").Add(raw),
				If(Id("raw").Op("<").Qual("math", fmt.Sprintf("MinInt%d", target.Bits)).Op("||").
					Id("raw").Op(">").Qual("math", fmt.Sprintf("MaxInt%d", target.Bits))).Block(
					Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
						Id("Path"): Lit(path), Id("Source"): Lit("-").Op("+").Id("fn"),
						Id("Err"): Qual("fmt", "Errorf").Call(Lit("%d overflows "+t), Id("raw")),
					}))),
				Id("pv").Op(":=").Id(t).Call(Id("raw")),
			}
		case t == "int64":
			return []Code{Id("pv").Op(":=").Add(raw)}
		default:
			return []Code{Id("pv").Op(":=").Id(t).Call(Add(raw))}
		}
	case KindUint:
		raw := Add(get).Assert(Uint64())
		switch {
		case target.Bits != 0 && target.Bits != 64:
			return []Code{
				Id("raw").Op(":=").Add(raw),
				If(Id("raw").Op(">").Qual("math", fmt.Sprintf("MaxUint%d", target.Bits))).Block(
					Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
						Id("Path"): Lit(path), Id("Source"): Lit("-").Op("+").Id("fn"),
						Id("Err"): Qual("fmt", "Errorf").Call(Lit("%d overflows "+t), Id("raw")),
					}))),
				Id("pv").Op(":=").Id(t).Call(Id("raw")),
			}
		case t == "uint64":
			return []Code{Id("pv").Op(":=").Add(raw)}
		default:
			return []Code{Id("pv").Op(":=").Id(t).Call(Add(raw))}
		}
	default:
		return nil
	}
}
