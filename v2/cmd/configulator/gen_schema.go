package main

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// emitJSONSchema renders a JSON Schema (draft-07 subset) for the config
// type: types, descriptions, defaults, required, additionalProperties
// false (matching the generated decoder's unknown-key rejection).
func emitJSONSchema(m *Model) ([]byte, error) {
	root := schemaObject(m.Fields)
	root["$schema"] = "http://json-schema.org/draft-07/schema#"
	root["title"] = m.TypeName
	b, err := jsonv2.Marshal(root, jsonv2.Deterministic(true), jsontext.WithIndent("  "))
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func schemaObject(fields []*Field) map[string]any {
	props := map[string]any{}
	var required []string
	for _, f := range fields {
		props[f.Tag] = schemaField(f)
		if f.Required {
			required = append(required, f.Tag)
		}
	}
	obj := map[string]any{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		obj["required"] = required
	}
	return obj
}

func schemaField(f *Field) map[string]any {
	s := map[string]any{}
	if f.Desc != "" {
		s["description"] = f.Desc
	}
	switch f.Kind {
	case KindString, KindDuration, KindStdSlot, KindTextLeaf:
		s["type"] = "string"
	case KindBool:
		s["type"] = "boolean"
	case KindInt, KindUint:
		s["type"] = "integer"
	case KindFloat:
		s["type"] = "number"
	case KindStruct:
		for k, v := range schemaObject(f.Fields) {
			s[k] = v
		}
	case KindPointer:
		inner := *f.Elem
		inner.Desc = ""
		for k, v := range schemaField(&inner) {
			s[k] = v
		}
	case KindSliceScalar, KindSliceStruct:
		s["type"] = "array"
		s["items"] = schemaField(f.Elem)
	case KindMapScalar, KindMapStruct:
		s["type"] = "object"
		s["additionalProperties"] = schemaField(f.Elem)
	}
	if f.Default != "" {
		if d := schemaDefault(f); d != nil {
			s["default"] = d
		}
	}
	return s
}

func schemaDefault(f *Field) any {
	switch f.Kind {
	case KindString, KindDuration, KindStdSlot:
		return f.Default
	case KindBool:
		v, _ := strconv.ParseBool(f.Default)
		return v
	case KindInt:
		v, _ := strconv.ParseInt(f.Default, 10, 64)
		return v
	case KindUint:
		v, _ := strconv.ParseUint(f.Default, 10, 64)
		return v
	case KindFloat:
		v, _ := strconv.ParseFloat(f.Default, 64)
		return v
	case KindSliceScalar:
		parts := strings.Split(f.Default, ",")
		out := make([]any, len(parts))
		for i, p := range parts {
			out[i] = p
		}
		return out
	case KindPointer:
		inner := *f.Elem
		inner.Default = f.Default
		return schemaDefault(&inner)
	}
	return nil
}

// emitSample renders a commented YAML sample: every key at its default,
// descriptions as comments. YAML is chosen because it carries comments;
// the sample documents shape and defaults regardless of which decoder the
// application wires.
func emitSample(m *Model) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# Sample configuration for %s.\n", m.TypeName)
	sampleFields(&b, m.Fields, 0)
	return []byte(b.String())
}

func sampleFields(b *strings.Builder, fields []*Field, depth int) {
	ind := strings.Repeat("  ", depth)
	for _, f := range fields {
		if f.Desc != "" {
			fmt.Fprintf(b, "%s# %s\n", ind, f.Desc)
		}
		switch f.Kind {
		case KindStruct:
			fmt.Fprintf(b, "%s%s:\n", ind, f.Tag)
			sampleFields(b, f.Fields, depth+1)
		case KindSliceStruct:
			fmt.Fprintf(b, "%s# %s: []  # list of objects:\n", ind, f.Tag)
			fmt.Fprintf(b, "%s# %s:\n", ind, f.Tag)
			for _, ef := range f.Elem.Fields {
				fmt.Fprintf(b, "%s#   - %s: %s\n", ind, ef.Tag, sampleValue(ef))
				break
			}
		case KindMapStruct, KindMapScalar:
			fmt.Fprintf(b, "%s# %s: {}  # map\n", ind, f.Tag)
		default:
			val := sampleValue(f)
			if f.Secret {
				val = `"(secret)"`
			}
			if f.Default == "" && f.Kind != KindBool {
				fmt.Fprintf(b, "%s# %s: %s\n", ind, f.Tag, val)
			} else {
				fmt.Fprintf(b, "%s%s: %s\n", ind, f.Tag, val)
			}
		}
	}
}

func sampleValue(f *Field) string {
	if f.Default != "" {
		switch f.Kind {
		case KindString, KindDuration, KindStdSlot:
			return strconv.Quote(f.Default)
		case KindSliceScalar:
			return "[" + f.Default + "]"
		default:
			return f.Default
		}
	}
	switch f.Kind {
	case KindString, KindTextLeaf, KindStdSlot:
		return `""`
	case KindBool:
		return "false"
	case KindInt, KindUint:
		return "0"
	case KindFloat:
		return "0.0"
	case KindDuration:
		return strconv.Quote(time.Duration(0).String())
	case KindSliceScalar:
		return "[]"
	case KindPointer:
		return sampleValue(f.Elem)
	}
	return `""`
}

