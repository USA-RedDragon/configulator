package main

import (
	"strings"

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
	return printFieldsAt(fields, prefix, func(f *Field) *Statement { return cfgSel("c", f) }, 0)
}

// printFieldsAt prints fields, selecting each one with sel. Inside an
// optional struct, depth numbers the variable holding the pointer.
func printFieldsAt(fields []*Field, prefix string, sel func(*Field) *Statement, depth int) []Code {
	var out []Code
	for _, f := range fields {
		path := joinPath(prefix, f.Tag)
		if f.Kind == KindStruct {
			out = append(out, printFieldsAt(f.Fields, path, sel, depth)...)
			continue
		}
		if f.Secret || hasSecret(f) {
			out = append(out, Id("b").Dot("WriteString").Call(Lit(path+" = (redacted)\n")))
			continue
		}
		var val *Statement
		switch f.Kind {
		case KindPointer:
			unset := Id("b").Dot("WriteString").Call(Lit(path + " = <unset>\n"))
			if f.Elem.Kind == KindStruct {
				p := depthName("p", depth)
				inner := printFieldsAt(f.Elem.Fields, path, func(sf *Field) *Statement { return relSel(Id(p), sf) }, depth+1)
				out = append(out, If(Id(p).Op(":=").Add(sel(f)), Id(p).Op("==").Nil()).Block(unset).Else().Block(inner...))
				continue
			}
			out = append(out, If(sel(f).Op("==").Nil()).Block(unset).Else().Block(
				Id("b").Dot("WriteString").Call(Qual("fmt", "Sprintf").Call(
					Lit(path+" = %v\n"), Op("*").Add(sel(f)))),
			))
			continue
		default:
			val = sel(f)
		}
		out = append(out, Id("b").Dot("WriteString").Call(
			Qual("fmt", "Sprintf").Call(Lit(path+" = %v\n"), val)))
	}
	return out
}

// relSel selects f from recv, the optional struct f belongs to. Fields of
// an optional struct carry Go paths starting with the pointer field's own
// name, which recv replaces.
func relSel(recv *Statement, f *Field) *Statement {
	s := recv.Clone()
	parts := strings.Split(f.GoName, ".")
	for _, p := range parts[1:] {
		s = s.Dot(p)
	}
	return s
}

// hasSecret reports whether a collection's elements hold a secret field
// anywhere, in which case the whole collection is redacted.
func hasSecret(f *Field) bool {
	switch f.Kind {
	case KindSliceStruct, KindMapStruct:
		return anySecret(f.Elem.Fields)
	default:
		return false
	}
}

func anySecret(fields []*Field) bool {
	for _, f := range fields {
		if f.Secret {
			return true
		}
		switch f.Kind {
		case KindStruct:
			if anySecret(f.Fields) {
				return true
			}
		case KindPointer, KindSliceStruct, KindMapStruct:
			if f.Elem.Kind == KindStruct && anySecret(f.Elem.Fields) {
				return true
			}
		default:
		}
	}
	return false
}
