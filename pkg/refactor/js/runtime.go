package js

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/refactorio/pkg/refactor/js/modules"
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
	Registry      *modules.Registry
	EnableConsole bool
	DisableTime   bool
	DisableRandom bool
	AllowFileJS   bool
}

// NewRuntime creates a goja runtime with Node-style require() and a strict allow-list of modules.
func NewRuntime(opts RuntimeOptions) (*goja.Runtime, *require.RequireModule, error) {
	vm := goja.New()
	reg := require.NewRegistry(require.WithLoader(disabledSourceLoader(opts.AllowFileJS)))

	if opts.Registry != nil && len(opts.Modules) > 0 {
		return nil, nil, errors.New("js runtime: use either Registry or Modules, not both")
	}
	if opts.Registry != nil {
		opts.Registry.Enable(reg)
	} else {
		for _, mod := range opts.Modules {
			if mod.Name == "" {
				return nil, nil, errors.New("js runtime: module name is required")
			}
			if mod.Loader == nil {
				return nil, nil, errors.Errorf("js runtime: module loader missing for %s", mod.Name)
			}
			reg.RegisterNativeModule(mod.Name, mod.Loader)
		}
	}

	reqMod := reg.Enable(vm)
	if opts.EnableConsole {
		console.Enable(vm)
	}
	if opts.DisableTime {
		if err := disableTime(vm); err != nil {
			return nil, nil, err
		}
	}
	if opts.DisableRandom {
		if err := disableRandom(vm); err != nil {
			return nil, nil, err
		}
	}

	return vm, reqMod, nil
}

func disabledSourceLoader(allowFileJS bool) require.SourceLoader {
	if allowFileJS {
		return nil
	}
	return func(string) ([]byte, error) {
		return nil, require.ModuleFileDoesNotExistError
	}
}

func disableTime(vm *goja.Runtime) error {
	dateValue := vm.Get("Date")
	if dateValue == nil || goja.IsUndefined(dateValue) || goja.IsNull(dateValue) {
		return nil
	}
	dateObj := dateValue.ToObject(vm)
	return dateObj.Set("now", func() int64 { return 0 })
}

func disableRandom(vm *goja.Runtime) error {
	mathValue := vm.Get("Math")
	if mathValue == nil || goja.IsUndefined(mathValue) || goja.IsNull(mathValue) {
		return nil
	}
	mathObj := mathValue.ToObject(vm)
	return mathObj.Set("random", func() float64 { return 0 })
}
