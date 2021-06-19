package container

import (
	"fmt"
	"reflect"
)

type Resolver interface{
	ResolveType(p reflect.Type) (interface{}, error)
	ResolveInScope(interface{}) error
}

func validateScopeResults(fnValue reflect.Value) error {
	funcType := fnValue.Type()

	if funcType.NumOut() == 0 || (funcType.NumOut() == 1 && funcType.Out(0).Name() == "error") {
		return nil
	}

	return fmt.Errorf("func must return either nothing or an error")
}