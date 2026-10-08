package main

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/USA-RedDragon/configulator/v2/impl"
)

// Kind classifies a field for emission.
type Kind int

const (
	KindString Kind = iota
	KindBool
	KindInt
	KindUint
	KindFloat
	KindStruct
	KindPointer // pointer to scalar or struct, for optional fields (spec rule 5)
	KindSliceScalar
	KindSliceStruct // slice of structs (file only, spec rule 6)
	KindMapScalar   // map[string]scalar (file only)
	KindMapStruct   // map[string]struct (file only)
	KindDuration    // time.Duration -> impl.Duration slot
	KindStdSlot     // other sentinel slots, including impl.Text for any TextUnmarshaler
)

// Field describes one config field for the generator.
type Field struct {
	GoName   string
	Tag      string // configulator name (name: tag, json/yaml fallback)
	Kind     Kind
	Type     types.Type
	Bits     int    // int/uint/float width (0 = platform int)
	Default  string // raw default: tag ("" = none)
	Desc     string
	EnvSkip  bool
	FlagSkip bool
	EnvName  string // env:"NAME" explicit opt-in (error on collections)
	FlagName string
	Required bool
	Secret   bool
	Short    string
	Opaque   bool
	SlotType string   // configulator slot type name for KindDuration/KindStdSlot
	Fields   []*Field // struct / element fields
	Elem     *Field   // element for slices/maps/pointers
	Embedded bool     // promoted from an embedded struct
}

// Model describes one config type for the generator.
type Model struct {
	TypeName    string
	PkgPath     string
	PkgName     string
	Fields      []*Field
	HasValidate bool
}

