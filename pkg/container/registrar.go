package container

import (
	"fmt"
	"reflect"
)

type RegistrationExistsError struct {
	TargetType reflect.Type
}

func (e *RegistrationExistsError) Error() string {
	return fmt.Sprintf("a registration for %s already exists", e.TargetType.Name())
}

type Lifetime int

const (

	// TransientLifetime specifies that a factory should be invoked every time it is requested
	TransientLifetime Lifetime = iota

	// SingletonLifetime specifies that a factory should only be invoked once, and the result should be re-used for all
	//	subsequent requests
	SingletonLifetime
)

// Registrar is responsible for registering services and keeping track of them
type Registrar struct {
	registeredServices []serviceRegistration
}

// RegistrationExists returns true if the given reflect.Type has been registered, false otherwise
func (m *Registrar) RegistrationExists(t reflect.Type) bool {
	for _, service := range m.registeredServices {
		if service.Type() == t {
			return true
		}
	}

	return false
}

// RegisterInstance registers the given instance
func (m *Registrar) RegisterInstance(instance interface{}) error {
	registration := &instanceRegistration{
		targetType: reflect.TypeOf(instance),
		instance:   instance,
	}

	if m.RegistrationExists(registration.Type()) {
		return &RegistrationExistsError{registration.Type()}
	}

	m.registeredServices = append(m.registeredServices, registration)
	return nil
}

// RegisterFactory registers the given factory function with the given Lifetime
func (m *Registrar) RegisterFactory(factory interface{}, lifetime Lifetime) error {

	fnType := reflect.TypeOf(factory)
	fn := reflect.ValueOf(factory)
	err := validateFactory(fn)
	if err != nil {
		return err
	}

	registration := &factoryRegistration{
		targetType:  reflect.TypeOf(factory).Out(0),
		factoryType: fnType,
		factory:     fn,
		lifetime:    lifetime,
	}

	if m.RegistrationExists(registration.Type()) {
		return &RegistrationExistsError{registration.Type()}
	}

	m.registeredServices = append(m.registeredServices, registration)
	return nil
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

	return nil
}
