package main

import (
	"go/types"
	"strconv"
	"strings"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitDecodeFile() {
	n := e.m.TypeName
	e.decl().Func().Id(lowerFirst(n)+"DecodeFile").Params(
		Id("data").Index().Byte(), Id("u").Qual(pkgCfg, "Unmarshal"),
		Id("cfg").Op("*").Id(n), Id("sep").String(), Id("set").Qual(pkgCfg, "SetOrigin"), Id("file").String(),
	).Error().Block(
		Var().Id("sh").Id(shadowName(n)),
		If(Err().Op(":=").Id("u").Call(Id("data"), Op("&").Id("sh")), Err().Op("!=").Nil()).Block(
			Return(Op("&").Qual(pkgCfg, "DecodeError").Values(Dict{Id("Path"): Id("file"), Id("Err"): Err()})),
		),
		Return(Id("sh").Dot("applyTo").Call(Id("cfg"), Id("sep"), Id("set"), Id("file"))),
	)
}

func (e *emitter) emitApplyTo(shadow string, fields []*Field, pathPrefix string) {
	body := append(e.applyToFields(fields, "s", pathPrefix), Return(Nil()))
	e.decl().Func().Params(Id("s").Op("*").Id(shadow)).Id("applyTo").Params(
		Id("cfg").Op("*").Id(e.typeForShadow(shadow)), Id("sep").String(), Id("set").Qual(pkgCfg, "SetOrigin"), Id("file").String(),
	).Error().Block(body...)
}

// typeForShadow maps a shadow name back to its config type name.
func (e *emitter) typeForShadow(shadow string) string {
	base := strings.TrimSuffix(shadow, "Shadow")
	if base == lowerFirst(e.m.TypeName) {
		return e.m.TypeName
	}
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
		case KindString, KindBool, KindInt, KindUint, KindFloat:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				cfgSel("cfg", f).Op("=").Add(convNamed(f.Type, Op("*").Add(sel.Clone()))), recFile,
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
		case KindSliceScalar:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				append(fileList(f, sel.Clone(), cfgSel("cfg", f), Lit(path)), recFile)...,
			))
		case KindMapScalar:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				append(fileList(f, sel.Clone(), cfgSel("cfg", f), Lit(path)), recFile)...,
			))
		case KindSliceStruct:
			out = append(out, e.collApply(f, sel, cfgSel("cfg", f), pathExpr{suffix: path}, 0)...)
		case KindMapStruct:
			out = append(out, e.collApply(f, sel, cfgSel("cfg", f), pathExpr{suffix: path}, 0)...)
		default:
		}
	}
	return out
}

// emitNestedApplyTo emits nothing: nested structs are merged inline in the
// parent's applyTo, and collApply sets element defaults.
func (e *emitter) emitNestedApplyTo(*Field) {}

func slotHint(slot string) string {
	hints := map[string]string{
		"Duration": `"30s"`, "IPNet": `"10.0.0.0/8"`, "FileMode": `"0644"`,
		"Location": `"America/New_York"`, "TCPAddr": `"127.0.0.1:8080"`,
		"UDPAddr": `"127.0.0.1:53"`, "HardwareAddr": `"aa:bb:cc:dd:ee:ff"`,
		"URL": `"https://example.com"`, "Month": `"January"`,
		"Complex64": `"1+2i"`, "Complex128": `"1+2i"`,
	}
	if h, ok := hints[slot]; ok {
		return h
	}
	return `"..."`
}

// castStd converts a slot value to the user's (possibly named) type.
func castStd(f *Field) *Statement {
	if named, ok := f.Type.(*types.Named); ok && named.Obj().Name() != "Duration" {
		if named.Obj().Pkg() != nil && named.Obj().Pkg().Path() != "time" &&
			named.Obj().Pkg().Path() != "net" && named.Obj().Pkg().Path() != "os" &&
			named.Obj().Pkg().Path() != "net/url" {
			return fieldGoType(f.Type).Call(Id("v"))
		}
	}
	return Id("v")
}

func (e *emitter) applyToPointer(f *Field, src, path string) []Code {
	return e.pointerApply(f, Id(src).Dot(goName(f)), cfgSel("cfg", f), pathExpr{suffix: path}, 0)
}

