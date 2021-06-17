package container

import (
	"fmt"
	"reflect"
)

type Resolver interface{
	Resolve(p reflect.Type) (interface{}, error)
	ResolveInScope(interface{}) error
}

func resolveInScope(r Resolver, fn interface{}) error {
	fnValue := reflect.ValueOf(fn)
	err := validateScopeFunc(fnValue)
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

func validateScopeFunc(maybeFn reflect.Value) error {
	if maybeFn.IsNil() {
		return fmt.Errorf("scope cannot be nil")
	}

	if maybeFn.Type().NumOut() > 1 || maybeFn.Type().Out(1).Name() != "error" {
		return fmt.Errorf("scope can either return nothing or an error")
	}

	return nil
}