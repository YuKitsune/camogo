package camogo

import (
	"fmt"
	"reflect"
)

type ServicePredicate = func (p reflect.Type) bool

// Container is an IoC container
type Container interface {

	// Resolve will invoke the given function, resolving all of the arguments.
	//	The returned error will either be from the Resolver failing to resolve an argument, or from the provided
	//	function if any error is returned
	Resolve(interface{}) error

	// ResolveWithResult will invoke the given function, resolving all of the arguments.
	//  The provided function must return something, which this function will return via it's first return value
	//	The returned error will either be from the Resolver failing to resolve an argument, or from the provided
	//	function if any error is returned
	ResolveWithResult(interface{}) (interface{}, error)

	// NewChild will create a new child Container instance from the current Container instance.
	//	Services registered with a ScopedLifetime will be treated as a SingletonLifetime per child Container.
	NewChild() Container

	// ResolveType will resolve the service with the given reflect.Type
	ResolveType(p reflect.Type) (interface{}, error)

	// ResolveMatchingType will resolve the service that matches the given ServicePredicate
	ResolveMatchingType(ServicePredicate) (interface{}, error)

	// ResolveMatchingTypes will resolve the services that match the given ServicePredicate
	ResolveMatchingTypes(ServicePredicate) ([]interface{}, error)
}

type defaultContainer struct {
	parent   Container
	services []service
	containerType reflect.Type
}

func (c *defaultContainer) ContainerType() reflect.Type {
	if c.containerType == nil {
		c.containerType = reflect.TypeOf(c)
	}

	return c.containerType
}

func (c *defaultContainer) Resolve(fn interface{}) error {

	if fn == nil {
		return fmt.Errorf("func cannot be nil")
	}

	// Ensure the fn returns the appropriate thing(s)
	fnValue := reflect.ValueOf(fn)
	err := validateFuncResults(fnValue)
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
		return errorOrNil(out[0])
	}

	return nil
}

func (c *defaultContainer) ResolveWithResult(fn interface{}) (interface{}, error) {

	if fn == nil {
		return nil, fmt.Errorf("func cannot be nil")
	}

	// Ensure the fn returns the appropriate thing(s)
	fnValue := reflect.ValueOf(fn)
	err := validateFuncResultsWithResult(fnValue)
	if err != nil {
		return nil, err
	}

	out, err := resolveFunc(c, fnValue)
	if err != nil {
		return nil, err
	}

	res := valueOrNil(out[0])
	var returnedErr error
	if len(out) == 2 {
		returnedErr = errorOrNil(out[1])
	}

	return res, returnedErr
}

func (c *defaultContainer) ResolveType(p reflect.Type) (interface{}, error) {

	if c.ContainerType().AssignableTo(p) {
		return c, nil
	}

	for _, svc := range c.services {
		if svc.Type() == p {
			return svc.Resolve(c)
		}
	}

	if c.parent != nil {
		return c.parent.ResolveType(p)
	}

	return nil, fmt.Errorf("no services of type %s were registered", p.Name())
}

func (c *defaultContainer) ResolveMatchingType(predicate ServicePredicate) (interface{}, error) {
	for _, s := range c.services {
		if predicate(s.Type()) {
			return s.Resolve(c)
		}
	}

	if c.parent != nil {
		return c.parent.ResolveMatchingType(predicate)
	}

	return nil, fmt.Errorf("no services matching the given predicate were registered")
}

func (c *defaultContainer) ResolveMatchingTypes(predicate ServicePredicate) ([]interface{}, error) {

	var foundServices []interface{}
	var errs errors
	for _, s := range c.services {
		if predicate(s.Type()) {
			resolvedService, err := s.Resolve(c)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			foundServices = append(foundServices, resolvedService)
		}
	}

	if c.parent != nil {
		foundServicesFromParent, errFromParent := c.parent.ResolveMatchingTypes(predicate)
		if errsFromParent, ok := errFromParent.(errors); ok && len(errsFromParent) > 0{
			errs = append(errs, errsFromParent)
		} else {
			foundServices = append(foundServices, foundServicesFromParent)
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	if len(foundServices) > 0 {
		return foundServices, nil
	}

	return nil, fmt.Errorf("no services matching the given predicate were registered")
}

func (c *defaultContainer) NewChild() Container {
	var svcs []service

	for _, svc := range c.services {
		switch v := svc.(type) {
		case *serviceFactory:

			// Kinda hacky
			if v.lifetime == ScopedLifetime {
				sf := v.copy()
				sf.lifetime = SingletonLifetime
				sf.instance = nil
				svcs = append(svcs, sf)
			}
		}
	}

	return &defaultContainer{
		parent:   c,
		services: svcs,
	}
}

func validateFuncResults(fnValue reflect.Value) error {
	funcType := fnValue.Type()

	if funcType.NumOut() == 0 || (funcType.NumOut() == 1 && funcType.Out(0).Name() == "error") {
		return nil
	}

	return fmt.Errorf("func must return either nothing or an error")
}

func validateFuncResultsWithResult(fnValue reflect.Value) error {
	funcType := fnValue.Type()

	if (funcType.NumOut() == 1 && funcType.Out(0).Name() != "error") ||
		(funcType.NumOut() == 2 && funcType.Out(0).Name() != "error" && funcType.Out(1).Name() == "error") {
		return nil
	}

	return fmt.Errorf("func must return something")
}