// stdSlot returns the configulator sentinel slot for the stdlib type full,
// written as import path, dot, type name.
func stdSlot(full string) (string, bool) {
	switch full {
	case "time.Duration":
		return "Duration", true
	case "net.IPNet":
		return "IPNet", true
	case "os.FileMode", "io/fs.FileMode":
		return "FileMode", true
	case "time.Location":
		return "Location", true
	case "net.TCPAddr":
		return "TCPAddr", true
	case "net.UDPAddr":
		return "UDPAddr", true
	case "net.HardwareAddr":
		return "HardwareAddr", true
	case "net/url.URL":
		return "URL", true
	case "time.Month":
		return "Month", true
	default:
		return "", false
	}
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
			f, ok := ms.At(i).Obj().(*types.Func)
			if !ok || f.Name() != "Validate" {
				continue
			}
			sig, ok := f.Type().(*types.Signature)
			if ok && sig.Params().Len() == 0 && sig.Results().Len() == 1 &&
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
			emb, ok := fv.Type().Underlying().(*types.Struct)
			if !ok {
				return nil, fmt.Errorf("%s.%s: embedded non-struct is not supported", path, fv.Name())
			}
			sub, err := walkStruct(emb, outPkg, path+"."+fv.Name())
			if err != nil {
				return nil, err
			}
			for _, f := range sub {
				f.GoName = fv.Name() + "." + f.GoName
				f.Embedded = true
			}
			out = append(out, sub...)
			continue
		}

		name := tagName(tag)
		if name == "" {
			if !fv.Exported() {
				continue
			}
			continue
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

// envSeg returns the field's env var segment: the tag name (folded at
// runtime by EnvName) or the env:"NAME" override as written. Validation
// limits the override to characters the fold doesn't change.
func (f *Field) envSeg() string {
	if f.EnvName != "" {
		return f.EnvName
	}
	return f.Tag
}

// flagSeg returns the field's flag name segment: the tag name or the
// flag:"name" override as written.
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
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
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
	t := types.Unalias(f.Type)
	f.Type = t
	fieldPath := path + "." + f.GoName

	if handled, err := classifySpecial(f, t, fieldPath); handled {
		return err
	}

	// Check sentinel slots first, so a stdlib struct like net.IPNet isn't
	// mirrored field by field.
	if named, ok := t.(*types.Named); ok {
		if handled, err := classifySlot(f, named, fieldPath); handled {
			return err
		}
	}

	if namedText(f, t) {
		return classifyText(f, fieldPath)
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		if err := classifyBasic(f, u, fieldPath); err != nil {
			return err
		}
	case *types.Struct:
		if err := classifyStruct(f, t, u, outPkg, fieldPath); err != nil || f.Kind == KindStdSlot {
			return err
		}
	case *types.Pointer:
		f.Kind = KindPointer
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem(), Opaque: f.Opaque}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		switch elem.Kind {
		case KindString, KindBool, KindInt, KindUint, KindFloat, KindStruct, KindDuration, KindStdSlot:
		default:
			return fmt.Errorf("%s: pointer to %s is not supported", fieldPath, types.TypeString(u.Elem(), nil))
		}
		f.Elem = elem
	case *types.Slice:
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem(), Opaque: f.Opaque}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		f.Elem = elem
		switch {
		case elem.Kind == KindStruct:
			f.Kind = KindSliceStruct
		case listElemOK(elem):
			f.Kind = KindSliceScalar
		default:
			return fmt.Errorf("%s: list of %s is not supported", fieldPath, types.TypeString(u.Elem(), nil))
		}
		if f.Kind == KindSliceStruct && (f.EnvName != "" || f.FlagName != "") {
			return fmt.Errorf("%s: env:/flag: opt-in on a list of structs (file-only, SPEC rule 6)", fieldPath)
		}
	case *types.Map:
		if b, ok := u.Key().Underlying().(*types.Basic); !ok || b.Info()&types.IsString == 0 {
			return fmt.Errorf("%s: map keys must be strings", fieldPath)
		}
		elem := &Field{GoName: f.GoName, Tag: f.Tag, Type: u.Elem(), Opaque: f.Opaque}
		if err := classify(elem, outPkg, path); err != nil {
			return err
		}
		f.Elem = elem
		switch {
		case elem.Kind == KindStruct:
			f.Kind = KindMapStruct
		case listElemOK(elem):
			f.Kind = KindMapScalar
		default:
			return fmt.Errorf("%s: map of %s is not supported", fieldPath, types.TypeString(u.Elem(), nil))
		}
		if f.EnvName != "" || f.FlagName != "" {
			return fmt.Errorf("%s: env:/flag: opt-in on a map (file-only, SPEC rule 6)", fieldPath)
		}
	default:
		if implementsTextUnmarshaler(t) {
			return classifyText(f, fieldPath)
		}
		return fmt.Errorf("%s: unsupported type %s", fieldPath, t)
	}

	return checkFieldDefault(f, fieldPath)
}

// classifyStruct classifies a struct-kind field as a nested struct, or as a
// leaf decoded with its UnmarshalText when it has one.
func classifyStruct(f *Field, t types.Type, u *types.Struct, outPkg *types.Package, fieldPath string) error {
	if implementsTextUnmarshaler(t) {
		return classifyText(f, fieldPath)
	}
	if f.Opaque {
		return fmt.Errorf("%s: opaque:\"true\" needs a type with an UnmarshalText method", fieldPath)
	}
	f.Kind = KindStruct
	sub, err := walkStruct(u, outPkg, fieldPath)
	if err != nil {
		return err
	}
	// Each level prepends its own name once, giving full paths like
	// DB.Pool.Size.
	prefixSubtree(sub, f.GoName)
	f.Fields = sub
	return checkSiblingCollisions(sub, fieldPath)
}

