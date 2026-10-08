// Package seam passes hooks between the core package and the flag adapters
// without adding anything to the core's public API. New registers them here
// and the adapter's Bind takes them.
package seam

import "sync"

// Flag holds the hooks a flag adapter uses. The type parameters avoid an
// import cycle with the core: C is the config type, FO is *FileOptions and
// SO is SetOrigin.
type Flag[C any, FO any, SO any] struct {
	// FileOptions is a func because WithFile can be called after New.
	// It returns nil without WithFile.
	FileOptions func() FO
	// Install stores the adapter's hooks for Load to run.
	Install func(apply func(*C, SO) error, configPath func() (string, bool), regErr error)
}

var (
	mu    sync.Mutex
	views = map[any]any{}
)

// Register stores view for the Configulator key.
func Register(key, view any) {
	mu.Lock()
	defer mu.Unlock()
	views[key] = view
}

// Take returns and removes the view for key. Removing it makes a second
// Bind on the same Configulator fail and lets bound Configulators be
// garbage collected.
func Take(key any) (any, bool) {
	mu.Lock()
	defer mu.Unlock()
	v, ok := views[key]
	if ok {
		delete(views, key)
	}
	return v, ok
}
