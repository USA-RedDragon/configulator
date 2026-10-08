package main

import (
	"go/types"
	"strconv"
	"strings"
	"time"

	. "github.com/dave/jennifer/jen"
)

func (e *emitter) emitApplyDefaults() {
	n := e.m.TypeName
	var body []Code
	var walk func(fields []*Field, pathPrefix []string)
	walk = func(fields []*Field, pathPrefix []string) {
		for _, f := range fields {
			path := strings.Join(append(append([]string{}, pathPrefix...), f.Tag), ".")
			if f.Kind == KindStruct {
				walk(f.Fields, append(pathPrefix, f.Tag))
				continue
			}
			if f.Default == "" {
				continue
			}
			body = append(body, e.defaultAssign(f, path)...)
		}
	}
	walk(e.m.Fields, nil)
	body = append(body, Return(Nil()))
	e.f.Func().Id(lowerFirst(n)+"ApplyDefaults").Params(
		Id("cfg").Op("*").Id(n), Id("set").Qual(pkgCfg, "SetOrigin"),
	).Error().Block(body...)
}

func (e *emitter) defaultAssign(f *Field, path string) []Code {
	rec := Id("set").Call(Lit(path), Qual(pkgCfg, "LayerDefault"), Lit("default tag"))
	if f.Kind == KindPointer {
		tmp := lowerFirst(strings.ReplaceAll(goName(f), ".", "")) + "Default"
		prep, val := defaultValue(f.Elem, f.Default, path)
		prep = append(prep, Id(tmp).Op(":=").Add(val), cfgSel("cfg", f).Op("=").Op("&").Id(tmp), rec)
		return []Code{Block(prep...)}
	}
	prep, val := defaultValue(f, f.Default, path)
	if len(prep) == 0 {
		return []Code{cfgSel("cfg", f).Op("=").Add(val), rec}
	}
	prep = append(prep, cfgSel("cfg", f).Op("=").Add(val), rec)
	return []Code{Block(prep...)}
}

// defaultValue returns statements to run first and an expression of f's type
// holding the default. checkDefault has already validated def.
func defaultValue(f *Field, def, path string) ([]Code, *Statement) {
	switch f.Kind {
	case KindString:
		return nil, convNamed(f.Type, Lit(def))
	case KindBool:
		v, _ := strconv.ParseBool(def)
		return nil, convNamed(f.Type, Lit(v))
	case KindInt, KindUint:
		return nil, castLit(f, intLit(f.Kind, def))
	case KindFloat:
		v, _ := strconv.ParseFloat(def, 64)
		return nil, convNamed(f.Type, Lit(v))
	case KindDuration:
		d, _ := time.ParseDuration(def)
		return nil, Qual("time", "Duration").Call(Lit(int64(d)))
	case KindStdSlot:
		return []Code{
			Var().Id("slot").Qual(pkgCfg, f.SlotType),
			Id("_").Op("=").Id("slot").Dot("UnmarshalText").Call(Index().Byte().Parens(Lit(def))),
			List(Id("v"), Id("_")).Op(":=").Id("slot").Dot("Value").Call(),
		}, Id("v")
	case KindSliceScalar:
		parts := strings.Split(def, ",")
		lits := make([]Code, 0, len(parts))
		for _, el := range parts {
			lits = append(lits, Lit(el))
		}
		return nil, fieldGoType(f.Type.Underlying()).Values(lits...)
	default:
		panic("defaultValue: unhandled kind for " + path)
	}
}

// intLit renders a validated integer default as an untyped constant.
func intLit(k Kind, def string) *Statement {
	if k == KindUint {
		v, _ := strconv.ParseUint(def, 10, 64)
		return Op(strconv.FormatUint(v, 10))
	}
	v, _ := strconv.ParseInt(def, 10, 64)
	return Lit(int(v))
}

// castLit wraps an integer literal in a conversion for sized types.
func castLit(f *Field, lit *Statement) *Statement {
	t, ok := f.Type.Underlying().(*types.Basic)
	if !ok || t.Kind() == types.Int {
		return lit
	}
	if _, named := f.Type.(*types.Named); named {
		return fieldGoType(f.Type).Call(lit)
	}
	return Id(t.Name()).Call(lit)
}