// untaggedFields returns the paths of exported fields under named that
// have no name, json or yaml tag. The generator skips them.
func untaggedFields(named *types.Named) []string {
	var out []string
	seen := map[types.Type]bool{}
	var walk func(st *types.Struct, path string)
	walk = func(st *types.Struct, path string) {
		for i := 0; i < st.NumFields(); i++ {
			fv := st.Field(i)
			tag := reflect.StructTag(st.Tag(i))
			if fv.Embedded() && tag.Get("name") == "" {
				if emb, ok := fv.Type().Underlying().(*types.Struct); ok {
					walk(emb, path)
				}
				continue
			}
			if tagName(tag) == "" {
				if fv.Exported() {
					out = append(out, path+"."+fv.Name())
				}
				continue
			}
			if sub, ok := nestedStruct(fv.Type()); ok && !seen[fv.Type()] {
				seen[fv.Type()] = true
				walk(sub, path+"."+fv.Name())
			}
		}
	}
	if st, ok := named.Underlying().(*types.Struct); ok {
		walk(st, named.Obj().Name())
	}
	return out
}

// nestedStruct returns the struct a config field nests: t itself, or the
// element of a pointer, slice or map. Types decoded from text don't count.
func nestedStruct(t types.Type) (*types.Struct, bool) {
	switch u := types.Unalias(t).Underlying().(type) {
	case *types.Pointer:
		t = u.Elem()
	case *types.Slice:
		t = u.Elem()
	case *types.Map:
		t = u.Elem()
	}
	if implementsTextUnmarshaler(t) {
		return nil, false
	}
	if n, ok := types.Unalias(t).(*types.Named); ok && isStdSlot(n) {
		return nil, false
	}
	st, ok := types.Unalias(t).Underlying().(*types.Struct)
	return st, ok
}

// namedText reports whether t is a named slice or map with UnmarshalText,
// like net.IP, which is one text value. A named scalar type with
// UnmarshalText uses it only when tagged opaque:"true".
func namedText(f *Field, t types.Type) bool {
	if _, named := t.(*types.Named); !named || !implementsTextUnmarshaler(t) {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Slice, *types.Map:
		return true
	case *types.Basic:
		return f.Opaque
	}
	return false
}

// classifyText makes f a leaf decoded with its type's UnmarshalText.
func classifyText(f *Field, fieldPath string) error {
	f.Kind = KindStdSlot
	f.SlotType = slotText
	return checkFieldDefault(f, fieldPath)
}

// classifySlot classifies named as a sentinel slot. handled reports whether
// named is a stdlib type with a slot, in which case err is the result.
func classifySlot(f *Field, named *types.Named, fieldPath string) (handled bool, err error) {
	full := named.Obj().Name()
	if pkg := named.Obj().Pkg(); pkg != nil {
		full = pkg.Path() + "." + full
	}
	if slot, ok := stdSlot(full); ok {
		if full == "time.Duration" {
			f.Kind = KindDuration
		} else {
			f.Kind = KindStdSlot
		}
		f.SlotType = slot
		return true, checkFieldDefault(f, fieldPath)
	}
	if full == "net/url.URL" {
		// Unreachable while url.URL has a slot, kept for the error message.
		return true, fmt.Errorf("%s: url.URL is undecodable as a leaf; use the generated slot", fieldPath)
	}
	return false, nil
}

func classifyBasic(f *Field, u *types.Basic, fieldPath string) error {
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
	case info&types.IsComplex != 0:
		f.Kind = KindStdSlot
		f.SlotType = "Complex128"
		if u.Kind() == types.Complex64 {
			f.SlotType = "Complex64"
		}
	default:
		return fmt.Errorf("%s: unsupported basic type %s", fieldPath, u)
	}
	return nil
}

func checkFieldDefault(f *Field, fieldPath string) error {
	if f.Default == "" {
		return nil
	}
	if err := checkDefault(f); err != nil {
		return fmt.Errorf("%s: default:%q: %w", fieldPath, f.Default, err)
	}
	if f.Kind == KindBool || (f.Kind == KindPointer && f.Elem.Kind == KindBool) {
		b, _ := strconv.ParseBool(f.Default)
		f.Default = strconv.FormatBool(b)
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
	default:
		return 0
	}
}

