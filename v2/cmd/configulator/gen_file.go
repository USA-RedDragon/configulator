package main

import (
	"fmt"
	"go/types"
	"strings"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitDecodeFile() {
	n := e.m.TypeName
	e.f.Func().Id(lowerFirst(n)+"DecodeFile").Params(
		Id("data").Index().Byte(), Id("u").Qual(pkgCfg, "Unmarshal"),
		Id("cfg").Op("*").Id(n), Id("set").Qual(pkgCfg, "SetOrigin"), Id("file").String(),
	).Error().Block(
		Var().Id("sh").Id(shadowName(n)),
		If(Err().Op(":=").Id("u").Call(Id("data"), Op("&").Id("sh")), Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "DecodeError").Values(Dict{Id("Path"): Id("file"), Id("Err"): Err()})),
		),
		Return(Id("sh").Dot("applyTo").Call(Id("cfg"), Id("set"), Id("file"))),
	)
}

// ---- file layer: applyTo ----

func (e *emitter) emitApplyTo(shadow string, fields []*Field, pathPrefix, recv string) {
	var body []Code
	body = append(body, e.applyToFields(fields, "s", pathPrefix)...)
	body = append(body, Return(Nil()))
	e.f.Func().Params(Id("s").Op("*").Id(shadow)).Id("applyTo").Params(
		Id("cfg").Op("*").Id(e.typeForShadow(shadow)), Id("set").Qual(pkgCfg, "SetOrigin"), Id("file").String(),
	).Error().Block(body...)
}

// typeForShadow maps a shadow name back to its config type name.
func (e *emitter) typeForShadow(shadow string) string {
	base := strings.TrimSuffix(shadow, "Shadow")
	if base == lowerFirst(e.m.TypeName) {
		return e.m.TypeName
	}
	// named struct: shadow is lowerFirst(Name)+"Shadow"
	r := []rune(base)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}

func (e *emitter) applyToFields(fields []*Field, src, pathPrefix string) []Code {
	var out []Code
	for _, f := range fields {
		path := joinPath(pathPrefix, f.Tag)
		sel := Id(src).Dot(goName(f))
		recFile := Id("set").Call(Lit(path), Qual(pkgCfg, "LayerFile"), Id("file"))
		switch f.Kind {
		case KindString, KindBool, KindInt, KindUint, KindFloat, KindTextLeaf:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				cfgSel("cfg", f).Op("=").Op("*").Add(sel.Clone()), recFile,
			))
		case KindDuration, KindStdSlot:
			hint := `"30s"`
			if f.Kind == KindStdSlot {
				hint = slotHint(f.SlotType)
			}
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				List(Id("v"), Id("ok")).Op(":=").Add(sel.Clone()).Dot("Value").Call(),
				If(Op("!").Id("ok")).Block(
					Return(Op("&").Qual(pkgCfg, "OpaqueSpellingError").Values(Dict{
						Id("Path"): Lit(path), Id("Hint"): Lit(hint),
					})),
				),
				cfgSel("cfg", f).Op("=").Add(castStd(f)),
				recFile,
			))
		case KindStruct:
			inner := e.applyToFields(f.Fields, src+"."+goName(f), path)
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(inner...))
			e.emitNestedApplyTo(f)
		case KindPointer:
			out = append(out, e.applyToPointer(f, src, path)...)
		case KindSliceScalar, KindMapScalar:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				cfgSel("cfg", f).Op("=").Op("*").Add(sel.Clone()), recFile,
			))
		case KindSliceStruct:
			out = append(out, e.applyToSliceStruct(f, src, path)...)
		case KindMapStruct:
			out = append(out, e.applyToMapStruct(f, src, path)...)
		}
	}
	return out
}

// emitNestedApplyTo: nested named structs get inline field application in
// the parent's applyTo (deep-merge); nothing extra to emit, but shadows
// for slice/map elements DO need their own defaults inline (handled in
// applyToSliceStruct/applyToMapStruct).
func (e *emitter) emitNestedApplyTo(*Field) {}