// emitMarkdown renders a flat Markdown reference table of every config
// key: file path, env var, flag, type, default, and description, with
// cells padded so the raw text reads as a table. Collections are
// file-only, so their env/flag cells are em-dashes; struct-collection
// element fields appear as servers[].addr / pools.<key>.size rows.
// Mirrors the Rust configulator-cli's --markdown output.
func emitMarkdown(m *Model, flagSep, envPrefix, envSep string) []byte {
	var rows [][6]string
	markdownFields(&rows, m.Fields, "", envPrefix, envSep, "", flagSep, true)

	header := [6]string{"Key", "Type", "Default", "Environment", "Flag", "Description"}
	var widths [6]int
	for i, h := range header {
		widths[i] = runeLen(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if l := runeLen(c); l > widths[i] {
				widths[i] = l
			}
		}
	}
	pad := func(c string, w int) string { return c + strings.Repeat(" ", w-runeLen(c)) }

	var b strings.Builder
	fmt.Fprintf(&b, "# %s configuration\n\n", m.TypeName)
	cells := make([]string, 6)
	for i, h := range header {
		cells[i] = pad(h, widths[i])
	}
	fmt.Fprintf(&b, "| %s |\n", strings.Join(cells, " | "))
	for i := range header {
		cells[i] = strings.Repeat("-", widths[i])
	}
	fmt.Fprintf(&b, "|-%s-|\n", strings.Join(cells, "-|-"))
	for _, r := range rows {
		for i, c := range r {
			cells[i] = pad(c, widths[i])
		}
		fmt.Fprintf(&b, "| %s |\n", strings.Join(cells, " | "))
	}
	return []byte(b.String())
}

func runeLen(s string) int { return len([]rune(s)) }

func markdownFields(rows *[][6]string, fields []*Field, path, env, envSep, flagPath, flagSep string, reachable bool) {
	for _, f := range fields {
		key := f.Tag
		if path != "" {
			key = path + "." + f.Tag
		}
		fEnv := env + envSegUpper(f)
		fFlag := f.flagSeg()
		if flagPath != "" {
			fFlag = flagPath + flagSep + f.flagSeg()
		}
		envCell, flagCell := "\u2014", "\u2014"
		if reachable && !f.EnvSkip {
			envCell = "`" + fEnv + "`"
		}
		if reachable && !f.FlagSkip {
			flagCell = "`--" + fFlag + "`"
		}
		desc := f.Desc
		if f.Required {
			if desc == "" {
				desc = "required"
			} else {
				desc += " (required)"
			}
		}
		if f.Secret {
			if desc == "" {
				desc = "secret"
			} else {
				desc += " (secret)"
			}
		}
		def := ""
		if f.Default != "" {
			def = "`" + f.Default + "`"
		}
		switch f.Kind {
		case KindStruct:
			markdownFields(rows, f.Fields, key, fEnv+envSep, envSep, fFlag, flagSep, reachable)
		case KindSliceStruct:
			*rows = append(*rows, [6]string{"`" + key + "`", "list of objects", "", "\u2014", "\u2014", desc})
			markdownFields(rows, f.Elem.Fields, key+"[]", "", envSep, "", flagSep, false)
		case KindMapStruct:
			*rows = append(*rows, [6]string{"`" + key + "`", "map of objects", "", "\u2014", "\u2014", desc})
			markdownFields(rows, f.Elem.Fields, key+".<key>", "", envSep, "", flagSep, false)
		case KindMapScalar:
			*rows = append(*rows, [6]string{"`" + key + "`", "map of " + markdownType(f.Elem), "", "\u2014", "\u2014", desc})
		case KindSliceScalar:
			*rows = append(*rows, [6]string{"`" + key + "`", "list of " + markdownType(f.Elem), def, envCell, flagCell, desc})
		case KindPointer:
			inner := *f.Elem
			if inner.Kind == KindStruct {
				markdownFields(rows, inner.Fields, key, fEnv+envSep, envSep, fFlag, flagSep, reachable)
				continue
			}
			*rows = append(*rows, [6]string{"`" + key + "`", markdownType(&inner), def, envCell, flagCell, desc})
		default:
			*rows = append(*rows, [6]string{"`" + key + "`", markdownType(f), def, envCell, flagCell, desc})
		}
	}
}

func markdownType(f *Field) string {
	switch f.Kind {
	case KindString, KindDuration, KindStdSlot, KindTextLeaf:
		return "string"
	case KindBool:
		return "boolean"
	case KindInt, KindUint:
		return "integer"
	case KindFloat:
		return "number"
	}
	return "string"
}

// envSegUpper is the env segment as EnvName folds it at runtime: the
// verbatim override, or the tag uppercased with - folded to _.
func envSegUpper(f *Field) string {
	if f.EnvName != "" {
		return f.EnvName
	}
	return strings.ReplaceAll(strings.ToUpper(f.Tag), "-", "_")
}
