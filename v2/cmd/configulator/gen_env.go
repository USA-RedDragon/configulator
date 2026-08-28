package main

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitApplyEnv() {
	n := e.m.TypeName
	var body []Code
	e.envFields(&body, e.m.Fields, nil, "")
	body = append(body, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyEnv").Params(
		Id("cfg").Op("*").Id(n), Id("ec").Qual(pkgCfg, "EnvContext"), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(body...)
}

func envNameCall(segments []string) *Statement {
	args := []Code{Id("ec").Dot("Opts").Dot("Prefix"), Id("ec").Dot("Opts").Dot("Separator")}
	for _, s := range segments {
		args = append(args, Lit(s))
	}
	return Qual(pkgCfg, "EnvName").Call(args...)
}

func (e *emitter) envFields(body *[]Code, fields []*Field, segPrefix []string, pathPrefix string) {
	for _, f := range fields {
		segs := append(append([]string{}, segPrefix...), f.Tag)
		path := joinPath(pathPrefix, f.Tag)
		if f.EnvSkip {
			e.f.Comment("// " + path + ": env:\"-\"")
			continue
		}
		switch f.Kind {
		case KindStruct:
			e.envFields(body, f.Fields, segs, path)
			continue
		case KindPointer:
			if f.Elem.Kind == KindStruct {
				e.envPtrStruct(body, f, segs, path)
				continue
			}
		case KindSliceStruct, KindMapStruct, KindMapScalar:
			*body = append(*body, Commentf("// %s: %s — skipped for env (SPEC rule 6, generator warning)", path, kindWord(f.Kind)))
			continue
		}
		inner := e.envAssign(f, path)
		*body = append(*body, If(
			Id("n").Op(":=").Add(envNameCall(segs)), True(),
		).Block(
			If(List(Id("v"), Id("ok")).Op(":=").Id("ec").Dot("Getenv").Call(Id("n")), Id("ok")).Block(inner...),
		))
	}
}

// envPtrStruct emits env handling for leaves inside a *Struct: the struct
// is allocated with its element defaults on first write, and CLONED on
// subsequent writes — never mutated through the shared pointer (rule 11
// staging contract).
func (e *emitter) envPtrStruct(body *[]Code, f *Field, segs []string, path string) {
	elemType := fieldGoType(f.Elem.Type)
	for _, lf := range f.Elem.Fields {
		lsegs := append(append([]string{}, segs...), lf.Tag)
		lpath := path + "." + lf.Tag
		if lf.Kind == KindStruct || lf.Kind == KindPointer {
			panic("envPtrStruct: nested composite inside *Struct not supported yet: " + lpath)
		}

		var alloc []Code
		alloc = append(alloc, Var().Id("e").Add(elemType.Clone()))
		var defaults []Code
		for _, df := range f.Elem.Fields {
			if df.Default == "" {
				continue
			}
			switch df.Kind {
			case KindString:
				defaults = append(defaults, Id("e").Dot(goName(df)).Op("=").Lit(df.Default))
			case KindUint, KindInt:
				v := 0
				fmt.Sscanf(df.Default, "%d", &v)
				defaults = append(defaults, Id("e").Dot(goName(df)).Op("=").Add(castLit(df, v)))
			}
			defaults = append(defaults, Id("set").Call(
				Lit(path+"."+df.Tag), Qual(pkgCfg, "LayerDefault"), Lit("element default")))
		}
		alloc = append(alloc, If(cfgSel("cfg", f).Op("!=").Nil()).Block(
			Id("e").Op("=").Op("*").Add(cfgSel("cfg", f)),
		).Else().Block(defaults...))

		switch lf.Kind {
		case KindString, KindBool, KindInt, KindUint, KindFloat:
		default:
			panic("envPtrStruct: leaf kind not supported at " + lpath)
		}

		// order: parse first (so a bad value errors before allocation),
		// then allocate/clone, then assign + record
		var inner []Code
		if lf.Kind == KindString {
			inner = append(inner, alloc...)
			inner = append(inner, Id("e").Dot(goName(lf)).Op("=").Id("v"))
		} else {
			inner = append(inner, parseNumericOrBool(lf, "p")...)
			inner = append(inner, If(Err().Op("!=").Nil()).Block(
				Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
					Id("Path"): Lit(lpath), Id("Source"): Id("n"), Id("Value"): Id("v"), Id("Err"): Err(),
				}))))
			inner = append(inner, alloc...)
			inner = append(inner, Id("e").Dot(goName(lf)).Op("=").Add(numConv(lf, "p")))
		}
		inner = append(inner,
			cfgSel("cfg", f).Op("=").Op("&").Id("e"),
			Id("set").Call(Lit(lpath), Qual(pkgCfg, "LayerEnv"), Id("n")),
		)

		*body = append(*body, If(
			Id("n").Op(":=").Add(envNameCall(lsegs)), True(),
		).Block(
			If(List(Id("v"), Id("ok")).Op(":=").Id("ec").Dot("Getenv").Call(Id("n")), Id("ok")).Block(inner...),
		))
	}
}

