package main

import (
	"strconv"
	"strings"
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

	// collect flat flag list (scalars + string slices; collections skipped)
	type flagField struct {
		f    *Field
		segs []string
		path string
	}
	var flags []flagField
	var walk func(fields []*Field, segs []string, path string)
	walk = func(fields []*Field, segs []string, path string) {
		for _, f := range fields {
			s2 := append(append([]string{}, segs...), f.Tag)
			p2 := joinPath(path, f.Tag)
			switch f.Kind {
			case KindStruct:
				walk(f.Fields, s2, p2)
			case KindSliceStruct, KindMapStruct, KindMapScalar:
				// skipped for CLI (SPEC rule 6)
			default:
				if !f.FlagSkip {
					flags = append(flags, flagField{f, s2, p2})
				}
			}
		}
	}
	walk(e.m.Fields, nil, "")

	flagName := func(segs []string) *Statement {
		var lits []Code
		for _, s := range segs {
			lits = append(lits, Lit(s))
		}
		return Qual("strings", "Join").Call(Index().String().Values(lits...), Id("o").Dot("Separator"))
	}

	// Register
	var reg []Code
	var lookups []Code
	for _, ff := range flags {
		lookups = append(lookups, flagName(ff.segs))
	}
	reg = append(reg, For(List(Id("_"), Id("name")).Op(":=").Range().Index().String().Values(lookups...)).Block(
		If(Id("fs").Dot("Lookup").Call(Id("name")).Op("!=").Nil()).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("flag --%s already registered on this FlagSet"), Id("name"))),
		),
	))
	for _, ff := range flags {
		reg = append(reg, e.registerFlag(ff.f, flagName(ff.segs)))
	}
	reg = append(reg, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"RegisterPFlags").Params(
		Id("fs").Op("*").Qual(pfl, "FlagSet"), Id("o").Op("*").Qual(pkgPFlag, "Options"),
	).Error().Block(reg...)

	// Apply
	var app []Code
	for _, ff := range flags {
		app = append(app, e.applyFlag(ff.f, ff.path, flagName(ff.segs)))
	}
	app = append(app, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyPFlags").Params(
		Id("cfg").Op("*").Id(n), Id("fs").Op("*").Qual(pfl, "FlagSet"),
		Id("o").Op("*").Qual(pkgPFlag, "Options"), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(app...)
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
			return "Uint8", "GetUint8", Id("uint8").Call(Lit(int(v))), true
		case 16:
			return "Uint16", "GetUint16", Id("uint16").Call(Lit(int(v))), true
		case 32:
			return "Uint32", "GetUint32", Id("uint32").Call(Lit(int(v))), true
		case 64:
			return "Uint64", "GetUint64", Lit(uint64(v)), true
		}
		return "Uint", "GetUint", Lit(uint(v)), true
	case KindPointer:
		// *scalar is flag-settable, allocate-on-set (SPEC rule 5); no
		// pflag default (an unset optional stays nil).
		reg, get, _, ok := pflagTypeOps(&Field{Kind: f.Elem.Kind, Bits: f.Elem.Bits, Type: f.Elem.Type, Elem: f.Elem.Elem})
		if !ok || f.Elem.Kind == KindStruct {
			return "", "", nil, false
		}
		return reg, get, zeroDefault(f.Elem), true
	case KindSliceScalar:
		if f.Elem.Kind == KindString {
			var d *Statement
			if f.Default == "" {
				d = Nil()
			} else {
				var lits []Code
				for _, s := range strings.Split(f.Default, ",") {
					lits = append(lits, Lit(s))
				}
				d = Index().String().Values(lits...)
			}
			return "StringSlice", "GetStringSlice", d, true
		}
	}
	return "", "", nil, false
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
	}
	return Lit(0)
}

func (e *emitter) registerFlag(f *Field, name *Statement) Code {
	reg, _, def, ok := pflagTypeOps(f)
	if !ok {
		return Commentf("// %s: no pflag registration (unsupported flag kind)", f.Tag)
	}
	return Id("fs").Dot(reg).Call(name, def, Lit(f.Desc))
}

func (e *emitter) applyFlag(f *Field, path string, name *Statement) Code {
	_, get, _, ok := pflagTypeOps(f)
	if !ok {
		return Commentf("// %s: no pflag apply (unsupported flag kind)", f.Tag)
	}
	return If(
		Id("n").Op(":=").Add(name), Id("fs").Dot("Changed").Call(Id("n")),
	).Block(
		List(Id("v"), Err()).Op(":=").Id("fs").Dot(get).Call(Id("n")),
		If(Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
				Id("Path"): Lit(path), Id("Source"): Lit("--").Op("+").Id("n"), Id("Err"): Err(),
			})),
		),
		flagAssign(f),
		Id("set").Call(Lit(path), Qual(pkgCfg, "LayerCLI"), Lit("--").Op("+").Id("n")),
	)
}

func flagAssign(f *Field) Code {
	switch f.Kind {
	case KindPointer:
		return cfgSel("cfg", f).Op("=").Op("&").Id("v")
	}
	// exact-width Get* returns the exact type; platform int/uint need no cast
	return cfgSel("cfg", f).Op("=").Id("v")
}
