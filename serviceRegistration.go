package camogo

import (
	"fmt"
	"reflect"
)

type serviceRegistration interface {
	Type() reflect.Type
	Resolve(Resolver) (interface{}, error)
}

type instanceRegistration struct {
	targetType reflect.Type
	instance   interface{}
}

func (r *instanceRegistration) Type() reflect.Type {
	return r.targetType
}

func (r *instanceRegistration) Resolve(_ Resolver) (interface{}, error) {
	return r.instance, nil
}

type factoryRegistration struct {
	targetType  reflect.Type
	factoryType reflect.Type
	factory     reflect.Value
	lifetime    Lifetime
	instance    interface{}
}

func (r *factoryRegistration) Type() reflect.Type {
	return r.targetType
}

func (r *factoryRegistration) Resolve(c Resolver) (interface{}, error) {

	// If this service is registered as a singleton, and we've already resolved an instance before, just return that
	// instance
	if r.lifetime == SingletonLifetime && r.instance != nil {
		return r.instance, nil
	}

	// Execute the factory method
	out, err := resolveFunc(c, r.factory)
	if err != nil {
		return nil, err
	}

	// Convert the reflection values to something we can use
	instance, err := getResult(out)
	if err != nil {
		return nil, err
	}

	// If this service is registered as a singleton, then store this new instance for later
	if r.lifetime == SingletonLifetime {
		r.instance = instance
		return r.instance, nil
	}

	return instance, nil
}

// resolveFunc executes the given fnValue as a func and uses the given Resolver to resolve any arguments.
func resolveFunc(r Resolver, fnValue reflect.Value) ([]reflect.Value, error) {
	var in []reflect.Value
	for i := 0; i < fnValue.Type().NumIn(); i++ {
		arg, err := r.resolveType(fnValue.Type().In(i))
		if err != nil {
			return []reflect.Value{}, err
		}

		dependencyValue := reflect.ValueOf(arg)
		in = append(in, dependencyValue)
	}

	return fnValue.Call(in), nil
}

// getResult converts the given []reflect.Value into (interface{}, error)
func getResult(out []reflect.Value) (interface{}, error) {

	if len(out) == 0 {
		return nil, fmt.Errorf("the factory did not return anything")
	}

	if len(out) > 2 {
		return nil, fmt.Errorf("the factory returned more results than expected")
	}

	result := out[0].Interface()

	var err error
	if len(out) == 2 {
		if out[1].Type().Name() == "error" {
			err = errorOrNil(out[1])
		} else {
			return nil, fmt.Errorf("if the factory returns two things, the second thing should be an error")
		}
	}

	return result, err
}
