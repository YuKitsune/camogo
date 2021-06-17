package container

import (
	"fmt"
	"reflect"
)

type ServiceRegistration interface {
	Type() reflect.Type
	Resolve(Resolver) (interface{}, error)
}

type InstanceRegistration struct {
	targetType reflect.Type
	instance interface{}
}

func (r *InstanceRegistration) Type() reflect.Type {
	return r.targetType
}

func (r *InstanceRegistration) Resolve(_ Resolver) (interface{}, error) {
	return r.instance, nil
}

type FactoryRegistration struct {
	targetType reflect.Type
	factoryType reflect.Type
	factory reflect.Value
	lifetime Lifetime
	instance interface{}
}

func (r *FactoryRegistration) Type() reflect.Type {
	return r.targetType
}

func (r *FactoryRegistration) Resolve(c Resolver) (interface{}, error) {

	if r.lifetime == SingletonLifetime && r.instance != nil {
		return r.instance, nil
	}

	out, err := resolveFunc(c, r.factory)
	if err != nil {
		return nil, err
	}

	instance, err := getResult(out)
	if err != nil {
		return nil, err
	}

	if r.lifetime == SingletonLifetime {
		r.instance = instance
		return r.instance, nil
	}

	return instance, nil
}

func resolveFunc(r Resolver, fnValue reflect.Value) ([]reflect.Value, error) {
	var in []reflect.Value
	for i := 0; i < fnValue.Type().NumIn(); i++ {
		dependency, err := r.Resolve(fnValue.Type().In(i))
		if err != nil {
			return []reflect.Value{}, err
		}

		dependencyValue := reflect.ValueOf(dependency)
		in = append(in, dependencyValue)
	}

	return fnValue.Call(in), nil
}

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
			err = out[1].Interface().(error)
		} else {
			return nil, fmt.Errorf("if the factory returns two things, the second thing should be an error")
		}
	}

	return result, err
}