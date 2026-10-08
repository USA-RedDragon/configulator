package main

import (
	"fmt"
	"strconv"

	. "github.com/dave/jennifer/jen"
)

const pkgStdFlag = "github.com/USA-RedDragon/configulator/v2/flags/std"
const stdFlag = "flag"

// emitStdFlagHooks emits -flags=std hooks for the stdlib flag package.
// stdlib flag has no shorthands (short: is an error in this mode), no
// slice types (skipped like collections) and no Changed, so Bind finds
// set flags with flag.Visit.
func (e *emitter) emitStdFlagHooks() {
	n := e.m.TypeName
	e.f.Comment(n + "StdFlagHooks returns the stdlib-flag hooks for " + n + ".")
	e.f.Func().Id(n+"StdFlagHooks").Params().Qual(pkgStdFlag, "Hooks").Index(Id(n)).Block(
		Return(Qual(pkgStdFlag, "Hooks").Index(Id(n)).Values(Dict{
			Id("Register"): Id(lowerFirst(n) + "RegisterStdFlags"),
			Id("Apply"):    Id(lowerFirst(n) + "ApplyStdFlags"),
		})),
	)

	type flagField struct {
		f    *Field
		segs []string
		path string
	}
	var flags []flagField
	var walk func(fields []*Field, segs []string, path string)
	walk = func(fields []*Field, segs []string, path string) {
		for _, f := range fields {
			s2 := append(append([]string{}, segs...), f.flagSeg())
			p2 := joinPath(path, f.Tag)
			switch f.Kind {
			case KindStruct:
				walk(f.Fields, s2, p2)
			case KindSliceStruct, KindMapStruct, KindMapScalar, KindSliceScalar:
				// can't be set from flags (stdlib flag has no slice types)
			default:
				if !f.FlagSkip {
					flags = append(flags, flagField{f, s2, p2})
				}
			}
		}
	}
	walk(e.m.Fields, nil, "")

	name := func(segs []string) *Statement {
		lits := make([]Code, 0, len(segs))
		for _, s := range segs {
			lits = append(lits, Lit(s))
		}
		return Qual("strings", "Join").Call(Index().String().Values(lits...), Id("o").Dot("Separator"))
	}

	reg := make([]Code, 0, len(flags)+2)
	lookups := make([]Code, 0, len(flags))
	for _, ff := range flags {
		lookups = append(lookups, name(ff.segs))
	}
	reg = append(reg, For(List(Id("_"), Id("fn")).Op(":=").Range().Index().String().Values(lookups...)).Block(
		If(Id("fs").Dot("Lookup").Call(Id("fn")).Op("!=").Nil()).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("flag -%s already registered on this FlagSet"), Id("fn"))),
		),
	))
	for _, ff := range flags {
		reg = append(reg, stdRegister(ff.f, name(ff.segs)))
	}
	reg = append(reg, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"RegisterStdFlags").Params(
		Id("fs").Op("*").Qual(stdFlag, "FlagSet"), Id("o").Op("*").Qual(pkgStdFlag, "Options"),
	).Error().Block(reg...)

	app := make([]Code, 0, len(flags)+1)
	for _, ff := range flags {
		app = append(app, stdApply(ff.f, ff.path, name(ff.segs))...)
	}
	app = append(app, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyStdFlags").Params(
		Id("cfg").Op("*").Id(n), Id("fs").Op("*").Qual(stdFlag, "FlagSet"),
		Id("o").Op("*").Qual(pkgStdFlag, "Options"), Id("isSet").Map(String()).Bool(),
		Id("set").Qual(pkgCfg, "SetOrigin"),
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

func stdRegister(f *Field, name *Statement) Code {
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
	case KindString, KindStdSlot:
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

func stdApply(f *Field, path string, name *Statement) []Code {
	_, ok := stdKind(f)
	if !ok {
		return nil
	}
	target := f
	if f.Kind == KindPointer {
		target = f.Elem
	}
	assignTo := func(v *Statement) *Statement {
		if f.Kind == KindPointer {
			return cfgSel("cfg", f).Op("=").Op("&").Id("pv")
		}
		_ = v
		return cfgSel("cfg", f).Op("=").Add(convNamed(f.Type, Id("pv")))
	}
	conv := stdConv(target, path)
	conv = append(conv,
		assignTo(nil),
		Id("set").Call(Lit(path), Qual(pkgCfg, "LayerCLI"), Lit("-").Op("+").Id("fn")),
	)
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
	case KindBool:
		return []Code{Id("pv").Op(":=").Add(get).Assert(Bool())}
	case KindFloat:
		if t == "float64" {
			return []Code{Id("pv").Op(":=").Add(get).Assert(Float64())}
		}
		return []Code{Id("pv").Op(":=").Id(t).Call(Add(get).Assert(Float64()))}
	case KindDuration:
		return []Code{Id("pv").Op(":=").Add(get).Assert(Qual("time", "Duration"))}
	case KindStdSlot:
		conv := append([]Code{Id("raw").Op(":=").Add(get).Assert(String())},
			slotParse(target, path, Lit("-").Op("+").Id("fn"), "raw", "sv")...)
		return append(conv, Id("pv").Op(":=").Add(Id("sv")))
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