func kindWord(k Kind) string {
	switch k {
	case KindSliceStruct:
		return "[]struct"
	case KindMapStruct, KindMapScalar:
		return "map"
	}
	return "?"
}

// envAssign parses string v into the field and records the origin.
func (e *emitter) envAssign(f *Field, path string) []Code {
	rec := Id("set").Call(Lit(path), Qual(pkgCfg, "LayerEnv"), Id("n"))
	errVal := func() Code {
		// secret: fields never leak their raw value through an error
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
		return []Code{cfgSel("cfg", f).Op("=").Id("v"), rec}
	case KindBool:
		return []Code{
			List(Id("p"), Err()).Op(":=").Qual("strconv", "ParseBool").Call(Id("v")),
			If(Err().Op("!=").Nil()).Block(parseErr()),
			cfgSel("cfg", f).Op("=").Id("p"), rec,
		}
	case KindInt, KindUint, KindFloat:
		return append(parseNumeric(f, "p"),
			If(Err().Op("!=").Nil()).Block(parseErr()),
			cfgSel("cfg", f).Op("=").Add(numConv(f, "p")), rec)
	case KindDuration:
		return []Code{
			List(Id("d"), Err()).Op(":=").Qual("time", "ParseDuration").Call(Id("v")),
			If(Err().Op("!=").Nil()).Block(parseErr()),
			cfgSel("cfg", f).Op("=").Add(castStdVar(f, "d")), rec,
		}
	case KindStdSlot, KindTextLeaf:
		// route through the slot/leaf's UnmarshalText
		return []Code{
			Var().Id("slot").Qual(pkgCfg, f.SlotType),
			If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(Id("v"))), Err().Op("!=").Nil()).Block(parseErr()),
			List(Id("sv"), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
			cfgSel("cfg", f).Op("=").Add(castStdVar(f, "sv")), rec,
		}
	case KindSliceScalar:
		return []Code{
			cfgSel("cfg", f).Op("=").Qual(pkgCfg, "SplitList").Call(Id("v"), Id("ec").Dot("ArraySeparator")),
			rec,
		}
	case KindPointer:
		switch f.Elem.Kind {
		case KindString:
			return []Code{cfgSel("cfg", f).Op("=").Op("&").Id("v").Comment("// allocate-on-set (SPEC rule 5)"), rec}
		case KindBool, KindInt, KindUint, KindFloat:
			conv := numConv(f.Elem, "p")
			return append(parseNumericOrBool(f.Elem, "p"),
				If(Err().Op("!=").Nil()).Block(parseErr()),
				Id("pv").Op(":=").Add(conv),
				cfgSel("cfg", f).Op("=").Op("&").Id("pv").Comment("// allocate-on-set (SPEC rule 5)"), rec)
		}
	}
	panic("envAssign: unhandled kind for " + path)
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
		return []Code{List(Id(dst), Err()).Op(":=").Qual("strconv", "ParseFloat").Call(Id("v"), Lit(64))}
	}
	panic("parseNumeric")
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

func castStdVar(f *Field, v string) *Statement {
	return Id(v)
}

// ---- pflag hooks ----
