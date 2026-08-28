//go:build goexperiment.jsonv2

package configulator

import "strings"

// EnvName constructs an environment variable name with PREFIX + upper(tag)
// with nested levels joined by sep, recursing with tag names. The "-" -> "_"
// fold applies ONLY to tag-derived segments; prefix and separator are used
// verbatim
func EnvName(prefix, sep string, segments ...string) string {
	folded := make([]string, len(segments))
	for i, s := range segments {
		folded[i] = strings.ReplaceAll(strings.ToUpper(s), "-", "_")
	}
	return prefix + strings.Join(folded, sep)
}

// SplitList splits a list-valued env or default string on the configured
// array separator. No trimming. An empty input is an empty list.
func SplitList(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
