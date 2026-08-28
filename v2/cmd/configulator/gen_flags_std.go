package main

import (
	"fmt"
	"strconv"

	. "github.com/dave/jennifer/jen"
)

const pkgStdFlag = "github.com/USA-RedDragon/configulator/v2/flags/std"
const stdFlag = "flag"

// emitStdFlagHooks emits -flags=std hooks over stdlib flag. Stdlib flag
// has no shorthand (short: is a generate-time error under this mode), no
// slice types (slice fields are skipped with a comment, like collections),
// and no Changed — presence comes from the isSet map Bind builds via
// flag.Visit.
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
	var skipped []string
	var walk func(fields []*Field, segs []string, path string)
	walk = func(fields []*Field, segs []string, path string) {
		for _, f := range fields {
			s2 := append(append([]string{}, segs...), f.Tag)
			p2 := joinPath(path, f.Tag)
			switch f.Kind {
			case KindStruct:
				walk(f.Fields, s2, p2)
			case KindSliceStruct, KindMapStruct, KindMapScalar:
				// collections: skipped for CLI (SPEC rule 6)
			case KindSliceScalar:
				skipped = append(skipped, p2+" (stdlib flag has no slice types)")
			default:
				if !f.FlagSkip {
					flags = append(flags, flagField{f, s2, p2})
				}
			}
		}
	}
	walk(e.m.Fields, nil, "")

	name := func(segs []string) *Statement {
		var lits []Code
		for _, s := range segs {
			lits = append(lits, Lit(s))
		}
		return Qual("strings", "Join").Call(Index().String().Values(lits...), Id("o").Dot("Separator"))
	}

	var reg []Code
	for _, s := range skipped {
		reg = append(reg, Comment("// "+s))
	}
	var lookups []Code
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

	var app []Code
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

// stdKind maps a field to stdlib flag's register method, defaulting sized
// ints through Int64/Uint64 with a range check at Apply.
func stdKind(f *Field) (reg string, ok bool) {
	k := f.Kind
	if k == KindPointer {
		k = f.Elem.Kind
	}
	switch k {
	case KindString:
		return "String", true
	case KindBool:
		return "Bool", true
	case KindFloat:
		return "Float64", true
	case KindInt, KindDuration:
		return "Int64", true
	case KindUint:
		return "Uint64", true
	}
	return "", false
}

func stdRegister(f *Field, name *Statement) Code {
	reg, ok := stdKind(f)
	if !ok {
		return Commentf("// %s: no std registration (unsupported flag kind)", f.Tag)
	}
	var def *Statement
	target := f
	if f.Kind == KindPointer {
		target = f.Elem
	}
	switch target.Kind {
	case KindString:
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
	}
	if target.Kind == KindDuration {
		// std registers durations as Int64 nanoseconds is hostile; use
		// fs.Duration which stdlib flag DOES have.
		reg = "Duration"
		def = durDefault(f)
	}
	return Id("fs").Dot(reg).Call(name, def, Lit(f.Desc))
}

func stdApply(f *Field, path string, name *Statement) []Code {
	_, ok := stdKind(f)
	if !ok {
		return []Code{Commentf("// %s: no std apply (unsupported flag kind)", f.Tag)}
	}
	target := f
	if f.Kind == KindPointer {
		target = f.Elem
	}
	// stdlib flag stores typed values behind flag.Getter
	get := Id("fs").Dot("Lookup").Call(Id("fn")).Dot("Value").Assert(Qual(stdFlag, "Getter")).Dot("Get").Call()
	var conv []Code
	assignTo := func(v *Statement) *Statement {
		if f.Kind == KindPointer {
			return cfgSel("cfg", f).Op("=").Op("&").Id("pv")
		}
		_ = v
		return cfgSel("cfg", f).Op("=").Id("pv")
	}
	t := target.Type.Underlying().String()
	switch target.Kind {
	case KindString:
		conv = []Code{Id("pv").Op(":=").Add(get).Assert(String())}
	case KindBool:
		conv = []Code{Id("pv").Op(":=").Add(get).Assert(Bool())}
	case KindFloat:
		if t == "float64" {
			conv = []Code{Id("pv").Op(":=").Add(get).Assert(Float64())}
		} else {
			conv = []Code{Id("pv").Op(":=").Id(t).Call(Add(get).Assert(Float64()))}
		}
	case KindDuration:
		conv = []Code{Id("pv").Op(":=").Add(get).Assert(Qual("time", "Duration"))}
	case KindInt:
		raw := Add(get).Assert(Int64())
		if target.Bits != 0 && target.Bits != 64 {
			conv = []Code{
				Id("raw").Op(":=").Add(raw),
				If(Id("raw").Op("<").Qual("math", fmt.Sprintf("MinInt%d", target.Bits)).Op("||").
					Id("raw").Op(">").Qual("math", fmt.Sprintf("MaxInt%d", target.Bits))).Block(
					Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
						Id("Path"): Lit(path), Id("Source"): Lit("-").Op("+").Id("fn"),
						Id("Err"): Qual("fmt", "Errorf").Call(Lit("%d overflows "+t), Id("raw")),
					}))),
				Id("pv").Op(":=").Id(t).Call(Id("raw")),
			}
		} else if t == "int64" {
			conv = []Code{Id("pv").Op(":=").Add(raw)}
		} else {
			conv = []Code{Id("pv").Op(":=").Id(t).Call(Add(raw))}
		}
	case KindUint:
		raw := Add(get).Assert(Uint64())
		if target.Bits != 0 && target.Bits != 64 {
			conv = []Code{
				Id("raw").Op(":=").Add(raw),
				If(Id("raw").Op(">").Qual("math", fmt.Sprintf("MaxUint%d", target.Bits))).Block(
					Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
						Id("Path"): Lit(path), Id("Source"): Lit("-").Op("+").Id("fn"),
						Id("Err"): Qual("fmt", "Errorf").Call(Lit("%d overflows "+t), Id("raw")),
					}))),
				Id("pv").Op(":=").Id(t).Call(Id("raw")),
			}
		} else if t == "uint64" {
			conv = []Code{Id("pv").Op(":=").Add(raw)}
		} else {
			conv = []Code{Id("pv").Op(":=").Id(t).Call(Add(raw))}
		}
	}
	body := append(conv,
		assignTo(nil),
		Id("set").Call(Lit(path), Qual(pkgCfg, "LayerCLI"), Lit("-").Op("+").Id("fn")),
	)
	return []Code{If(
		Id("fn").Op(":=").Add(name), Id("isSet").Index(Id("fn")),
	).Block(body...)}
}
