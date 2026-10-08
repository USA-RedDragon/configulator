package impl

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

// LookupEnv builds a field's env var name like EnvName and looks it up with
// getenv. It returns the name even when the variable is not set.
func LookupEnv(getenv func(string) (string, bool), prefix, sep string, segments ...string) (name, value string, ok bool) {
	name = EnvName(prefix, sep, segments...)
	value, ok = getenv(name)
	return name, value, ok
}
