package main

import (
	"strconv"

	. "github.com/dave/jennifer/jen"
)

const (
	pkgJSONText = "encoding/json/jsontext"
	pkgJSONv2   = "encoding/json/v2"
)

// jsonKind renders the jsontext.Kind constant for the token kind r.
func jsonKind(r rune) *Statement {
	names := map[rune]string{
		'n': "KindNull", 'f': "KindFalse", 't': "KindTrue", '"': "KindString", '0': "KindNumber",
		'{': "KindBeginObject", '}': "KindEndObject", '[': "KindBeginArray", ']': "KindEndArray",
	}
	return Qual(pkgJSONText, names[r])
}

// emitFastPaths gives the root shadow an UnmarshalJSONFrom method and every
// other shadow a decodeJSON method it calls, so json/v2 never falls back
// to reflection. json/v2 skips struct tags when the method exists, so the
// generated code does its own case-sensitive key matching and string-only
// sentinel slots. It rejects unknown keys only when the decoder sets
// json.RejectUnknownMembers, as StrictJSON does, and skips them otherwise.
// Errors about a value are ParseErrors and errors about a key are
// UnknownKeyErrors, both carrying the full dotted path.
func (e *emitter) emitFastPaths() {
	root := shadowName(e.m.TypeName)
	for _, sh := range e.shadowOrder {
		e.emitFastPath(sh, e.shadowFields[sh], sh == root)
	}
	e.decl().Comment(e.jsonErrorName() + " returns a ParseError for the JSON token v at path.")
	e.f.Func().Id(e.jsonErrorName()).Params(
		Id("path").String(), Id("v").Qual(pkgJSONText, "Token"), Err().Error(),
	).Error().Block(
		Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
			Id("Path"): Id("path"), Id("Value"): Id("v").Dot("String").Call(), Id("Err"): Err(),
		})),
	)
}

// jsonErrorName returns the name of the helper that builds a ParseError
// for a JSON token.
func (e *emitter) jsonErrorName() string {
	return lowerFirst(e.m.TypeName) + "JSONError"
}

// jsonError returns a statement that returns the ParseError for the token
// in the variable tok, caused by err. path renders the dotted path.
func (e *emitter) jsonError(path func() *Statement, tok string, err Code) Code {
	return Return(Id(e.jsonErrorName()).Call(path(), Id(tok), err))
}

// kindError returns a statement that returns a ParseError saying the token
// in tok is not want. path renders the dotted path.
func (e *emitter) kindError(path func() *Statement, tok, want string) Code {
	return e.jsonError(path, tok, Qual("fmt", "Errorf").Call(Lit("expected "+want+", got %v"), Id(tok).Dot("Kind").Call()))
}

// fieldPath returns the path of the member tag of the object being
// decoded: a literal at the root, or the object's path variable plus the
// tag in a nested object.
func fieldPath(root bool, tag string) pathExpr {
	if root {
		return pathExpr{suffix: tag}
	}
	return pathExpr{v: "path", suffix: "." + tag}
}

