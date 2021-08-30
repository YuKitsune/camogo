package camogo

import (
	"fmt"
	"reflect"
)

type ContainerBuilder interface {

	// RegistrationExists returns true if the given reflect.Type has been registered, false otherwise
	RegistrationExists(reflect.Type) bool

	// RegisterInstance registers the given instance
	RegisterInstance(interface{}) error

	// RegisterFactory registers the given factory function with the given Lifetime
	RegisterFactory(interface{}, Lifetime) error

	// RegisterModule registers the given Module with the Container
	RegisterModule(Module) error

	// Build returns a new Container instance
	Build() Container
}

type containerBuilder struct {
	services []service
}

func NewBuilder() ContainerBuilder {
	return &containerBuilder{}
}

func (cb *containerBuilder) RegistrationExists(t reflect.Type) bool {
	for _, service := range cb.services {
		if service.Type() == t {
			return true
		}
	}

	return false
}


func (cb *containerBuilder) RegisterInstance(instance interface{}) error {
	registration := &serviceInstance{
		targetType: reflect.TypeOf(instance),
		instance:   instance,
	}

	if cb.RegistrationExists(registration.Type()) {
		return fmt.Errorf("a %s instance has already been registered", registration.Type().Name())
	}

	cb.services = append(cb.services, registration)
	return nil
}


func (cb *containerBuilder) RegisterFactory(factory interface{}, lifetime Lifetime) error {

	fnType := reflect.TypeOf(factory)
	fn := reflect.ValueOf(factory)
	err := validateFactory(fn)
	if err != nil {
		return err
	}

	registration := &serviceFactory{
		targetType:  reflect.TypeOf(factory).Out(0),
		factoryType: fnType,
		factory:     fn,
		lifetime:    lifetime,
	}

	if cb.RegistrationExists(registration.Type()) {
		return fmt.Errorf("a factory for %s has already been registered", registration.Type().Name())
	}

	cb.services = append(cb.services, registration)
	return nil
}

func (cb *containerBuilder) RegisterModule(module Module) error {
	return module.Register(cb)
}

func (cb *containerBuilder) Build() Container {
	return &defaultContainer{
		cb.services,
	}
}

func validateFactory(maybeFn reflect.Value) error {
	if maybeFn.IsNil() {
		return fmt.Errorf("factory cannot be nil")
	}

	if maybeFn.Type().NumOut() == 0 {
		return fmt.Errorf("factory does not return anything")
	}

	if maybeFn.Type().NumOut() == 1 && maybeFn.Type().Out(0).Name() == "error" {
		return fmt.Errorf("factory only returns an error")
	}

	if maybeFn.Type().Out(0).Name() == "error" {
		return fmt.Errorf("the first value returned from a factory must not be an error")
	}

	if maybeFn.Type().NumOut() == 2 && maybeFn.Type().Out(1).Name() != "error" {
		return fmt.Errorf("if the factory returns two values, the second one must be an error")
	}

	if maybeFn.Type().NumOut() > 2 {
		return fmt.Errorf("factory ccannot return more than two things")
	}

	return nil
}
