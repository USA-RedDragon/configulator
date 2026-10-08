package main

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// emitJSONSchema renders a draft-07 JSON Schema for the config type.
// additionalProperties is false because the generated decoder rejects
// unknown keys.
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
	if f.Default != "" && !f.Secret {
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
			out[i] = schemaDefault(&Field{Kind: f.Elem.Kind, Type: f.Elem.Type, Default: p})
		}
		return out
	case KindPointer:
		inner := *f.Elem
		inner.Default = f.Default
		return schemaDefault(&inner)
	default:
		return nil
	}
}

// emitSample renders a YAML sample with every key at its default and the
// descriptions as comments. It's YAML because YAML has comments, whichever
// decoder the application uses.
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
		case KindSliceStruct, KindMapStruct, KindMapScalar:
			for _, line := range exampleField(f) {
				fmt.Fprintf(b, "%s# %s\n", ind, line)
			}
		case KindPointer:
			if f.Elem.Kind == KindStruct {
				for _, line := range exampleField(f) {
					fmt.Fprintf(b, "%s# %s\n", ind, line)
				}
				continue
			}
			if f.Default != "" && !f.Secret {
				fmt.Fprintf(b, "%s%s: %s\n", ind, f.Tag, sampleValue(f))
				continue
			}
			fmt.Fprintf(b, "%s# %s: %s\n", ind, f.Tag, secretOr(f, sampleValue(f)))
		default:
			val := secretOr(f, sampleValue(f))
			if f.Secret || (f.Default == "" && f.Kind != KindBool) {
				fmt.Fprintf(b, "%s# %s: %s\n", ind, f.Tag, val)
			} else {
				fmt.Fprintf(b, "%s%s: %s\n", ind, f.Tag, val)
			}
		}
	}
}

// exampleField renders f as uncommented YAML lines showing one example
// element for every collection, at any depth.
func exampleField(f *Field) []string {
	switch f.Kind {
	case KindStruct:
		return append([]string{f.Tag + ":"}, indentLines(exampleFields(f.Fields), "  ")...)
	case KindPointer:
		if f.Elem.Kind == KindStruct {
			return append([]string{f.Tag + ":"}, indentLines(exampleFields(f.Elem.Fields), "  ")...)
		}
		return []string{f.Tag + ": " + secretOr(f, sampleValue(f))}
	case KindSliceStruct:
		item := exampleFields(f.Elem.Fields)
		if len(item) == 0 {
			return []string{f.Tag + ": []"}
		}
		lines := make([]string, 0, len(item)+1)
		lines = append(lines, f.Tag+":", "  - "+item[0])
		return append(lines, indentLines(item[1:], "    ")...)
	case KindMapStruct:
		elem := exampleFields(f.Elem.Fields)
		lines := make([]string, 0, len(elem)+2)
		lines = append(lines, f.Tag+":", "  example:")
		return append(lines, indentLines(elem, "    ")...)
	case KindMapScalar:
		return []string{f.Tag + ":", "  example: " + sampleValue(f.Elem)}
	default:
		return []string{f.Tag + ": " + secretOr(f, sampleValue(f))}
	}
}

// secretOr returns val, or a "(secret)" placeholder for a secret field.
func secretOr(f *Field, val string) string {
	if f.Secret {
		return `"(secret)"`
	}
	return val
}

func exampleFields(fields []*Field) []string {
	lines := make([]string, 0, len(fields))
	for _, f := range fields {
		lines = append(lines, exampleField(f)...)
	}
	return lines
}

func indentLines(lines []string, prefix string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = prefix + l
	}
	return out
}

func sampleValue(f *Field) string {
	if f.Kind == KindPointer {
		return sampleValue(&Field{Kind: f.Elem.Kind, Type: f.Elem.Type, Default: f.Default, Elem: f.Elem.Elem})
	}
	if f.Default != "" {
		switch f.Kind {
		case KindString, KindDuration, KindStdSlot, KindTextLeaf:
			return strconv.Quote(f.Default)
		case KindSliceScalar:
			parts := strings.Split(f.Default, ",")
			for i, p := range parts {
				parts[i] = sampleValue(&Field{Kind: f.Elem.Kind, Type: f.Elem.Type, Default: p})
			}
			return "[" + strings.Join(parts, ", ") + "]"
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
	default:
		return `""`
	}
}

// emitMarkdown renders a Markdown table of every config key with its file
// path, env var, flag, type, default and description. Collections are
// file-only, so their env and flag cells are em-dash characters.
func emitMarkdown(m *Model, flagSep, envPrefix, envSep string, title bool) []byte {
	var rows [][6]string
	markdownFields(&rows, m.Fields, "", envPrefix, envSep, "", flagSep, true, true)

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
	if title {
		fmt.Fprintf(&b, "# %s configuration\n\n", m.TypeName)
	}
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

func markdownFields(rows *[][6]string, fields []*Field, path, env, envSep, flagPath, flagSep string, envOK, flagOK bool) {
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
		if envOK && !f.EnvSkip {
			envCell = "`" + fEnv + "`"
		}
		if flagOK && !f.FlagSkip {
			flagCell = "`--" + fFlag + "`"
			if f.Short != "" {
				flagCell = "`-" + f.Short + "`, " + flagCell
			}
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
		if f.Default != "" && !f.Secret {
			def = "`" + f.Default + "`"
		}
		switch f.Kind {
		case KindStruct:
			markdownFields(rows, f.Fields, key, fEnv+envSep, envSep, fFlag, flagSep, envOK && !f.EnvSkip, flagOK && !f.FlagSkip)
		case KindSliceStruct:
			*rows = append(*rows, [6]string{"`" + key + "`", "list of objects", "", "\u2014", "\u2014", desc})
			markdownFields(rows, f.Elem.Fields, key+"[]", "", envSep, "", flagSep, false, false)
		case KindMapStruct:
			*rows = append(*rows, [6]string{"`" + key + "`", "map of objects", "", "\u2014", "\u2014", desc})
			markdownFields(rows, f.Elem.Fields, key+".<key>", "", envSep, "", flagSep, false, false)
		case KindMapScalar:
			*rows = append(*rows, [6]string{"`" + key + "`", "map of " + markdownType(f.Elem), "", "\u2014", "\u2014", desc})
		case KindSliceScalar:
			*rows = append(*rows, [6]string{"`" + key + "`", "list of " + markdownType(f.Elem), def, envCell, flagCell, desc})
		case KindPointer:
			inner := *f.Elem
			if inner.Kind == KindStruct {
				markdownFields(rows, inner.Fields, key, fEnv+envSep, envSep, fFlag, flagSep, envOK && !f.EnvSkip, flagOK && !f.FlagSkip)
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
	default:
		return "string"
	}
}

// envSegUpper returns the env segment the same way EnvName builds it at
// runtime: the env override as is, or the tag uppercased with - as _.
func envSegUpper(f *Field) string {
	if f.EnvName != "" {
		return f.EnvName
	}
	return strings.ReplaceAll(strings.ToUpper(f.Tag), "-", "_")
}
