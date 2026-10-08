//go:build goexperiment.jsonv2

package configulator

import "strings"

// EnvName builds an environment variable name: prefix followed by the
// uppercased segments joined by sep. "-" becomes "_" in the segments only;
// prefix and sep are used as-is.
func EnvName(prefix, sep string, segments ...string) string {
	folded := make([]string, len(segments))
	for i, s := range segments {
		folded[i] = strings.ReplaceAll(strings.ToUpper(s), "-", "_")
	}
	return prefix + strings.Join(folded, sep)
}

// SplitList splits s on sep without trimming. An empty s gives an empty
// list.
func SplitList(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
