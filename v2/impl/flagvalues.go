package impl

import (
	"flag"
	"fmt"
	"strconv"
)

// Int is a pflag.Value for an int flag. pflag's own int flag parses 64
// bits and wraps a bigger number on 32-bit platforms. Int returns a range
// error instead. Its Type is "int", so FlagSet.GetInt reads it.
type Int int

// NewInt returns an Int holding v.
func NewInt(v int) *Int { return (*Int)(&v) }

// Set parses s like pflag's int flag does.
func (i *Int) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = Int(v)
	return nil
}

// String returns the value in decimal.
func (i *Int) String() string { return strconv.Itoa(int(*i)) }

// Type returns "int".
func (*Int) Type() string { return "int" }

// Uint is Int for uint flags. Its Type is "uint", so FlagSet.GetUint
// reads it.
type Uint uint

// NewUint returns a Uint holding v.
func NewUint(v uint) *Uint { return (*Uint)(&v) }

// Set parses s like pflag's uint flag does.
func (u *Uint) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*u = Uint(v)
	return nil
}

// String returns the value in decimal.
func (u *Uint) String() string { return strconv.FormatUint(uint64(*u), 10) }

// Type returns "uint".
func (*Uint) Type() string { return "uint" }

// FlagValue returns the value of the flag name on fs as a T. It panics if fs
// has no such flag or the flag holds another type, which means the flag was
// replaced after Bind.
func FlagValue[T any](fs *flag.FlagSet, name string) T {
	if f := fs.Lookup(name); f != nil {
		if g, ok := f.Value.(flag.Getter); ok {
			if v, ok := g.Get().(T); ok {
				return v
			}
		}
	}
	panic(fmt.Sprintf("configulator: flag %q is missing or does not hold a %T", name, *new(T)))
}
