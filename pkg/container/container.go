package container

import (
	"fmt"
	"reflect"
)

type RegistrationFunc func(*Registrar) error

type Container interface {

	// RegisterModule registers the given Module with the Container
	RegisterModule(Module) error

	// Register invokes the given RegistrationFunc to register a set of services
	Register(RegistrationFunc) error

	Resolver
}

type defaultContainer struct {
	*Registrar
}

// New returns a new Container instance
func New() Container {
	return &defaultContainer{
		&Registrar{},
	}
}

func (c *defaultContainer) RegisterModule(m Module) error {
	return m.Register(c.Registrar)
}

func (c *defaultContainer) Register(fn RegistrationFunc) error {
	return fn(c.Registrar)
}

func (c *defaultContainer) Resolve(fn interface{}) error {

	if fn == nil {
		return fmt.Errorf("func cannot be nil")
	}

	// Ensure the fn returns the appropriate thing(s)
	fnValue := reflect.ValueOf(fn)
	err := validateScopeResults(fnValue)
	if err != nil {
		return err
	}

	out, err := resolveFunc(c, fnValue)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return nil
	}

	if len(out) == 1 && out[0].Type().Name() == "error" {
		if out[0].IsNil() {
			return nil
		}

		return out[0].Interface().(error)
	}

	return nil
}

func (c *defaultContainer) resolveType(p reflect.Type) (interface{}, error) {
	for _, registration := range c.registeredServices {
		if registration.Type() == p {
			return registration.Resolve(c)
		}
	}

	return nil, fmt.Errorf("no services of type %s were registered", p.Name())
}