func (e *emitter) emitFastPath(shadow string, fields []*Field, root bool) {
	cases := make([]Code, 0, len(fields)+1)
	for _, f := range fields {
		p := fieldPath(root, f.Tag)
		body := e.fastField(f, p.code)
		if f.Secret {
			body = redactFastErrors(p.code, body)
		}
		cases = append(cases, Case(Lit(f.Tag)).Block(body...))
	}
	unknown := Id(e.quoteKeyName()).Call(Id("key"))
	if !root {
		unknown = Id("path").Op("+").Lit(".").Op("+").Add(unknown)
	}
	cases = append(cases, Default().Block(
		If(
			List(Id("reject"), Id("_")).Op(":=").Qual(pkgJSONv2, "GetOption").Call(Id("dec").Dot("Options").Call(), Qual(pkgJSONv2, "RejectUnknownMembers")),
			Id("reject"),
		).Block(
			Return(Op("&").Qual(pkgCfg, "UnknownKeyError").Values(Dict{Id("Path"): unknown})),
		),
		If(Err().Op(":=").Id("dec").Dot("SkipValue").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
	))
	loop := For().Block(
		List(Id("tok"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
		If(Id("tok").Dot("Kind").Call().Op("==").Add(jsonKind('}'))).Block(Return(Nil())),
		Switch(Id("key").Op(":=").Id("tok").Dot("String").Call(), Id("key")).Block(cases...),
	)

	if !root {
		e.decl().Comment("decodeJSON decodes the members of an object whose opening brace has")
		e.f.Comment("been read. path is the object's dotted path.")
		e.f.Func().Params(Id("s").Op("*").Id(shadow)).Id("decodeJSON").Params(
			Id("dec").Op("*").Qual(pkgJSONText, "Decoder"), Id("path").String(),
		).Error().Block(loop)
		return
	}
	e.decl().Func().Params(Id("s").Op("*").Id(shadow)).Id("UnmarshalJSONFrom").Params(
		Id("dec").Op("*").Qual(pkgJSONText, "Decoder"),
	).Error().Block(
		List(Id("tok"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
		If(Id("tok").Dot("Kind").Call().Op("!=").Add(jsonKind('{'))).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("expected an object, got %v"), Id("tok").Dot("Kind").Call())),
		),
		loop,
	)
	e.decl().Var().Id("_").Qual(pkgJSONv2, "UnmarshalerFrom").Op("=").Parens(Op("*").Id(shadow)).Call(Nil())
}

// redactFastErrors runs body in a closure and replaces any error it returns
// with one that leaves out the value, since decode errors, even syntax
// errors, can quote part of a secret.
func redactFastErrors(p func() *Statement, body []Code) []Code {
	run := Func().Params().Error().Block(append(body, Return(Nil()))...).Call()
	return []Code{If(Err().Op(":=").Add(run), Err().Op("!=").Nil()).Block(
		Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
			Id("Path"): p(), Id("Value"): Lit("(redacted)"), Id("Err"): Qual("errors", "New").Call(Lit("invalid value")),
		})),
	)}
}

// readToken reads the next token into v, returning on an error.
func readToken(v string) []Code {
	return []Code{
		List(Id(v), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(),
		If(Err().Op("!=").Nil()).Block(Return(Err())),
	}
}

// fastField emits one case body that decodes the next value, at the path p
// renders, into s.<Field>. null means absent, so it's skipped without
// assigning.
func (e *emitter) fastField(f *Field, p func() *Statement) []Code {
	sel := func() *Statement { return Id("s").Dot(goName(f)) }
	switch f.Kind {
	case KindString:
		return e.fastScalar(p, []rune{'"'}, "a string", []Code{
			Id("str").Op(":=").Id("v").Dot("String").Call(),
			sel().Op("=").Op("&").Id("str"),
		})
	case KindBool:
		return e.fastScalar(p, []rune{'t', 'f'}, "a bool", []Code{
			Id("b").Op(":=").Id("v").Dot("Bool").Call(),
			sel().Op("=").Op("&").Id("b"),
		})
	case KindInt, KindUint, KindFloat:
		read := append(e.readNumber(f, p, "num"), sel().Op("=").Op("&").Id("num"))
		return e.fastScalar(p, []rune{'0'}, "a number", read)
	case KindDuration, KindStdSlot:
		kinds := []rune{'"'}
		if isComplexSlot(f) {
			kinds = append(kinds, '0')
		}
		return e.fastScalar(p, kinds, "a text scalar (e.g. "+slotHint(f.SlotType)+")", []Code{
			Var().Id("slot").Add(slotCode(f)),
			If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(
				Index().Byte().Parens(Id("v").Dot("String").Call())), Err().Op("!=").Nil()).Block(e.jsonError(p, "v", Err())),
			sel().Op("=").Op("&").Id("slot"),
		})
	case KindStruct:
		return nullOrOpen(e, p, '{', "an object", []Code{
			Var().Id("sub").Id(e.shadowFor(f.Type)),
			If(Err().Op(":=").Id("sub").Dot("decodeJSON").Call(Id("dec"), p()), Err().Op("!=").Nil()).Block(Return(Err())),
			sel().Op("=").Op("&").Id("sub"),
		})
	case KindPointer:
		inner := *f.Elem
		inner.GoName = f.GoName
		inner.Tag = f.Tag
		return e.fastField(&inner, p)
	case KindSliceScalar, KindSliceStruct, KindMapScalar, KindMapStruct:
		return e.fastCollection(f, p, sel)
	}
	panic("fastField: unhandled kind for " + f.Tag)
}

