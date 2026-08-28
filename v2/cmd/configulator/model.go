// Package main implements the configulator code generator.
package main

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Kind classifies a field for emission.
type Kind int

const (
	KindString Kind = iota
	KindBool
	KindInt   // any signed int type
	KindUint  // any unsigned int type
	KindFloat // float32/float64
	KindStruct
	KindPointer     // pointer to scalar or struct (optionals, SPEC rule 5)
	KindSliceScalar // slice of scalars
	KindSliceStruct // list of structs (file-only, SPEC rule 6)
	KindMapScalar   // map[string]scalar (file-only)
	KindMapStruct   // map[string]struct (file-only)
	KindDuration    // time.Duration -> configulator.Duration slot
	KindStdSlot     // other stdtypes sentinel slot
	KindTextLeaf    // non-struct kind implementing TextUnmarshaler: safe naked leaf
)

// Field is the generator IR for one config field.
type Field struct {
	GoName   string
	Tag      string // configulator name (name: tag, json/yaml fallback)
	Kind     Kind
	Type     types.Type
	Bits     int    // int/uint/float width (0 = platform int)
	Default  string // raw default: tag ("" = none)
	Desc     string
	EnvSkip  bool   // env:"-"
	FlagSkip bool   // flag:"-"
	EnvName  string // env:"NAME" explicit opt-in (error on collections)
	FlagName string // flag:"name"
	Required bool
	Secret   bool
	Short    string
	Opaque   bool     // opaque:"true"
	SlotType string   // configulator slot type name for KindDuration/KindStdSlot
	Fields   []*Field // struct / element fields
	Elem     *Field   // element for slices/maps/pointers
	Embedded bool     // flattened embedded struct marker (fields promoted)
}

// Model is the IR for one config type.
type Model struct {
	TypeName    string
	PkgPath     string
	PkgName     string
	Fields      []*Field
	HasValidate bool
}

// stdSlots maps stdlib types to their configulator sentinel slots (G8/G9).
var stdSlots = map[string]string{
	"time.Duration":    "Duration",
	"net.IPNet":        "IPNet",
	"os.FileMode":      "FileMode",
	"time.Location":    "Location",
	"net.TCPAddr":      "TCPAddr",
	"net.UDPAddr":      "UDPAddr",
	"net.HardwareAddr": "HardwareAddr",
	"net/url.URL":      "URL",
	"time.Month":       "Month",
}

func buildModel(named *types.Named, outPkg *types.Package, noValidate bool) (*Model, error) {
	obj := named.Obj()
	if named.TypeParams() != nil && named.TypeParams().Len() > 0 {
		return nil, fmt.Errorf("%s: generic config types are not supported", obj.Name())
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%s: config type must be a struct", obj.Name())
	}
	m := &Model{TypeName: obj.Name(), PkgPath: obj.Pkg().Path(), PkgName: obj.Pkg().Name()}

	if !noValidate && !hasValidate(named) {
		return nil, fmt.Errorf("%s: no Validate() error method; add one or pass -no-validate", obj.Name())
	}
	m.HasValidate = hasValidate(named)

	fields, err := walkStruct(st, outPkg, obj.Name())
	if err != nil {
		return nil, err
	}
	m.Fields = fields
	if err := checkSiblingCollisions(m.Fields, obj.Name()); err != nil {
		return nil, err
	}
	return m, nil
}

func hasValidate(t types.Type) bool {
	for _, recv := range []types.Type{t, types.NewPointer(t)} {
		ms := types.NewMethodSet(recv)
		for i := 0; i < ms.Len(); i++ {
			f := ms.At(i).Obj().(*types.Func)
			if f.Name() != "Validate" {
				continue
			}
			sig := f.Type().(*types.Signature)
			if sig.Params().Len() == 0 && sig.Results().Len() == 1 &&
				sig.Results().At(0).Type().String() == "error" {
				return true
			}
		}
	}
	return false
}