// pathExpr is an origin path: a literal known at generate time, or a
// runtime string variable followed by a literal suffix.
type pathExpr struct {
	v      string
	suffix string
}

func (p pathExpr) child(tag string) pathExpr {
	return pathExpr{v: p.v, suffix: p.suffix + "." + tag}
}

// with returns the path followed by lit as a single expression.
func (p pathExpr) with(lit string) *Statement {
	if p.v == "" {
		return Lit(p.suffix + lit)
	}
	if p.suffix+lit == "" {
		return Id(p.v)
	}
	return Id(p.v).Op("+").Lit(p.suffix + lit)
}

func (p pathExpr) code() *Statement { return p.with("") }

func depthName(name string, depth int) string {
	if depth == 0 {
		return name
	}
	return name + strconv.Itoa(depth)
}

// elemDefaults sets the defaults of a newly constructed element, recursing
// into nested structs, and records their origin. sep splits list defaults.
func (e *emitter) elemDefaults(fields []*Field, dst *Statement, p pathExpr, sep Code) []Code {
	var out []Code
	for _, f := range fields {
		target := dst.Clone().Dot(goName(f))
		fp := p.child(f.Tag)
		if f.Kind == KindStruct {
			out = append(out, e.elemDefaults(f.Fields, target, fp, sep)...)
			continue
		}
		if f.Default == "" {
			continue
		}
		var assign []Code
		if f.Kind == KindPointer {
			prep, val := defaultValue(pointee(f), f.Default, fp.code(), sep)
			assign = append(assign, prep...)
			assign = append(assign, Id("d").Op(":=").Add(val), target.Op("=").Op("&").Id("d"))
		} else {
			prep, val := defaultValue(f, f.Default, fp.code(), sep)
			assign = append(assign, prep...)
			assign = append(assign, target.Op("=").Add(val))
		}
		if len(assign) == 1 {
			out = append(out, assign[0])
		} else {
			out = append(out, Block(assign...))
		}
		out = append(out, Id("set").Call(fp.code(), Qual(pkgCfg, "LayerDefault"), Lit("element default")))
	}
	return out
}

// elemApply copies the fields present in the shadow src onto dst.
func (e *emitter) elemApply(fields []*Field, src, dst *Statement, p pathExpr, depth int) []Code {
	out := make([]Code, 0, len(fields))
	for _, f := range fields {
		sel := src.Clone().Dot(goName(f))
		target := dst.Clone().Dot(goName(f))
		fp := p.child(f.Tag)
		rec := Id("set").Call(fp.code(), Qual(pkgCfg, "LayerFile"), Id("file"))
		switch f.Kind {
		case KindString, KindBool, KindInt, KindUint, KindFloat:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				target.Op("=").Add(convNamed(f.Type, Op("*").Add(sel.Clone()))), rec,
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
						Id("Path"): fp.code(), Id("Hint"): Lit(hint),
					})),
				),
				target.Op("=").Add(castStd(f)), rec,
			))
		case KindSliceScalar:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				append(fileList(f, sel.Clone(), target, fp.code()), rec)...,
			))
		case KindMapScalar:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				append(fileList(f, sel.Clone(), target, fp.code()), rec)...,
			))
		case KindStruct:
			out = append(out, If(sel.Clone().Op("!=").Nil()).Block(
				e.elemApply(f.Fields, sel, target, fp, depth)...,
			))
		case KindPointer:
			out = append(out, e.pointerApply(f, sel, target, fp, depth)...)
		case KindSliceStruct, KindMapStruct:
			out = append(out, e.collApply(f, sel, target, fp, depth)...)
		default:
		}
	}
	return out
}

