package main

import (
	"fmt"
	"go/types"

	. "github.com/dave/jennifer/jen"
)

var outPkgPath string

// fieldGoType renders the user-side Go type for casts and temporaries.
func fieldGoType(t types.Type) *Statement {
	switch u := t.(type) {
	case *types.Named:
		pkg := u.Obj().Pkg()
		if pkg == nil || pkg.Path() == outPkgPath {
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
		return Op("*").Qual(pkgCfg, f.SlotType)
	case KindTextLeaf:
		return Op("*").Add(fieldGoType(f.Type))
	case KindStruct:
		e.ensureShadow(f)
		return Op("*").Id(e.shadowFor(f.Type))
	case KindPointer:
		return e.shadowFieldType(f.Elem)
	case KindSliceScalar:
		return Op("*").Index().Add(fieldGoType(f.Type.Underlying().(*types.Slice).Elem()))
	case KindSliceStruct:
		e.ensureShadow(f.Elem)
		return Op("*").Index().Id(e.shadowFor(f.Elem.Type))
	case KindMapScalar:
		return Op("*").Map(String()).Add(fieldGoType(f.Type.Underlying().(*types.Map).Elem()))
	case KindMapStruct:
		e.ensureShadow(f.Elem)
		return Op("*").Map(String()).Id(e.shadowFor(f.Elem.Type))
	}
	panic("shadowFieldType: unhandled kind")
}

func (e *emitter) ensureShadow(f *Field) {
	// emitShadowStruct owns the emitted-guard (marking here too would
	// cancel it and emit nothing).
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

// goName is the shadow field identifier (last element of a promoted path).
