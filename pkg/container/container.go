package container

import (
	"fmt"
	"reflect"
)

type Container interface {
	RegisterModule(Module) error
}

type defaultContainer struct {
	registrar *Registrar
}

func New() Container {
	return &defaultContainer{
		registrar: &Registrar{},
	}
}

func (c *defaultContainer) RegisterModule(m Module) error {
	return m.Register(c.registrar)
}


func (c *defaultContainer) Resolve(p reflect.Type) (interface{}, error) {
	for _, registration := range c.registrar.registeredServices {
		if registration.Type() == p {
			return registration.Resolve(c)
		}
	}

	return nil, fmt.Errorf("no services of type %s were registered", p.Name())
}

func (c *defaultContainer) ResolveInScope(fn interface{}) error {
	return resolveInScope(c, fn)
}