// fastScalar reads one token into v, skips null, runs read for the token
// kinds in kinds and returns a ParseError for any other kind.
func (e *emitter) fastScalar(p func() *Statement, kinds []rune, want string, read []Code) []Code {
	cases := make([]Code, 0, len(kinds))
	for _, k := range kinds {
		cases = append(cases, jsonKind(k))
	}
	return append(readToken("v"), Switch(Id("v").Dot("Kind").Call()).Block(
		Case(jsonKind('n')).Block(),
		Case(cases...).Block(read...),
		Default().Block(e.kindError(p, "v", want)),
	))
}

// readNumber declares dst holding the number token v as f's type, with a
// ParseError for the path p renders when it doesn't fit.
func (e *emitter) readNumber(f *Field, p func() *Statement, dst string) []Code {
	t := f.Type.Underlying().String()
	raw := dst
	if t != "int64" && t != "uint64" && t != "float64" {
		raw = "raw"
	}
	read := []Code{
		List(Id(raw), Err()).Op(":=").Id("v").Dot(numMethod(f.Kind)).Call(),
		If(Err().Op("!=").Nil()).Block(e.jsonError(p, "v", Err())),
	}
	read = append(read, e.rangeCheck(f, p, raw)...)
	if raw != dst {
		read = append(read, Id(dst).Op(":=").Id(t).Call(Id(raw)))
	}
	return read
}

// rangeCheck returns a ParseError for the path p renders when the decoded
// number in the variable v overflows f's type. Plain int and uint are
// checked too, since they are 32 bits on 32-bit platforms.
func (e *emitter) rangeCheck(f *Field, p func() *Statement, v string) []Code {
	if f.Bits == 64 {
		return nil
	}
	t := f.Type.Underlying().String()
	overflow := func(verb string) Code {
		return e.jsonError(p, "v", Qual("fmt", "Errorf").Call(Lit(verb+" overflows "+t), Id(v)))
	}
	switch f.Kind {
	case KindInt:
		return []Code{If(Id(v).Op("<").Qual("math", "Min"+intLimit(f)).
			Op("||").Id(v).Op(">").Qual("math", "Max"+intLimit(f))).Block(overflow("%d"))}
	case KindUint:
		return []Code{If(Id(v).Op(">").Qual("math", "Max"+intLimit(f))).Block(overflow("%d"))}
	case KindFloat:
		return []Code{If(Qual("math", "Abs").Call(Id(v)).Op(">").Qual("math", "MaxFloat32")).Block(overflow("%v"))}
	default:
		return nil
	}
}

// intLimit names the math constants that bound f's integer type, without
// the Min or Max: "Int16", or "Int" and "Uint" for the platform types.
func intLimit(f *Field) string {
	name := "Int"
	if f.Kind == KindUint {
		name = "Uint"
	}
	if f.Bits == 0 {
		return name
	}
	return name + strconv.Itoa(f.Bits)
}

// numMethod returns the jsontext.Token method that reads a number of kind k.
func numMethod(k Kind) string {
	switch k {
	case KindInt:
		return "Int"
	case KindUint:
		return "Uint"
	default:
		return "Float"
	}
}

// isComplexSlot reports whether f is a complex number, which a file may
// also spell as a plain JSON number.
func isComplexSlot(f *Field) bool {
	return f.Kind == KindStdSlot && (f.SlotType == "Complex64" || f.SlotType == "Complex128")
}

