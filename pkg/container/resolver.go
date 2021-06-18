package container

import (
	"fmt"
	"reflect"
)

type Resolver interface{
	Resolve(p reflect.Type) (interface{}, error)
	ResolveInScope(interface{}) error
}

// resolveInScope executes the given func (fn) using the given Resolver to resolve any arguments.
func resolveInScope(r Resolver, fn interface{}) error {

	if r == nil {
		return fmt.Errorf("resolver cannot be nil")
	}

	if fn == nil {
		return fmt.Errorf("func cannot be nil")
	}

	// Ensure the fn returns the appropriate thing(s)
	fnValue := reflect.ValueOf(fn)
	err := validateScopeResults(fnValue)
	if err != nil {
		return err
	}

	out, err := resolveFunc(r, fnValue)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return nil
	}

	if len(out) == 1 && out[0].Type().Name() == "error" {
		return out[0].Interface().(error)
	}

	return nil
}

func validateScopeResults(fnValue reflect.Value) error {
	funcType := fnValue.Type()
	if (funcType.NumOut() > 1 && (funcType.Out(0).Name() == "error" || funcType.Out(1).Name() != "error")) ||
		funcType.NumOut() > 2 {
		return fmt.Errorf("func must return either nothing, an error, or some result and an error")
	}

	return nil
}