func walkStruct(st *types.Struct, outPkg *types.Package, path string) ([]*Field, error) {
	var out []*Field
	for i := 0; i < st.NumFields(); i++ {
		fv := st.Field(i)
		tag := reflect.StructTag(st.Tag(i))

		if fv.Embedded() && tag.Get("name") == "" {
			// Flatten: promote the embedded struct's fields (Go promotion).
			emb, ok := fv.Type().Underlying().(*types.Struct)
			if !ok {
				return nil, fmt.Errorf("%s.%s: embedded non-struct is not supported", path, fv.Name())
			}
			sub, err := walkStruct(emb, outPkg, path+"."+fv.Name())
			if err != nil {
				return nil, err
			}
			for _, f := range sub {
				f.GoName = fv.Name() + "." + f.GoName // promotion path
				f.Embedded = true
			}
			out = append(out, sub...)
			continue
		}

		name := tagName(tag)
		if name == "" {
			if !fv.Exported() {
				continue // untagged unexported: skipped, as v1
			}
			continue // untagged exported: skipped, as v1
		}
		if !fv.Exported() {
			return nil, fmt.Errorf("%s.%s: unexported field carries a name: tag", path, fv.Name())
		}

		f := &Field{
			GoName:   fv.Name(),
			Tag:      name,
			Type:     fv.Type(),
			Default:  tag.Get("default"),
			Desc:     tag.Get("description"),
			Required: tag.Get("required") == "true",
			Secret:   tag.Get("secret") == "true",
			Short:    tag.Get("short"),
			Opaque:   tag.Get("opaque") == "true",
		}
		switch tag.Get("env") {
		case "-":
			f.EnvSkip = true
		case "":
		default:
			f.EnvName = tag.Get("env")
			if !isUpperEnvSeg(f.EnvName) {
				return nil, fmt.Errorf(
					"%s.%s: env:%q overrides must be uppercase A-Z, 0-9, and _ (they are used verbatim in both configulator implementations)",
					path, f.GoName, f.EnvName)
			}
		}
		switch tag.Get("flag") {
		case "-":
			f.FlagSkip = true
		case "":
		default:
			f.FlagName = tag.Get("flag")
		}

		if err := classify(f, outPkg, path); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

// envSeg is this field's env var segment: the tag name (folded at runtime
// by EnvName) or the verbatim env:"NAME" override, which validation pins
// to characters the fold cannot change.
func (f *Field) envSeg() string {
	if f.EnvName != "" {
		return f.EnvName
	}
	return f.Tag
}

// flagSeg is this field's flag name segment: the tag name or the verbatim
// flag:"name" override.
func (f *Field) flagSeg() string {
	if f.FlagName != "" {
		return f.FlagName
	}
	return f.Tag
}

func isUpperEnvSeg(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func tagName(tag reflect.StructTag) string {
	if n := tag.Get("name"); n != "" {
		return n
	}
	for _, k := range []string{"json", "yaml"} {
		if n := strings.SplitN(tag.Get(k), ",", 2)[0]; n != "" && n != "-" {
			return n
		}
	}
	return ""
}

func classify(f *Field, outPkg *types.Package, path string) error {
	t := f.Type
	fieldPath := path + "." + f.GoName

	// stdtypes sentinel slots take precedence (opaque-leaf detection runs
	// before mirroring; the slot set is fixed and format-independent).
	if named, ok := t.(*types.Named); ok {
		full := named.Obj().Pkg().Path() + "." + named.Obj().Name()
		if named.Obj().Pkg() == nil {
			full = named.Obj().Name()
		}
		if slot, ok := stdSlots[full]; ok {
			if full == "time.Duration" {
				f.Kind = KindDuration
			} else {
				f.Kind = KindStdSlot
			}
			f.SlotType = slot
			return nil
		}
		if full == "net/url.URL" {
			// unreachable (in stdSlots) but kept as the documented error site
			return fmt.Errorf("%s: url.URL is undecodable as a leaf; use the generated slot", fieldPath)
		}
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		info := u.Info()
		switch {
		case info&types.IsBoolean != 0:
			f.Kind = KindBool
		case info&types.IsUnsigned != 0:
			f.Kind = KindUint
			f.Bits = basicBits(u)
		case info&types.IsInteger != 0:
			f.Kind = KindInt
			f.Bits = basicBits(u)
		case info&types.IsFloat != 0:
			f.Kind = KindFloat
			f.Bits = basicBits(u)
		case info&types.IsString != 0:
			f.Kind = KindString
		default:
			return fmt.Errorf("%s: unsupported basic type %s", fieldPath, u)
		}
	case *types.Struct:
		// TextUnmarshaler struct-kind => sentinel slot required; without
		// one and without exported mirror-safe fields this is an error.
		if implementsTextUnmarshaler(t) {
			return fmt.Errorf("%s: struct-kind TextUnmarshaler %s needs a sentinel slot; no stdtypes slot exists — wrap it or mark opaque:\"true\" knowingly", fieldPath, t)
		}
		f.Kind = KindStruct
		sub, err := walkStruct(u, outPkg, fieldPath)
		if err != nil {
			return err
		}
		// Prefix the whole subtree's Go selector paths with this
		// field's name so emission renders cfg.DB.Pool.Size at any
		// depth. Descendants already carry their intermediate path, so
		// each node gets the parent prepended exactly once. Element and
		// shadow emission use only the last segment.
		prefixSubtree(sub, f.GoName)
		f.Fields = sub
		if err := checkSiblingCollisions(sub, fieldPath); err != nil {
			return err
		}
	case *types.Pointer:
		f.Kind = KindPointer
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem()}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		f.Elem = elem
	case *types.Slice:
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem()}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		f.Elem = elem
		if elem.Kind == KindStruct {
			f.Kind = KindSliceStruct
		} else {
			f.Kind = KindSliceScalar
		}
		if f.Kind == KindSliceStruct && (f.EnvName != "" || f.FlagName != "") {
			return fmt.Errorf("%s: env:/flag: opt-in on a list of structs (file-only, SPEC rule 6)", fieldPath)
		}
	case *types.Map:
		if b, ok := u.Key().Underlying().(*types.Basic); !ok || b.Info()&types.IsString == 0 {
			return fmt.Errorf("%s: map keys must be strings", fieldPath)
		}
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem()}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		f.Elem = elem
		if elem.Kind == KindStruct {
			f.Kind = KindMapStruct
		} else {
			f.Kind = KindMapScalar
		}
		if f.EnvName != "" || f.FlagName != "" {
			return fmt.Errorf("%s: env:/flag: opt-in on a map (file-only, SPEC rule 6)", fieldPath)
		}
	default:
		if implementsTextUnmarshaler(t) && !isStructKind(t) {
			f.Kind = KindTextLeaf // safe naked leaf (G8: non-struct kinds error loudly)
			return nil
		}
		return fmt.Errorf("%s: unsupported type %s", fieldPath, t)
	}

	if f.Default != "" {
		if err := checkDefault(f); err != nil {
			return fmt.Errorf("%s: default:%q: %v", fieldPath, f.Default, err)
		}
	}
	return nil
}

