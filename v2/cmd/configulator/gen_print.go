package main

import (
	. "github.com/dave/jennifer/jen"
)

// emitPrintConfig emits a PrintConfig method on the config type that
// prints one "path = value" line per field in emission order, with
// secret:"true" values redacted. Report never holds values, so this is
// the only place redaction happens.
func (e *emitter) emitPrintConfig() {
	n := e.m.TypeName
	lines := e.printFields(e.m.Fields, "")
	body := make([]Code, 0, len(lines)+2)
	body = append(body, Var().Id("b").Qual("strings", "Builder"))
	body = append(body, lines...)
	body = append(body, Return(Id("b").Dot("String").Call()))
	e.f.Comment("PrintConfig renders every field as \"path = value\" lines, redacting")
	e.f.Comment("fields tagged secret:\"true\". The origin Report holds no values,")
	e.f.Comment("so this is the only place redaction happens.")
	e.f.Func().Params(Id("c").Op("*").Id(n)).Id("PrintConfig").Params().String().Block(body...)
}

func (e *emitter) printFields(fields []*Field, prefix string) []Code {
	var out []Code
	for _, f := range fields {
		path := joinPath(prefix, f.Tag)
		if f.Kind == KindStruct {
			out = append(out, e.printFields(f.Fields, path)...)
			continue
		}
		if f.Secret {
			out = append(out, Id("b").Dot("WriteString").Call(Lit(path+" = (redacted)\n")))
			continue
		}
		var val *Statement
		switch f.Kind {
		case KindPointer:
			out = append(out, If(cfgSel("c", f).Op("==").Nil()).Block(
				Id("b").Dot("WriteString").Call(Lit(path+" = <unset>\n")),
			).Else().Block(
				Id("b").Dot("WriteString").Call(Qual("fmt", "Sprintf").Call(
					Lit(path+" = %v\n"), Op("*").Add(cfgSel("c", f)))),
			))
			continue
		default:
			val = cfgSel("c", f)
		}
		out = append(out, Id("b").Dot("WriteString").Call(
			Qual("fmt", "Sprintf").Call(Lit(path+" = %v\n"), val)))
	}
	return out
}
