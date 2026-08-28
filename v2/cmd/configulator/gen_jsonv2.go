package main

import (
	"fmt"
	"go/types"

	. "github.com/dave/jennifer/jen"
)

const (
	pkgJSONText = "encoding/json/jsontext"
	pkgJSONv2   = "encoding/json/v2"
)

// emitFastPaths emits an UnmarshalJSONFrom method for every shadow struct.
// Under json/v2 the method's presence bypasses struct tags entirely, so
// the generated decoder owns three semantics itself (G7): case-sensitive
// key matching, unknown-key rejection, and sentinel-slot enforcement
// (slots decode from string tokens only). goccy and go-toml ignore the
// method and use the struct tags. Emitting it for every shadow — element
// and nested shadows included — is what makes the whole file-layer decode
// genuinely reflection-free; a top-level-only method would re-enter
// reflection for each sub-struct.
func (e *emitter) emitFastPaths() {
	for _, sh := range e.shadowOrder {
		e.emitFastPath(sh, e.shadowFields[sh])
	}
}

func (e *emitter) emitFastPath(shadow string, fields []*Field) {
	var cases []Code
	for _, f := range fields {
		cases = append(cases, Case(Lit(f.Tag)).Block(e.fastField(f)...))
	}
	cases = append(cases, Default().Block(
		Return(Qual("fmt", "Errorf").Call(Lit("unknown key %q"), Id("tok").Dot("String").Call())),
	))

	e.f.Func().Params(Id("s").Op("*").Id(shadow)).Id("UnmarshalJSONFrom").Params(
		Id("dec").Op("*").Qual(pkgJSONText, "Decoder"),
	).Error().Block(
		List(Id("tok"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
		If(Id("tok").Dot("Kind").Call().Op("!=").LitRune('{')).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("expected object, got %v"), Id("tok").Dot("Kind").Call())),
		),
		For().Block(
			List(Id("tok"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())),
			If(Id("tok").Dot("Kind").Call().Op("==").LitRune('}')).Block(Return(Nil())),
			Switch(Id("tok").Dot("String").Call()).Block(cases...),
		),
	)
	e.f.Var().Id("_").Qual(pkgJSONv2, "UnmarshalerFrom").Op("=").Parens(Op("*").Id(shadow)).Call(Nil())
}

// fastField emits one case body: decode the next value into s.<Field>.
// null means absent everywhere (SPEC rule 5): skip without assigning.
func (e *emitter) fastField(f *Field) []Code {
	sel := func() *Statement { return Id("s").Dot(goName(f)) }
	switch f.Kind {
	case KindString:
		return e.fastScalar(f, sel, '"',
			func() []Code {
				return []Code{
					Id("str").Op(":=").Id("v").Dot("String").Call(),
					sel().Op("=").Op("&").Id("str"),
				}
			}, "a string")
	case KindBool:
		return []Code{
			List(Id("v"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())),
			Switch(Id("v").Dot("Kind").Call()).Block(
				Case(LitRune('n')).Block(),
				Case(LitRune('t'), LitRune('f')).Block(
					Id("b").Op(":=").Id("v").Dot("Bool").Call(),
					sel().Op("=").Op("&").Id("b"),
				),
				Default().Block(Return(Qual("fmt", "Errorf").Call(
					Lit(f.Tag+": expected a bool, got %v"), Id("v").Dot("Kind").Call()))),
			),
		}
	case KindInt, KindUint, KindFloat:
		return e.fastNumeric(f, sel)
	case KindDuration, KindStdSlot:
		hint := slotHint(f.SlotType)
		return []Code{
			List(Id("v"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())),
			Switch(Id("v").Dot("Kind").Call()).Block(
				Case(LitRune('n')).Block(),
				Case(LitRune('"')).Block(
					Var().Id("slot").Qual(pkgCfg, f.SlotType),
					If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(
						Index().Byte().Parens(Id("v").Dot("String").Call())), Err().Op("!=").Nil()).Block(Return(Err())),
					sel().Op("=").Op("&").Id("slot"),
				),
				Default().Block(Return(Qual("fmt", "Errorf").Call(
					Lit(f.Tag+": expected a text scalar (e.g. "+hint+"), got %v"), Id("v").Dot("Kind").Call()))),
			),
		}
	case KindStruct:
		sub := e.shadowFor(f.Type)
		return []Code{
			If(Id("dec").Dot("PeekKind").Call().Op("==").LitRune('n')).Block(
				If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
			).Else().Block(
				Var().Id("sub").Id(sub),
				If(Err().Op(":=").Id("sub").Dot("UnmarshalJSONFrom").Call(Id("dec")), Err().Op("!=").Nil()).Block(Return(Err())),
				sel().Op("=").Op("&").Id("sub"),
			),
		}
	case KindPointer:
		inner := *f.Elem
		inner.GoName = f.GoName
		inner.Tag = f.Tag
		return e.fastField(&inner)
	case KindSliceScalar:
		return e.fastSliceScalar(f, sel)
	case KindSliceStruct:
		return e.fastSliceStruct(f, sel)
	case KindMapScalar:
		return e.fastMapScalar(f, sel)
	case KindMapStruct:
		return e.fastMapStruct(f, sel)
	case KindTextLeaf:
		return e.fastScalar(f, sel, '"', func() []Code {
			return []Code{
				Var().Id("leaf").Add(fieldGoType(f.Type)),
				If(Err().Op(":=").Id("leaf").Dot("UnmarshalText").Call(
					Index().Byte().Parens(Id("v").Dot("String").Call())), Err().Op("!=").Nil()).Block(Return(Err())),
				sel().Op("=").Op("&").Id("leaf"),
			}
		}, "a text scalar")
	}
	panic("fastField: unhandled kind for " + f.Tag)
}