func slotHint(slot string) string {
	hints := map[string]string{
		"Duration": `"30s"`, "IPNet": `"10.0.0.0/8"`, "FileMode": `"0644"`,
		"Location": `"America/New_York"`, "TCPAddr": `"127.0.0.1:8080"`,
		"UDPAddr": `"127.0.0.1:53"`, "HardwareAddr": `"aa:bb:cc:dd:ee:ff"`,
		"URL": `"https://example.com"`, "Month": `"January"`,
	}
	if h, ok := hints[slot]; ok {
		return h
	}
	return `"..."`
}

// castStd converts a slot value to the user's (possibly named) type.
func castStd(f *Field) *Statement {
	if named, ok := f.Type.(*types.Named); ok && named.Obj().Name() != "Duration" {
		// e.g. type Timeout time.Duration — cast through the named type
		if named.Obj().Pkg() != nil && named.Obj().Pkg().Path() != "time" &&
			named.Obj().Pkg().Path() != "net" && named.Obj().Pkg().Path() != "os" &&
			named.Obj().Pkg().Path() != "net/url" {
			return fieldGoType(f.Type).Call(Id("v"))
		}
	}
	return Id("v")
}

func (e *emitter) applyToPointer(f *Field, src, path string) []Code {
	sel := Id(src).Dot(goName(f))
	recFile := Id("set").Call(Lit(path), Qual(pkgCfg, "LayerFile"), Id("file"))
	switch f.Elem.Kind {
	case KindString, KindBool, KindInt, KindUint, KindFloat:
		return []Code{If(sel.Clone().Op("!=").Nil()).Block(
			Id("v").Op(":=").Op("*").Add(sel.Clone()).Comment("// never alias the shadow"),
			cfgSel("cfg", f).Op("=").Op("&").Id("v"),
			recFile,
		)}
	case KindStruct:
		elemType := fieldGoType(f.Elem.Type)
		var inner []Code
		inner = append(inner, Id("e").Op(":=").Add(elemType.Clone()).Values())
		inner = append(inner, e.elementDefaults(f.Elem.Fields, path)...)
		inner = append(inner, If(cfgSel("cfg", f).Op("!=").Nil()).Block(
			Id("e").Op("=").Op("*").Add(cfgSel("cfg", f)).Comment("// deep-merge onto the existing value (cloned)"),
		))
		inner = append(inner, e.applyElementFields(f.Elem.Fields, src+"."+goName(f), "e", path)...)
		inner = append(inner, cfgSel("cfg", f).Op("=").Op("&").Id("e"))
		return []Code{If(sel.Clone().Op("!=").Nil()).Block(inner...)}
	}
	panic("applyToPointer: unsupported element kind at " + path)
}

// elementDefaults emits element-default assignments + origin records for a
// freshly constructed element (SPEC rule 3).
func (e *emitter) elementDefaults(fields []*Field, pathExpr string) []Code {
	var out []Code
	for _, f := range fields {
		if f.Default == "" {
			continue
		}
		switch f.Kind {
		case KindString:
			out = append(out, Id("e").Dot(goName(f)).Op("=").Lit(f.Default))
		case KindUint, KindInt:
			v := 0
			fmt.Sscanf(f.Default, "%d", &v)
			out = append(out, Id("e").Dot(goName(f)).Op("=").Add(castLit(f, v)))
		default:
			panic("element default kind not supported yet")
		}
		out = append(out, Id("set").Call(
			Lit(pathExpr+"."+f.Tag), Qual(pkgCfg, "LayerDefault"), Lit("element default"),
		))
	}
	return out
}

// dynamic variant for indexed/keyed paths where the path is an expression
func (e *emitter) elementDefaultsExpr(fields []*Field, pathVar string) []Code {
	var out []Code
	for _, f := range fields {
		if f.Default == "" {
			continue
		}
		switch f.Kind {
		case KindString:
			out = append(out, Id("e").Dot(goName(f)).Op("=").Lit(f.Default))
		case KindUint, KindInt:
			v := 0
			fmt.Sscanf(f.Default, "%d", &v)
			out = append(out, Id("e").Dot(goName(f)).Op("=").Add(castLit(f, v)))
		default:
			panic("element default kind not supported yet")
		}
		out = append(out, Id("set").Call(
			Id(pathVar).Op("+").Lit("."+f.Tag), Qual(pkgCfg, "LayerDefault"), Lit("element default"),
		))
	}
	return out
}