func basicBits(b *types.Basic) int {
	switch b.Kind() {
	case types.Int8, types.Uint8:
		return 8
	case types.Int16, types.Uint16:
		return 16
	case types.Int32, types.Uint32, types.Float32:
		return 32
	case types.Int64, types.Uint64, types.Float64:
		return 64
	}
	return 0
}

func isStructKind(t types.Type) bool {
	_, ok := t.Underlying().(*types.Struct)
	return ok
}

func implementsTextUnmarshaler(t types.Type) bool {
	// encoding.TextUnmarshaler: UnmarshalText([]byte) error, value or ptr recv
	for _, recv := range []types.Type{t, types.NewPointer(t)} {
		ms := types.NewMethodSet(recv)
		for i := 0; i < ms.Len(); i++ {
			fn := ms.At(i).Obj().(*types.Func)
			if fn.Name() != "UnmarshalText" {
				continue
			}
			sig := fn.Type().(*types.Signature)
			if sig.Params().Len() == 1 && sig.Results().Len() == 1 &&
				sig.Params().At(0).Type().String() == "[]byte" &&
				sig.Results().At(0).Type().String() == "error" {
				return true
			}
		}
	}
	return false
}

func checkDefault(f *Field) error {
	switch f.Kind {
	case KindString, KindSliceScalar:
		return nil
	case KindBool:
		_, err := strconv.ParseBool(f.Default)
		return err
	case KindInt:
		bits := f.Bits
		if bits == 0 {
			bits = 64
		}
		_, err := strconv.ParseInt(f.Default, 10, bits)
		return err
	case KindUint:
		bits := f.Bits
		if bits == 0 {
			bits = 64
		}
		_, err := strconv.ParseUint(f.Default, 10, bits)
		return err
	case KindFloat:
		_, err := strconv.ParseFloat(f.Default, 64)
		return err
	case KindDuration:
		_, err := time.ParseDuration(f.Default)
		return err
	case KindPointer:
		return checkDefault(&Field{Kind: f.Elem.Kind, Bits: f.Elem.Bits, Default: f.Default, Elem: f.Elem.Elem})
	case KindStdSlot, KindTextLeaf:
		return fmt.Errorf("default: on a custom TextUnmarshaler type is a generate-time error; set it in code")
	case KindMapScalar, KindMapStruct:
		return fmt.Errorf("default: on a map is not supported")
	case KindSliceStruct:
		return fmt.Errorf("default: on a list of structs is not supported")
	}
	return nil
}

func prefixSubtree(fields []*Field, parent string) {
	for _, sf := range fields {
		sf.GoName = parent + "." + sf.GoName
		prefixSubtree(sf.Fields, parent)
	}
}

// checkSiblingCollisions rejects same-struct siblings colliding under
// case- and -/_ folding (the env layer's fold, SPEC rule 7).
func checkSiblingCollisions(fields []*Field, path string) error {
	seen := map[string]string{}
	for _, f := range fields {
		folded := strings.ReplaceAll(strings.ToUpper(f.Tag), "-", "_")
		if prev, ok := seen[folded]; ok {
			return fmt.Errorf("%s: fields %q and %q collide under env-name folding (%s)", path, prev, f.Tag, folded)
		}
		seen[folded] = f.Tag
	}
	return nil
}

// findShortTag returns the first dotted path carrying a short: tag, for
// the -flags=std generate-time error.
func findShortTag(fields []*Field, prefix string) string {
	for _, f := range fields {
		p := joinPath(prefix, f.Tag)
		if f.Short != "" {
			return p
		}
		if f.Kind == KindStruct {
			if s := findShortTag(f.Fields, p); s != "" {
				return s
			}
		}
	}
	return ""
}