// pointerApply copies an optional field from the shadow, allocating the
// target on first write.
func (e *emitter) pointerApply(f *Field, sel, target *Statement, p pathExpr, depth int) []Code {
	rec := Id("set").Call(p.code(), Qual(pkgCfg, "LayerFile"), Id("file"))
	switch f.Elem.Kind {
	case KindString, KindBool, KindInt, KindUint, KindFloat:
		return []Code{If(sel.Clone().Op("!=").Nil()).Block(
			Id("v").Op(":=").Add(convNamed(f.Elem.Type, Op("*").Add(sel.Clone()))),
			target.Clone().Op("=").Op("&").Id("v"),
			rec,
		)}
	case KindDuration, KindStdSlot:
		return []Code{If(sel.Clone().Op("!=").Nil()).Block(
			List(Id("v"), Id("ok")).Op(":=").Add(sel.Clone()).Dot("Value").Call(),
			If(Op("!").Id("ok")).Block(
				Return(Op("&").Qual(pkgCfg, "OpaqueSpellingError").Values(Dict{
					Id("Path"): p.code(), Id("Hint"): Lit(slotHint(f.Elem.SlotType)),
				})),
			),
			Id("pv").Op(":=").Add(castStd(f.Elem)),
			target.Clone().Op("=").Op("&").Id("pv"),
			rec,
		)}
	case KindStruct:
		n := depthName("e", depth)
		defaults := e.elemDefaults(f.Elem.Fields, Id(n), p, Id("sep"))
		apply := e.elemApply(f.Elem.Fields, sel, Id(n), p, depth+1)
		inner := make([]Code, 0, 3+len(defaults)+len(apply))
		inner = append(inner, Id(n).Op(":=").Add(fieldGoType(f.Elem.Type)).Values())
		inner = append(inner, defaults...)
		inner = append(inner, If(target.Clone().Op("!=").Nil()).Block(
			Id(n).Op("=").Op("*").Add(target.Clone()),
		))
		inner = append(inner, apply...)
		inner = append(inner, target.Clone().Op("=").Op("&").Id(n))
		return []Code{If(sel.Clone().Op("!=").Nil()).Block(inner...)}
	default:
		panic("pointerApply: unsupported element kind at " + p.suffix)
	}
}

// collApply rebuilds a list or map of structs from its shadow, applying
// element defaults to each element first.
func (e *emitter) collApply(f *Field, sel, target *Statement, p pathExpr, depth int) []Code {
	elemType := fieldGoType(f.Elem.Type)
	el, esh, outName := depthName("e", depth), depthName("esh", depth), depthName("out", depth)
	var idx, keyVar string
	var keyStmt, makeOut Code
	var rangeVars *Statement
	if f.Kind == KindSliceStruct {
		i := depthName("i", depth)
		idx = depthName("idx", depth)
		keyStmt = Id(idx).Op(":=").Add(p.with("[")).Op("+").Qual("strconv", "Itoa").Call(Id(i)).Op("+").Lit("]")
		makeOut = Id(outName).Op(":=").Make(Index().Add(elemType.Clone()), Len(Op("*").Add(sel.Clone())))
		rangeVars = List(Id(i), Id(esh))
		keyVar = i
	} else {
		k := depthName("k", depth)
		idx = depthName("key", depth)
		keyStmt = Id(idx).Op(":=").Add(p.with(".")).Op("+").Id(e.quoteKeyName()).Call(Id(k))
		makeOut = Id(outName).Op(":=").Make(Map(String()).Add(elemType.Clone()), Len(Op("*").Add(sel.Clone())))
		rangeVars = List(Id(k), Id(esh))
		keyVar = k
	}
	ep := pathExpr{v: idx}
	defaults := e.elemDefaults(f.Elem.Fields, Id(el), ep, Id("sep"))
	apply := e.elemApply(f.Elem.Fields, Id(esh), Id(el), ep, depth+1)
	loop := make([]Code, 0, 3+len(defaults)+len(apply))
	loop = append(loop, Var().Id(el).Add(elemType.Clone()), keyStmt)
	loop = append(loop, defaults...)
	loop = append(loop, apply...)
	loop = append(loop, Id(outName).Index(Id(keyVar)).Op("=").Id(el))
	return []Code{If(sel.Clone().Op("!=").Nil()).Block(
		makeOut,
		For(rangeVars.Op(":=").Range().Op("*").Add(sel.Clone())).Block(loop...),
		target.Clone().Op("=").Id(outName),
		Id("set").Call(p.code(), Qual(pkgCfg, "LayerFile"), Id("file")),
	)}
}

func joinPath(prefix, tag string) string {
	if prefix == "" {
		return tag
	}
	return prefix + "." + tag
}
