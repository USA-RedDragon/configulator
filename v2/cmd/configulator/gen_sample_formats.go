package main

import (
	"bytes"
	"encoding/json/jsontext"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// kv is an order-preserving sample tree node. Values are string, int64,
// uint64, float64, bool, []any (scalar list), or []kv (nested).
type kv struct {
	k string
	v any
}

func sampleLeafTyped(f *Field) any {
	if f.Default != "" {
		if d := schemaDefault(f); d != nil {
			return d
		}
	}
	switch f.Kind {
	case KindBool:
		return false
	case KindInt:
		return int64(0)
	case KindUint:
		return uint64(0)
	case KindFloat:
		return float64(0)
	case KindDuration:
		return "0s"
	default:
		return ""
	}
}

func sampleTree(fields []*Field) []kv {
	out := make([]kv, 0, len(fields))
	for _, f := range fields {
		if f.Secret {
			continue
		}
		var v any
		switch f.Kind {
		case KindStruct:
			v = sampleTree(f.Fields)
		case KindPointer:
			if f.Elem.Kind == KindStruct {
				v = sampleTree(f.Elem.Fields)
			} else {
				inner := *f.Elem
				inner.Tag, inner.Default, inner.Secret = f.Tag, f.Default, f.Secret
				v = sampleLeafTyped(&inner)
			}
		case KindSliceStruct:
			v = []any{}
		case KindMapStruct, KindMapScalar:
			v = []kv{}
		case KindSliceScalar:
			items := []any{}
			if f.Default != "" && !f.Secret {
				for _, part := range strings.Split(f.Default, ",") {
					elem := *f.Elem
					elem.Default = part
					items = append(items, sampleLeafTyped(&elem))
				}
			}
			v = items
		default:
			v = sampleLeafTyped(f)
		}
		out = append(out, kv{f.Tag, v})
	}
	return out
}

func emitSampleJSON(m *Model) ([]byte, error) {
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf, jsontext.WithIndent("  "))
	if err := writeJSONValue(enc, sampleTree(m.Fields)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeJSONValue(enc *jsontext.Encoder, v any) error {
	switch v := v.(type) {
	case []kv:
		if err := enc.WriteToken(jsontext.BeginObject); err != nil {
			return err
		}
		for _, pair := range v {
			if err := enc.WriteToken(jsontext.String(pair.k)); err != nil {
				return err
			}
			if err := writeJSONValue(enc, pair.v); err != nil {
				return err
			}
		}
		return enc.WriteToken(jsontext.EndObject)
	case []any:
		if err := enc.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
		for _, item := range v {
			if err := writeJSONValue(enc, item); err != nil {
				return err
			}
		}
		return enc.WriteToken(jsontext.EndArray)
	case string:
		return enc.WriteToken(jsontext.String(v))
	case bool:
		return enc.WriteToken(jsontext.Bool(v))
	case int64:
		return enc.WriteToken(jsontext.Int(v))
	case uint64:
		return enc.WriteToken(jsontext.Uint(v))
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return fmt.Errorf("sample value %v has no JSON spelling", v)
		}
		return enc.WriteValue(jsontext.Value(formatFloat(v)))
	}
	return fmt.Errorf("unhandled sample value %T", v)
}

func emitSampleTOML(m *Model) []byte {
	var b strings.Builder
	writeTOMLTable(&b, sampleTree(m.Fields), nil)
	return []byte(b.String())
}

// writeTOMLTable writes scalar keys first, then [table] sections. TOML
// requires that order, and it matches the Rust toml crate's output.
func writeTOMLTable(b *strings.Builder, pairs []kv, path []string) {
	type table struct {
		k     string
		pairs []kv
	}
	var tables []table
	for _, pair := range pairs {
		switch v := pair.v.(type) {
		case []kv:
			tables = append(tables, table{pair.k, v})
		case []any:
			parts := make([]string, len(v))
			for i, item := range v {
				parts[i] = tomlScalar(item)
			}
			fmt.Fprintf(b, "%s = [%s]\n", pair.k, strings.Join(parts, ", "))
		default:
			fmt.Fprintf(b, "%s = %s\n", pair.k, tomlScalar(pair.v))
		}
	}
	for _, t := range tables {
		full := append(append([]string{}, path...), t.k)
		fmt.Fprintf(b, "\n[%s]\n", strings.Join(full, "."))
		writeTOMLTable(b, t.pairs, full)
	}
}

func tomlScalar(v any) string {
	switch v := v.(type) {
	case string:
		return strconv.Quote(v)
	case bool:
		return strconv.FormatBool(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		switch {
		case math.IsNaN(v):
			return "nan"
		case math.IsInf(v, 1):
			return "inf"
		case math.IsInf(v, -1):
			return "-inf"
		}
		return formatFloat(v)
	}
	return `""`
}

// formatFloat spells v so TOML reads it as a float, not an integer: 1000
// becomes "1000.0". JSON samples use it too so both match configulator-rs.
func formatFloat(v float64) string {
	if a := math.Abs(v); a != 0 && (a < 1e-5 || a >= 1e16) {
		return strings.Replace(strconv.FormatFloat(v, 'e', -1, 64), "e+", "e", 1)
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
