package camogo

import (
	"fmt"
	"reflect"
)

// Lifetime defined the lifetime of a service in an IoC container
type Lifetime int

const (

	// TransientLifetime specifies that a factory should be invoked every time it is requested
	TransientLifetime Lifetime = iota

	// SingletonLifetime specifies that a factory should only be invoked once, and the result should be re-used for all
	//	subsequent requests
	SingletonLifetime
)

type service interface {
	Type() reflect.Type
	Resolve(Container) (interface{}, error)
}

type serviceInstance struct {
	targetType reflect.Type
	instance   interface{}
}

func (s *serviceInstance) Type() reflect.Type {
	return s.targetType
}

func (s *serviceInstance) Resolve(_ Container) (interface{}, error) {
	return s.instance, nil
}

type serviceFactory struct {
	targetType  reflect.Type
	factoryType reflect.Type
	factory     reflect.Value
	lifetime    Lifetime
	instance    interface{}
}

func (s *serviceFactory) Type() reflect.Type {
	return s.targetType
}

func (s *serviceFactory) Resolve(c Container) (interface{}, error) {

	// If this service is registered as a singleton, and we've already resolved an instance before, just return that
	// instance
	if s.lifetime == SingletonLifetime && s.instance != nil {
		return s.instance, nil
	}

	// Execute the factory method
	out, err := resolveFunc(c, s.factory)
	if err != nil {
		return nil, err
	}

	// Convert the reflection values to something we can use
	instance, err := getResult(out)
	if err != nil {
		return nil, err
	}

	// If this service is registered as a singleton, then store this new instance for later
	if s.lifetime == SingletonLifetime {
		s.instance = instance
		return s.instance, nil
	}

	return instance, nil
}

// resolveFunc executes the given fnValue as a func and uses the given Resolver to resolve any arguments.
func resolveFunc(c Container, fnValue reflect.Value) ([]reflect.Value, error) {
	var in []reflect.Value
	for i := 0; i < fnValue.Type().NumIn(); i++ {
		arg, err := c.resolveType(fnValue.Type().In(i))
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
