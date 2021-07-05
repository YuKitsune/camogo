package camogo

import (
	"fmt"
	"reflect"
)

// Resolver provides mechanisms for resolving services
type Resolver interface {

	// Resolve will invoke the given function, resolving all of the arguments.
	//	The returned error will either be from the Resolver failing to resolve an argument, or from the provided
	//	function if any error is returned
	Resolve(interface{}) error

	// ResolveWithResult will invoke the given function, resolving all of the arguments.
	//  The provided function must return something, which this function will return via it's first return value
	//	The returned error will either be from the Resolver failing to resolve an argument, or from the provided
	//	function if any error is returned
	ResolveWithResult(interface{}) (interface{}, error)

	// resolveType will resolve the service with the given reflect.Type
	resolveType(p reflect.Type) (interface{}, error)
}

func validateScopeResults(fnValue reflect.Value) error {
	funcType := fnValue.Type()

	if funcType.NumOut() == 0 || (funcType.NumOut() == 1 && funcType.Out(0).Name() == "error") {
		return nil
	}

	return fmt.Errorf("func must return either nothing or an error")
}

func validateScopeResultsWithResponse(fnValue reflect.Value) error {
	funcType := fnValue.Type()

	if (funcType.NumOut() == 1 && funcType.Out(0).Name() != "error") ||
		(funcType.NumOut() == 2 && funcType.Out(0).Name() != "error" && funcType.Out(1).Name() == "error") {
		return nil
	}

	return fmt.Errorf("func must return something")
}