func implementsTextUnmarshaler(t types.Type) bool {
	for _, recv := range []types.Type{t, types.NewPointer(t)} {
		ms := types.NewMethodSet(recv)
		for i := 0; i < ms.Len(); i++ {
			fn, ok := ms.At(i).Obj().(*types.Func)
			if !ok || fn.Name() != "UnmarshalText" {
				continue
			}
			sig, ok := fn.Type().(*types.Signature)
			if ok && sig.Params().Len() == 1 && sig.Results().Len() == 1 &&
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
		bits := f.Bits
		if bits == 0 {
			bits = 64
		}
		_, err := strconv.ParseFloat(f.Default, bits)
		return err
	case KindDuration:
		_, err := time.ParseDuration(f.Default)
		return err
	case KindPointer:
		return checkDefault(&Field{Kind: f.Elem.Kind, Bits: f.Elem.Bits, SlotType: f.Elem.SlotType, Default: f.Default, Elem: f.Elem.Elem})
	case KindStdSlot:
		if f.SlotType == slotText {
			return fmt.Errorf("default: on a type decoded with its own UnmarshalText is not supported; set it in code")
		}
		return parseStdSlot(f.SlotType, []byte(f.Default))
	case KindMapScalar, KindMapStruct:
		return fmt.Errorf("default: on a map is not supported")
	case KindSliceStruct:
		return fmt.Errorf("default: on a list of structs is not supported")
	case KindStruct:
		return fmt.Errorf("default: on a struct is not supported; set defaults on its fields")
	default:
		return nil
	}
}

func prefixSubtree(fields []*Field, parent string) {
	for _, sf := range fields {
		sf.GoName = parent + "." + sf.GoName
		prefixSubtree(sf.Fields, parent)
	}
}

// checkSiblingCollisions rejects sibling fields whose tags collide once
// folded for env names: uppercased, with - turned into _ (spec rule 7).
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

// findShortTag returns the first dotted path with a short: tag. -flags=std
// can't support shorthands, so main reports it as an error.
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

func isNamed(t types.Type, pkgPath, name string) bool {
	n, ok := types.Unalias(t).(*types.Named)
	return ok && n.Obj().Pkg() != nil && n.Obj().Pkg().Path() == pkgPath && n.Obj().Name() == name
}

// slotText is the SlotType of an opaque:"true" field, decoded by impl.Text.
const slotText = "Text"

// parseStdSlot validates a default: value for slot with the same code the
// generated loader runs.
func parseStdSlot(slot string, text []byte) error {
	switch slot {
	case "IPNet":
		return new(impl.IPNet).UnmarshalText(text)
	case "FileMode":
		return new(impl.FileMode).UnmarshalText(text)
	case "Location":
		return new(impl.Location).UnmarshalText(text)
	case "TCPAddr":
		return new(impl.TCPAddr).UnmarshalText(text)
	case "UDPAddr":
		return new(impl.UDPAddr).UnmarshalText(text)
	case "HardwareAddr":
		return new(impl.HardwareAddr).UnmarshalText(text)
	case "URL":
		return new(impl.URL).UnmarshalText(text)
	case "Month":
		return new(impl.Month).UnmarshalText(text)
	case "Complex64":
		return new(impl.Complex64).UnmarshalText(text)
	case "Complex128":
		return new(impl.Complex128).UnmarshalText(text)
	default:
		return fmt.Errorf("no default parser for slot %s", slot)
	}
}

// classifySpecial handles types that classify's general rules get wrong.
func classifySpecial(f *Field, t types.Type, fieldPath string) (bool, error) {
	if p, ok := t.(*types.Pointer); ok && isNamed(p.Elem(), "time", "Location") {
		f.Kind = KindStdSlot
		f.SlotType = "Location"
		return true, checkFieldDefault(f, fieldPath)
	}
	if isNamed(t, "time", "Location") {
		return true, fmt.Errorf("%s: use *time.Location, not time.Location", fieldPath)
	}
	if _, ok := t.(*types.Struct); ok {
		return true, fmt.Errorf("%s: anonymous struct types are not supported; declare a named type", fieldPath)
	}
	return false, nil
}
