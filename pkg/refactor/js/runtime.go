package js

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/pkg/errors"
)

// ModuleSpec describes a native module that can be exposed to a goja runtime.
// Only modules provided in RuntimeOptions will be registered.
type ModuleSpec struct {
	Name   string
	Doc    string
	Loader func(*goja.Runtime, *goja.Object)
}

// RuntimeOptions configures the goja runtime creation.
type RuntimeOptions struct {
	Modules       []ModuleSpec
	EnableConsole bool
}

// NewRuntime creates a goja runtime with Node-style require() and a strict allow-list of modules.
func NewRuntime(opts RuntimeOptions) (*goja.Runtime, *require.RequireModule, error) {
	vm := goja.New()
	reg := require.NewRegistry()

	for _, mod := range opts.Modules {
		if mod.Name == "" {
			return nil, nil, errors.New("js runtime: module name is required")
		}
		if mod.Loader == nil {
			return nil, nil, errors.Errorf("js runtime: module loader missing for %s", mod.Name)
		}
		reg.RegisterNativeModule(mod.Name, mod.Loader)
	}

	reqMod := reg.Enable(vm)
	if opts.EnableConsole {
		console.Enable(vm)
	}

	return vm, reqMod, nil
}
