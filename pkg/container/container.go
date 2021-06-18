package container

import (
	"fmt"
	"reflect"
)

type Container interface {
	RegisterModule(Module) error
	RegisterWith(RegistrationFunc) error
	Resolver
}

type defaultContainer struct {
	*Registrar
}

func New() Container {
	return &defaultContainer{
		&Registrar{},
	}
}

func (c *defaultContainer) RegisterModule(m Module) error {
	return m.Register(c.Registrar)
}

func (c *defaultContainer) RegisterWith(fn RegistrationFunc) error {
	return fn(c.Registrar)
}

func (c *defaultContainer) Resolve(p reflect.Type) (interface{}, error) {
	for _, registration := range c.registeredServices {
		if registration.Type() == p {
			return registration.Resolve(c)
		}
	}

	return nil, fmt.Errorf("no services of type %s were registered", p.Name())
}

func (c *defaultContainer) ResolveInScope(fn interface{}) error {
	return resolveInScope(c, fn)
}
