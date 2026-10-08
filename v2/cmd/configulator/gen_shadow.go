package main

import (
	"fmt"
	"go/types"

	. "github.com/dave/jennifer/jen"
)

// convNamed wraps expr in a conversion to t when t is a named type over a
// basic type (e.g. `type LogLevel string`). Decoded values arrive as the
// underlying type, and Go won't assign that to the named type without a
// conversion. Other types pass through unchanged, as do the stdlib types
// configulator has wrappers for.
func convNamed(t types.Type, expr *Statement) *Statement {
	t = types.Unalias(t)
	if named, ok := t.(*types.Named); ok && !isStdSlot(named) {
		if _, basic := named.Underlying().(*types.Basic); basic {
			return fieldGoType(t).Call(expr)
		}
	}
	if b, ok := t.(*types.Basic); ok && b.Kind() == types.Uintptr {
		return Id("uintptr").Call(expr)
	}
	return expr
}

// slotCode renders the impl slot type that decodes f.
func slotCode(f *Field) *Statement {
	if f.SlotType == slotText {
		return Qual(pkgImpl, slotText).Types(fieldGoType(f.Type), Op("*").Add(fieldGoType(f.Type)))
	}
	return Qual(pkgImpl, f.SlotType)
}

// isStdSlot reports whether n is decoded by a configulator wrapper type, whose
// Value already returns n.
func isStdSlot(n *types.Named) bool {
	pkg := n.Obj().Pkg()
	if pkg == nil {
		return false
	}
	_, ok := stdSlot(pkg.Path() + "." + n.Obj().Name())
	return ok
}

// fieldGoType renders the user-side Go type for casts and temporaries.
func fieldGoType(t types.Type) *Statement {
	switch u := types.Unalias(t).(type) {
	case *types.Named:
		pkg := u.Obj().Pkg()
		if pkg == nil {
			return Id(u.Obj().Name())
		}
		return Qual(pkg.Path(), u.Obj().Name())
	case *types.Basic:
		return Id(u.Name())
	case *types.Slice:
		return Index().Add(fieldGoType(u.Elem()))
	case *types.Map:
		return Map(fieldGoType(u.Key())).Add(fieldGoType(u.Elem()))
	case *types.Pointer:
		return Op("*").Add(fieldGoType(u.Elem()))
	}
	panic(fmt.Sprintf("fieldGoType: %T", t))
}

// shadowFieldType renders a shadow field's type.
func (e *emitter) shadowFieldType(f *Field) *Statement {
	switch f.Kind {
	case KindString:
		return Op("*").String()
	case KindBool:
		return Op("*").Bool()
	case KindInt, KindUint, KindFloat:
		return Op("*").Add(fieldGoType(f.Type.Underlying()))
	case KindDuration, KindStdSlot:
		return Op("*").Add(slotCode(f))
	case KindStruct:
		e.ensureShadow(f)
		return Op("*").Id(e.shadowFor(f.Type))
	case KindPointer:
		return e.shadowFieldType(f.Elem)
	case KindSliceScalar:
		return Op("*").Index().Add(listShadowElem(f.Elem))
	case KindSliceStruct:
		e.ensureShadow(f.Elem)
		return Op("*").Index().Id(e.shadowFor(f.Elem.Type))
	case KindMapScalar:
		return Op("*").Map(String()).Add(listShadowElem(f.Elem))
	case KindMapStruct:
		e.ensureShadow(f.Elem)
		return Op("*").Map(String()).Id(e.shadowFor(f.Elem.Type))
	}
	panic("shadowFieldType: unhandled kind")
}

func (e *emitter) ensureShadow(f *Field) {
	// emitShadowStruct marks the shadow as emitted itself. Marking it here
	// too would make it skip the struct and emit nothing.
	e.emitShadowStruct(e.shadowFor(f.Type), f.Fields)
}

func (e *emitter) emitShadowStruct(name string, fields []*Field) {
	if e.shadows[name] {
		return
	}
	e.shadows[name] = true
	e.shadowOrder = append(e.shadowOrder, name)
	e.shadowFields[name] = fields
	var defs []Code
	for _, f := range fields {
		tags := map[string]string{"json": f.Tag, "yaml": f.Tag, "toml": f.Tag}
		defs = append(defs, Id(goName(f)).Add(e.shadowFieldType(f)).Tag(tags))
	}
	e.f.Type().Id(name).Struct(defs...)
}
