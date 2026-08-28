// Package seam is the module-internal channel between the core and the
// flag adapter packages. It adds ZERO public API surface to the core:
// the core registers an adapter view here at construction, and an
// adapter's Bind pops it. Nothing outside this module can import this
// package, so nothing outside can observe the machinery at all.
package seam

import "sync"

// Flag is the adapter view. Type parameters keep this package import-free
// (no cycle with the core): C is the config type, FO the core's
// *FileOptions, SO the core's SetOrigin.
type Flag[C any, FO any, SO any] struct {
	// FileOptions returns the file options configured at the time Bind
	// runs (nil without WithFile) — a getter because WithFile may be
	// chained after New.
	FileOptions func() FO
	// Install stores the type-erased hooks the Load pipeline runs.
	Install func(apply func(*C, SO) error, configPath func() (string, bool), regErr error)
}

var (
	mu    sync.Mutex
	views = map[any]any{}
)

// Register associates an adapter view with a Configulator. Called by the
// core's New.
func Register(key, view any) {
	mu.Lock()
	defer mu.Unlock()
	views[key] = view
}

// Take returns and removes the view for a Configulator. Called by an
// adapter's Bind; popping makes a second Bind on the same Configulator a
// loud error rather than a silent overwrite, and keeps the registry from
// retaining bound Configulators.
func Take(key any) (any, bool) {
	mu.Lock()
	defer mu.Unlock()
	v, ok := views[key]
	if ok {
		delete(views, key)
	}
	return v, ok
}