func (e *emitter) applyElementFields(fields []*Field, src, dst, path string) []Code {
	var out []Code
	for _, f := range fields {
		sel := Id(src).Dot(goName(f))
		switch f.Kind {
		case KindString, KindBool, KindInt, KindUint, KindFloat:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				Id(dst).Dot(goName(f)).Op("=").Op("*").Add(sel.Clone()),
				Id("set").Call(Lit(path+"."+f.Tag), Qual(pkgCfg, "LayerFile"), Id("file")),
			))
		default:
			panic("nested element field kind not supported yet")
		}
	}
	return out
}

// element fields with a dynamic path expression (idx/key variable)
func (e *emitter) applyElementFieldsExpr(fields []*Field, src, dst, pathVar string) []Code {
	var out []Code
	for _, f := range fields {
		sel := Id(src).Dot(goName(f))
		out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
			Id(dst).Dot(goName(f)).Op("=").Op("*").Add(sel.Clone()),
			Id("set").Call(Id(pathVar).Op("+").Lit("."+f.Tag), Qual(pkgCfg, "LayerFile"), Id("file")),
		))
	}
	return out
}

func (e *emitter) applyToSliceStruct(f *Field, src, path string) []Code {
	sel := Id(src).Dot(goName(f))
	elemType := fieldGoType(f.Elem.Type)
	loop := []Code{
		Var().Id("e").Add(elemType.Clone()),
		Id("idx").Op(":=").Lit(path+"[").Op("+").Qual("strconv", "Itoa").Call(Id("i")).Op("+").Lit("]"),
	}
	loop = append(loop, e.elementDefaultsExpr(f.Elem.Fields, "idx")...)
	loop = append(loop, e.applyElementFieldsExpr(f.Elem.Fields, "esh", "e", "idx")...)
	loop = append(loop, Id("out").Index(Id("i")).Op("=").Id("e"))
	return []Code{If(sel.Clone().Op("!=").Nil()).Block(
		Id("out").Op(":=").Make(Index().Add(elemType.Clone()), Len(Op("*").Add(sel.Clone()))),
		For(List(Id("i"), Id("esh")).Op(":=").Range().Op("*").Add(sel.Clone())).Block(loop...),
		cfgSel("cfg", f).Op("=").Id("out"),
		Id("set").Call(Lit(path), Qual(pkgCfg, "LayerFile"), Id("file")),
	)}
}

func (e *emitter) applyToMapStruct(f *Field, src, path string) []Code {
	sel := Id(src).Dot(goName(f))
	elemType := fieldGoType(f.Elem.Type)
	loop := []Code{
		Var().Id("e").Add(elemType.Clone()),
		Id("key").Op(":=").Lit(path + ".").Op("+").Id(e.quoteKeyName()).Call(Id("k")),
	}
	loop = append(loop, e.elementDefaultsExpr(f.Elem.Fields, "key")...)
	loop = append(loop, e.applyElementFieldsExpr(f.Elem.Fields, "esh", "e", "key")...)
	loop = append(loop, Id("out").Index(Id("k")).Op("=").Id("e"))
	return []Code{If(sel.Clone().Op("!=").Nil()).Block(
		Id("out").Op(":=").Make(Map(String()).Add(elemType.Clone()), Len(Op("*").Add(sel.Clone()))),
		For(List(Id("k"), Id("esh")).Op(":=").Range().Op("*").Add(sel.Clone())).Block(loop...),
		cfgSel("cfg", f).Op("=").Id("out"),
		Id("set").Call(Lit(path), Qual(pkgCfg, "LayerFile"), Id("file")),
	)}
}

func joinPath(prefix, tag string) string {
	if prefix == "" {
		return tag
	}
	return prefix + "." + tag
}