// scalarElemReader emits code that reads one scalar token into a new
// variable dst, for a list or map element of f's kind at the path p
// renders.
func (e *emitter) scalarElemReader(f *Field, p func() *Statement, dst string) []Code {
	read := readToken("v")
	check := func(want string, kinds ...rune) Code {
		var bad *Statement
		for _, k := range kinds {
			if bad != nil {
				bad = bad.Op("&&")
			} else {
				bad = Null()
			}
			bad = bad.Id("v").Dot("Kind").Call().Op("!=").Add(jsonKind(k))
		}
		return If(bad).Block(e.kindError(p, "v", want))
	}
	switch f.Kind {
	case KindString:
		return append(read, check("a string", '"'), Id(dst).Op(":=").Id("v").Dot("String").Call())
	case KindBool:
		return append(read, check("a bool", 't', 'f'), Id(dst).Op(":=").Id("v").Dot("Bool").Call())
	case KindInt, KindUint, KindFloat:
		return append(append(read, check("a number", '0')), e.readNumber(f, p, dst)...)
	case KindDuration, KindStdSlot:
		kinds := []rune{'"'}
		if isComplexSlot(f) {
			kinds = append(kinds, '0')
		}
		return append(read,
			check("a text scalar (e.g. "+slotHint(f.SlotType)+")", kinds...),
			Var().Id(dst).Add(slotCode(f)),
			If(Err().Op(":=").Id(dst).Dot("UnmarshalText").Call(Index().Byte().Parens(Id("v").Dot("String").Call())), Err().Op("!=").Nil()).Block(e.jsonError(p, "v", Err())),
		)
	default:
		panic("scalarElemReader: unsupported element kind for " + f.Tag)
	}
}

// nullOrOpen wraps body in the null check and opening token read that
// every struct, list and map field needs. null means absent, otherwise the
// next token must be open.
func nullOrOpen(e *emitter, p func() *Statement, open rune, want string, body []Code) []Code {
	return []Code{
		If(Id("dec").Dot("PeekKind").Call().Op("==").Add(jsonKind('n'))).Block(
			If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		).Else().Block(append(append(readToken("open"),
			If(Id("open").Dot("Kind").Call().Op("!=").Add(jsonKind(open))).Block(e.kindError(p, "open", want)),
		), body...)...),
	}
}

// fastCollection decodes a list or map field at path p. Each element's path
// is the list path with its index, or the map path with its quoted key.
func (e *emitter) fastCollection(f *Field, p func() *Statement, sel func() *Statement) []Code {
	list := f.Kind == KindSliceScalar || f.Kind == KindSliceStruct
	open, want, end := '{', "an object", '}'
	var elemT, coll *Statement
	switch f.Kind {
	case KindSliceScalar:
		elemT = listShadowElem(f.Elem)
	case KindMapScalar:
		elemT = listShadowElem(f.Elem)
	default:
		elemT = Id(e.shadowFor(f.Elem.Type))
	}
	elemPath := func() *Statement {
		return p().Op("+").Lit(".").Op("+").Id(e.quoteKeyName()).Call(Id("key"))
	}
	var body []Code
	if list {
		open, want, end = '[', "an array", ']'
		coll = Index().Add(elemT)
		elemPath = func() *Statement {
			return p().Op("+").Lit("[").Op("+").Qual("strconv", "Itoa").Call(Len(Id("out"))).Op("+").Lit("]")
		}
	} else {
		coll = Map(String()).Add(elemT)
		body = append(readToken("kt"), Id("key").Op(":=").Id("kt").Dot("String").Call())
	}
	store := Id("out").Index(Id("key")).Op("=").Id("el")
	if list {
		store = Id("out").Op("=").Append(Id("out"), Id("el"))
	}
	if f.Kind == KindSliceScalar || f.Kind == KindMapScalar {
		body = append(body, e.scalarElemReader(f.Elem, elemPath, "el")...)
	} else {
		body = append(body, Id("ep").Op(":=").Add(elemPath()))
		body = append(body, readToken("et")...)
		body = append(body,
			If(Id("et").Dot("Kind").Call().Op("!=").Add(jsonKind('{'))).Block(e.kindError(func() *Statement { return Id("ep") }, "et", "an object")),
			Var().Id("el").Add(elemT.Clone()),
			If(Err().Op(":=").Id("el").Dot("decodeJSON").Call(Id("dec"), Id("ep")), Err().Op("!=").Nil()).Block(Return(Err())),
		)
	}
	body = append(body, store)
	return nullOrOpen(e, p, open, want, []Code{
		Id("out").Op(":=").Add(coll).Values(),
		For(Id("dec").Dot("PeekKind").Call().Op("!=").Add(jsonKind(end))).Block(body...),
		If(List(Id("_"), Err()).Op(":=").Id("dec").Dot("ReadToken").Call(), Err().Op("!=").Nil()).Block(Return(Err())),
		sel().Op("=").Op("&").Id("out"),
	})
}
