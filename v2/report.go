//go:build goexperiment.jsonv2

package configulator

import "sort"

// Layer identifies which configuration source supplied a value.
type Layer uint8

const (
	// LayerDefault is a `default:` tag, including element defaults.
	LayerDefault Layer = iota
	// LayerFile is the config file.
	LayerFile
	// LayerEnv is an environment variable.
	LayerEnv
	// LayerCLI is a command-line flag.
	LayerCLI
)

func (l Layer) String() string {
	switch l {
	case LayerDefault:
		return "default"
	case LayerFile:
		return "file"
	case LayerEnv:
		return "env"
	case LayerCLI:
		return "cli"
	}
	return "unknown"
}

// Origin records where a single field's value came from. Detail is the file
// path, environment variable name, or flag name
type Origin struct {
	Layer  Layer
	Detail string
}

// FilePathOrigin says how the loaded config file's path was chosen.
type FilePathOrigin uint8

const (
	// PathNone means no file was loaded.
	PathNone FilePathOrigin = iota
	// PathSearch means a configured search path matched.
	PathSearch
	// PathConfigFlag means the operator passed --config.
	PathConfigFlag
	// PathExplicit means FileOptions.Explicit was set.
	PathExplicit
)

// FileResult reports which config file, if any, was loaded and how.
type FileResult struct {
	// Path is the file that loaded, or "" if none did.
	Path string
	// Origin says how Path was chosen.
	Origin FilePathOrigin
	// Searched lists every path tried, in order.
	Searched []string
}

// Report maps dotted field paths to the layer that supplied each value.
// Paths use tag names joined by ".", list indices as "[i]", and map keys
// quoted when they contain "." or "[". Fields never set by any layer and
// lacking a default have no entry.
type Report struct {
	origins map[string]Origin
	file    FileResult
}

// newReport returns an empty report. Generated code never constructs one —
// it receives *Report and calls Set; user code gets one from
// Configulator.Report.
func newReport() *Report {
	return &Report{origins: make(map[string]Origin)}
}

// set is the write side, deliberately unexported: generated appliers
// receive a SetOrigin closure over this method, so nothing outside the
// package can forge an origin. The read side (Origin, Paths, File) is the
// public API.
func (r *Report) set(path string, layer Layer, detail string) {
	r.origins[path] = Origin{Layer: layer, Detail: detail}
}

// Origin returns the recorded origin for a dotted path.
func (r *Report) Origin(path string) (Origin, bool) {
	if r == nil {
		return Origin{}, false
	}
	o, ok := r.origins[path]
	return o, ok
}

// Paths returns every recorded path, sorted.
func (r *Report) Paths() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.origins))
	for p := range r.origins {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// File reports which config file was loaded.
func (r *Report) File() FileResult {
	if r == nil {
		return FileResult{}
	}
	return r.file
}

func (r *Report) setFile(f FileResult) {
	if r != nil {
		r.file = f
	}
}