func (e *emitter) fastScalar(f *Field, sel func() *Statement, kind rune, assign func() []Code, want string) []Code {
	body := []Code{
		List(Id("v"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
	}
	inner := assign()
	body = append(body, Switch(Id("v").Dot("Kind").Call()).Block(
		Case(LitRune('n')).Block(),
		Case(LitRune(kind)).Block(inner...),
		Default().Block(Return(Qual("fmt", "Errorf").Call(
			Lit(f.Tag+": expected "+want+", got %v"), Id("v").Dot("Kind").Call()))),
	))
	return body
}

func (e *emitter) fastNumeric(f *Field, sel func() *Statement) []Code {
	t := f.Type.Underlying().String()
	var read []Code
	errCheck := If(Err().Op("!=").Nil()).Block(Return(Err()))
	switch f.Kind {
	case KindInt:
		read = []Code{List(Id("num"), Err()).Op(":=").Id("v").Dot("Int").Call(), errCheck}
		if f.Bits != 0 && f.Bits != 64 {
			read = append(read, If(Id("num").Op("<").Qual("math", fmt.Sprintf("MinInt%d", f.Bits)).
				Op("||").Id("num").Op(">").Qual("math", fmt.Sprintf("MaxInt%d", f.Bits))).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": %d overflows "+t), Id("num")))))
		}
	case KindUint:
		read = []Code{List(Id("num"), Err()).Op(":=").Id("v").Dot("Uint").Call(), errCheck}
		if f.Bits != 0 && f.Bits != 64 {
			read = append(read, If(Id("num").Op(">").Qual("math", fmt.Sprintf("MaxUint%d", f.Bits))).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": %d overflows "+t), Id("num")))))
		}
	case KindFloat:
		read = []Code{List(Id("num"), Err()).Op(":=").Id("v").Dot("Float").Call(), errCheck}
	}
	conv := Id(t).Call(Id("num"))
	if t == "int64" && f.Kind == KindInt || t == "uint64" && f.Kind == KindUint || t == "float64" && f.Kind == KindFloat {
		conv = Id("num")
	}
	read = append(read, Id("val").Op(":=").Add(conv), sel().Op("=").Op("&").Id("val"))
	return []Code{
		List(Id("v"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
		Switch(Id("v").Dot("Kind").Call()).Block(
			Case(LitRune('n')).Block(),
			Case(LitRune('0')).Block(read...),
			Default().Block(Return(Qual("fmt", "Errorf").Call(
				Lit(f.Tag+": expected a number, got %v"), Id("v").Dot("Kind").Call()))),
		),
	}
}

// scalarElemReader emits statements reading ONE scalar token into a new
// variable named dst, for a sequence/map element of the given kind.
func scalarElemReader(f *Field, dst string) []Code {
	read := []Code{
		List(Id("v"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
	}
	switch f.Kind {
	case KindString:
		read = append(read,
			If(Id("v").Dot("Kind").Call().Op("!=").LitRune('"')).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": expected a string element, got %v"), Id("v").Dot("Kind").Call()))),
			Id(dst).Op(":=").Id("v").Dot("String").Call(),
		)
	case KindBool:
		read = append(read,
			If(Id("v").Dot("Kind").Call().Op("!=").LitRune('t').Op("&&").Id("v").Dot("Kind").Call().Op("!=").LitRune('f')).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": expected a bool element, got %v"), Id("v").Dot("Kind").Call()))),
			Id(dst).Op(":=").Id("v").Dot("Bool").Call(),
		)
	case KindInt, KindUint, KindFloat:
		t := f.Type.Underlying().String()
		read = append(read,
			If(Id("v").Dot("Kind").Call().Op("!=").LitRune('0')).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": expected a number element, got %v"), Id("v").Dot("Kind").Call()))))
		var method string
		switch f.Kind {
		case KindInt:
			method = "Int"
		case KindUint:
			method = "Uint"
		case KindFloat:
			method = "Float"
		}
		read = append(read,
			List(Id("raw"), Err()).Op(":=").Id("v").Dot(method).Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())))
		if (f.Kind == KindInt && t == "int64") || (f.Kind == KindUint && t == "uint64") || (f.Kind == KindFloat && t == "float64") {
			read = append(read, Id(dst).Op(":=").Id("raw"))
		} else {
			read = append(read, Id(dst).Op(":=").Id(t).Call(Id("raw")))
		}
	default:
		panic("scalarElemReader: unsupported element kind for " + f.Tag)
	}
	return read
}

