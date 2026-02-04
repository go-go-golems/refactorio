package modules

import (
	"log"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

// NativeModule is the interface every JS-exposed module must satisfy.
type NativeModule interface {
	Name() string
	Doc() string
	Loader(*goja.Runtime, *goja.Object)
}

// Registry manages a collection of native modules for a goja runtime.
type Registry struct {
	modules []NativeModule
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{
		modules: []NativeModule{},
	}
}

// Register adds a module to the registry. Typically called from init().
func (r *Registry) Register(m NativeModule) {
	r.modules = append(r.modules, m)
}

// Enable registers all modules from this registry with a goja require.Registry.
func (r *Registry) Enable(gojaRegistry *require.Registry) {
	for _, m := range r.modules {
		log.Printf("refactorio js: registering native module %s", m.Name())
		gojaRegistry.RegisterNativeModule(m.Name(), m.Loader)
	}
}

// DefaultRegistry is the default global module registry.
var DefaultRegistry = NewRegistry()

// Register adds a module implementation to the default registry.
func Register(m NativeModule) {
	DefaultRegistry.Register(m)
}

// EnableAll registers all modules from the default registry with a goja require.Registry.
func EnableAll(reg *require.Registry) {
	DefaultRegistry.Enable(reg)
}
