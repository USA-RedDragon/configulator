package main

import (
	"go/types"

	. "github.com/dave/jennifer/jen"
)

// listElemOK reports whether a list or map of elem is supported.
func listElemOK(elem *Field) bool {
	switch elem.Kind {
	case KindString, KindBool, KindInt, KindUint, KindFloat, KindDuration, KindStdSlot:
		return true
	default:
		return false
	}
}

// listShadowElem renders the shadow element type of a list or map: the basic
// type under a named one, or the sentinel slot for text-decoded types.
func listShadowElem(elem *Field) *Statement {
	switch elem.Kind {
	case KindDuration, KindStdSlot:
		return slotCode(elem)
	default:
		return fieldGoType(elem.Type.Underlying())
	}
}

// listNeedsConv reports whether a list or map decoded into its shadow type
// must be converted element by element.
func listNeedsConv(elem *Field) bool {
	switch elem.Kind {
	case KindDuration, KindStdSlot:
		return true
	default:
		_, named := types.Unalias(elem.Type).(*types.Named)
		return named
	}
}

// fileList copies the decoded shadow list or map sel onto target,
// converting each element when needed.
func fileList(f *Field, sel, target *Statement, path Code) []Code {
	if !listNeedsConv(f.Elem) {
		return []Code{target.Op("=").Op("*").Add(sel)}
	}
	var conv []Code
	switch f.Elem.Kind {
	case KindDuration, KindStdSlot:
		hint := slotHint(f.Elem.SlotType)
		conv = []Code{
			List(Id("lv"), Id("ok")).Op(":=").Id("lx").Dot("Value").Call(),
			If(Op("!").Id("ok")).Block(Return(Op("&").Qual(pkgCfg, "OpaqueSpellingError").Values(Dict{
				Id("Path"): path, Id("Hint"): Lit(hint),
			}))),
			Id("lst").Index(Id("li")).Op("=").Add(convNamed(f.Elem.Type, Id("lv"))),
		}
	default:
		conv = []Code{Id("lst").Index(Id("li")).Op("=").Add(fieldGoType(f.Elem.Type)).Call(Id("lx"))}
	}
	coll := Index().Add(fieldGoType(f.Elem.Type))
	if f.Kind == KindMapScalar {
		coll = Map(String()).Add(fieldGoType(f.Elem.Type))
	}
	return []Code{Block(
		Id("lst").Op(":=").Make(coll, Len(Op("*").Add(sel.Clone()))),
		For(List(Id("li"), Id("lx")).Op(":=").Range().Op("*").Add(sel.Clone())).Block(conv...),
		target.Op("=").Id("lst"),
	)}
}

// parseList declares lst holding the strings in parts, an expression of type
// []string, parsed into f's element type. A bad element returns a
// ParseError with raw as its Value, or "(redacted)" for a secret field.
func parseList(f *Field, parts Code, path, source, raw Code) []Code {
	elem := f.Elem
	if elem.Kind == KindString && !listNeedsConv(elem) {
		return []Code{Id("lst").Op(":=").Add(parts)}
	}
	cause := Err()
	if f.Secret {
		raw, cause = Lit("(redacted)"), Qual(pkgImpl, "Redact").Call(Err(), Id("ls"))
	}
	fail := func() Code {
		return Return(Op("&").Qual(pkgCfg, "ParseError").Values(Dict{
			Id("Path"): path, Id("Source"): source, Id("Value"): raw, Id("Err"): cause.Clone(),
		}))
	}
	body := append(parseElem(elem, "ls", "lx", fail), Id("lst").Index(Id("li")).Op("=").Id("lx"))
	return []Code{
		Id("lp").Op(":=").Add(parts),
		Id("lst").Op(":=").Make(Index().Add(fieldGoType(elem.Type)), Len(Id("lp"))),
		For(List(Id("li"), Id("ls")).Op(":=").Range().Id("lp")).Block(body...),
	}
}