// nullOrOpen wraps body in the null-check + open-token boilerplate shared
// by every sequence and map field: null means absent; otherwise the next
// token must be open ('[' or '{').
func nullOrOpen(f *Field, open rune, want string, body []Code) []Code {
	return []Code{
		If(Id("dec").Dot("PeekKind").Call().Op("==").LitRune('n')).Block(
			If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		).Else().Block(append([]Code{
			List(Id("tok"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())),
			If(Id("tok").Dot("Kind").Call().Op("!=").LitRune(open)).Block(
				Return(Qual("fmt", "Errorf").Call(Lit(f.Tag+": expected "+want+", got %v"), Id("tok").Dot("Kind").Call()))),
		}, body...)...),
	}
}

func (e *emitter) fastSliceScalar(f *Field, sel func() *Statement) []Code {
	elemT := fieldGoType(f.Type.Underlying().(*types.Slice).Elem())
	body := []Code{
		Id("out").Op(":=").Index().Add(elemT.Clone()).Values(),
		For(Id("dec").Dot("PeekKind").Call().Op("!=").LitRune(']')).Block(
			append(scalarElemReader(f.Elem, "el"), Id("out").Op("=").Append(Id("out"), Id("el")))...,
		),
		If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		sel().Op("=").Op("&").Id("out"),
	}
	return nullOrOpen(f, '[', "an array", body)
}

func (e *emitter) fastSliceStruct(f *Field, sel func() *Statement) []Code {
	sub := e.shadowFor(f.Elem.Type)
	body := []Code{
		Id("out").Op(":=").Index().Id(sub).Values(),
		For(Id("dec").Dot("PeekKind").Call().Op("!=").LitRune(']')).Block(
			Var().Id("el").Id(sub),
			If(Err().Op(":=").Id("el").Dot("UnmarshalJSONFrom").Call(Id("dec")), Err().Op("!=").Nil()).Block(Return(Err())),
			Id("out").Op("=").Append(Id("out"), Id("el")),
		),
		If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		sel().Op("=").Op("&").Id("out"),
	}
	return nullOrOpen(f, '[', "an array", body)
}

func (e *emitter) fastMapScalar(f *Field, sel func() *Statement) []Code {
	elemT := fieldGoType(f.Type.Underlying().(*types.Map).Elem())
	body := []Code{
		Id("out").Op(":=").Map(String()).Add(elemT.Clone()).Values(),
		For(Id("dec").Dot("PeekKind").Call().Op("!=").LitRune('}')).Block(
			append(append([]Code{
				List(Id("kt"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
				If(Err().Op("!=").Nil()).Block(Return(Err())),
				Id("mk").Op(":=").Id("kt").Dot("String").Call(),
			}, scalarElemReader(f.Elem, "el")...),
				Id("out").Index(Id("mk")).Op("=").Id("el"))...,
		),
		If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		sel().Op("=").Op("&").Id("out"),
	}
	return nullOrOpen(f, '{', "an object", body)
}

func (e *emitter) fastMapStruct(f *Field, sel func() *Statement) []Code {
	sub := e.shadowFor(f.Elem.Type)
	body := []Code{
		Id("out").Op(":=").Map(String()).Id(sub).Values(),
		For(Id("dec").Dot("PeekKind").Call().Op("!=").LitRune('}')).Block(
			List(Id("kt"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
			If(Err().Op("!=").Nil()).Block(Return(Err())),
			Id("mk").Op(":=").Id("kt").Dot("String").Call(),
			Var().Id("el").Id(sub),
			If(Err().Op(":=").Id("el").Dot("UnmarshalJSONFrom").Call(Id("dec")), Err().Op("!=").Nil()).Block(Return(Err())),
			Id("out").Index(Id("mk")).Op("=").Id("el"),
		),
		If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		sel().Op("=").Op("&").Id("out"),
	}
	return nullOrOpen(f, '{', "an object", body)
}
