package main

import (
	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitApplyEnv() {
	n := e.m.TypeName
	body := append(e.envFields(e.m.Fields), Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyEnv").Params(
		Id("cfg").Op("*").Id(n), Id("ec").Qual(pkgCfg, "EnvContext"), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(body...)
}

func envNameCall(segments []string) *Statement {
	args := make([]Code, 0, 2+len(segments))
	args = append(args, Id("ec").Dot("Opts").Dot("Prefix"), Id("ec").Dot("Opts").Dot("Separator"))
	for _, s := range segments {
		args = append(args, Lit(s))
	}
	return Qual(pkgImpl, "EnvName").Call(args...)
}

func (e *emitter) envFields(fields []*Field) []Code {
	ls := leaves(fields, (*Field).envSeg, func(f *Field) bool { return f.EnvSkip })
	out := make([]Code, 0, len(ls))
	for _, l := range ls {
		inner, val := e.envValue(l.f, l.path)
		inner = append(inner, e.chainAssign(l, Id("ec").Dot("ArraySeparator"), func(t *Statement) []Code {
			return []Code{t.Op("=").Add(val)}
		})...)
		inner = append(inner, Id("set").Call(Lit(l.path), Qual(pkgCfg, "LayerEnv"), Id("n")))
		out = append(out, If(
			Id("n").Op(":=").Add(envNameCall(l.segs)), True(),
		).Block(
			If(List(Id("v"), Id("ok")).Op(":=").Id("ec").Dot("Getenv").Call(Id("n")), Id("ok")).Block(inner...),
		))
	}
	return out
}

// envValue returns statements that parse the env value v, returning a
// ParseError on failure, and the expression to assign to the field.
func (e *emitter) envValue(f *Field, path string) ([]Code, *Statement) {
	errVal := func() Code {
		if f.Secret {
			return Lit("(redacted)")
		}
		return Id("v")
	}
	parseErr := func() Code {
		return Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
			Id("Path"): Lit(path), Id("Source"): Id("n"), Id("Value"): errVal(), Id("Err"): Err(),
		}))
	}
	switch f.Kind {
	case KindString:
		return nil, convNamed(f.Type, Id("v"))
	case KindBool:
		return []Code{
			List(Id("p"), Err()).Op(":=").Qual("strconv", "ParseBool").Call(Id("v")),
			If(Err().Op("!=").Nil()).Block(parseErr()),
		}, convNamed(f.Type, Id("p"))
	case KindInt, KindUint, KindFloat:
		return append(parseNumeric(f, "p"), If(Err().Op("!=").Nil()).Block(parseErr())),
			convNamed(f.Type, numConv(f, "p"))
	case KindDuration:
		return []Code{
			List(Id("d"), Err()).Op(":=").Qual("time", "ParseDuration").Call(Id("v")),
			If(Err().Op("!=").Nil()).Block(parseErr()),
		}, Id("d")
	case KindStdSlot:
		return []Code{
			Var().Id("slot").Add(slotCode(f)),
			If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(Id("v"))), Err().Op("!=").Nil()).Block(parseErr()),
			List(Id("sv"), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
		}, convNamed(f.Type, Id("sv"))
	case KindSliceScalar:
		return parseList(f, Qual(pkgImpl, "SplitList").Call(Id("v"), Id("ec").Dot("ArraySeparator")), Lit(path), Id("n"), Id("v")),
			Id("lst")
	case KindPointer:
		switch f.Elem.Kind {
		case KindString:
			return []Code{Id("pv").Op(":=").Add(convNamed(f.Elem.Type, Id("v")))}, Op("&").Id("pv")
		case KindBool, KindInt, KindUint, KindFloat:
			return append(parseNumericOrBool(f.Elem, "p"),
				If(Err().Op("!=").Nil()).Block(parseErr()),
				Id("pv").Op(":=").Add(convNamed(f.Elem.Type, numConv(f.Elem, "p")))), Op("&").Id("pv")
		default:
			elem := *f.Elem
			elem.Secret = f.Secret
			prep, val := e.envValue(&elem, path)
			return append(prep, Id("pv").Op(":=").Add(val)), Op("&").Id("pv")
		}
	default:
	}
	panic("envValue: unhandled kind for " + path)
}

func parseNumeric(f *Field, dst string) []Code {
	bits := f.Bits
	if bits == 0 {
		bits = 64
	}
	switch f.Kind {
	case KindInt:
		return []Code{List(Id(dst), Err()).Op(":=").Qual("strconv", "ParseInt").Call(Id("v"), Lit(10), Lit(bits))}
	case KindUint:
		return []Code{List(Id(dst), Err()).Op(":=").Qual("strconv", "ParseUint").Call(Id("v"), Lit(10), Lit(bits))}
	case KindFloat:
		return []Code{List(Id(dst), Err()).Op(":=").Qual("strconv", "ParseFloat").Call(Id("v"), Lit(bits))}
	default:
		panic("parseNumeric")
	}
}

func parseNumericOrBool(f *Field, dst string) []Code {
	if f.Kind == KindBool {
		return []Code{List(Id(dst), Err()).Op(":=").Qual("strconv", "ParseBool").Call(Id("v"))}
	}
	return parseNumeric(f, dst)
}

// numConv converts the strconv result to the field's exact type.
func numConv(f *Field, v string) *Statement {
	if f.Kind == KindBool {
		return Id(v)
	}
	t := f.Type.Underlying().String()
	switch t {
	case "int64", "uint64", "float64":
		return Id(v)
	}
	return Id(t).Call(Id(v))
}