// parseElem parses the string variable in into a new variable out of
// elem's type, running fail on an error.
func parseElem(elem *Field, in, out string, fail func() Code) []Code {
	s := Id(in)
	check := func() Code { return If(Err().Op("!=").Nil()).Block(fail()) }
	bits := bitSize(elem)
	switch elem.Kind {
	case KindString:
		return []Code{Id(out).Op(":=").Add(convNamed(elem.Type, s))}
	case KindBool:
		return []Code{
			List(Id("pb"), Err()).Op(":=").Qual("strconv", "ParseBool").Call(s), check(),
			Id(out).Op(":=").Add(convNamed(elem.Type, Id("pb"))),
		}
	case KindInt:
		return []Code{
			List(Id("pn"), Err()).Op(":=").Qual("strconv", "ParseInt").Call(s, Lit(10), bits), check(),
			Id(out).Op(":=").Add(fieldGoType(elem.Type)).Call(Id("pn")),
		}
	case KindUint:
		return []Code{
			List(Id("pn"), Err()).Op(":=").Qual("strconv", "ParseUint").Call(s, Lit(10), bits), check(),
			Id(out).Op(":=").Add(fieldGoType(elem.Type)).Call(Id("pn")),
		}
	case KindFloat:
		return []Code{
			List(Id("pn"), Err()).Op(":=").Qual("strconv", "ParseFloat").Call(s, bits), check(),
			Id(out).Op(":=").Add(fieldGoType(elem.Type)).Call(Id("pn")),
		}
	case KindDuration:
		return []Code{List(Id(out), Err()).Op(":=").Qual("time", "ParseDuration").Call(s), check()}
	case KindStdSlot:
		return []Code{
			Var().Id("slot").Add(slotCode(elem)),
			If(Err().Op(":=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(s)), Err().Op("!=").Nil()).Block(fail()),
			List(Id("sv"), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
			Id(out).Op(":=").Add(convNamed(elem.Type, Id("sv"))),
		}
	default:
		panic("parseElem: unsupported element kind for " + elem.Tag)
	}
}

// hop is an optional struct on the way from the config root to a leaf.
type hop struct {
	f    *Field
	path string
}

// leaf is a field set by a single env var or flag: its name segments, its
// dotted origin path and the optional structs above it, outermost first.
type leaf struct {
	f     *Field
	segs  []string
	path  string
	chain []hop
}

// leaves returns the fields env vars or flags can set, in order, naming
// each level with seg. A field skip accepts is left out with its subtree.
// Collections are file-only.
func leaves(fields []*Field, seg func(*Field) string, skip func(*Field) bool) []leaf {
	var out []leaf
	var walk func(fields []*Field, segs []string, path string, chain []hop)
	walk = func(fields []*Field, segs []string, path string, chain []hop) {
		for _, f := range fields {
			if skip(f) {
				continue
			}
			s2 := append(append([]string{}, segs...), seg(f))
			p2 := joinPath(path, f.Tag)
			switch {
			case f.Kind == KindStruct:
				walk(f.Fields, s2, p2, chain)
			case f.Kind == KindPointer && f.Elem.Kind == KindStruct:
				walk(f.Elem.Fields, s2, p2, append(append([]hop{}, chain...), hop{f, p2}))
			case f.Kind == KindSliceStruct || f.Kind == KindMapStruct || f.Kind == KindMapScalar:
			default:
				out = append(out, leaf{f, s2, p2, chain})
			}
		}
	}
	walk(fields, nil, "", nil)
	return out
}

// chainAssign returns assign applied to l's field. For a field inside
// optional structs, each struct is copied, or built with its element
// defaults if nil, and stored back after the write, so a layer that fails
// later never modifies a struct the previous layers' config shares. sep is
// the list separator for the element defaults.
func (e *emitter) chainAssign(l leaf, sep Code, assign func(target *Statement) []Code) []Code {
	if len(l.chain) == 0 {
		return assign(cfgSel("cfg", l.f))
	}
	var pre, post []Code
	prev := ""
	for i, h := range l.chain {
		v := depthName("e", i)
		sel := cfgSel("cfg", h.f)
		if prev != "" {
			sel = relSel(Id(prev), h.f)
		}
		copyIn := If(sel.Clone().Op("!=").Nil()).Block(Id(v).Op("=").Op("*").Add(sel.Clone()))
		if defaults := e.elemDefaults(h.f.Elem.Fields, Id(v), pathExpr{suffix: h.path}, sep); len(defaults) > 0 {
			copyIn = copyIn.Else().Block(defaults...)
		}
		pre = append(pre, Var().Id(v).Add(fieldGoType(h.f.Elem.Type)), copyIn)
		post = append([]Code{sel.Clone().Op("=").Op("&").Id(v)}, post...)
		prev = v
	}
	pre = append(pre, assign(relSel(Id(prev), l.f))...)
	return append(pre, post...)
}
