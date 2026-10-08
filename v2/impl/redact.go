package impl

import "strings"

// Redact hides raw, the input for a field tagged secret:"true", in err's
// message. It returns nil for a nil err.
func Redact(err error, raw string) error {
	if err == nil {
		return nil
	}
	return &redacted{err: err, raw: raw}
}

type redacted struct {
	err error
	raw string
}

func (r *redacted) Error() string {
	if r.raw == "" {
		return r.err.Error()
	}
	return strings.ReplaceAll(r.err.Error(), r.raw, "(redacted)")
}

func (r *redacted) Unwrap() error { return r.err }
