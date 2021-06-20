package container

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

type testInterface interface {
	GetValue() string
}

type testInstance struct {
	stringValue string
}

type testInstance2 struct {
	testInstance
}

func (v *testInstance) GetValue() string {
	return v.stringValue
}

type testModule struct {
	instancesToRegister []interface{}
	factoriesToRegister []struct {
		factory interface{}
		lifetime Lifetime
	}
}

func (m *testModule) Register(r *Registrar) error {
	for _, i := range m.instancesToRegister {
		r.RegisterInstance(i)
	}

	for _, f := range m.factoriesToRegister {
		err := r.RegisterFactory(f.factory, f.lifetime)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestRegistrarDoesNotAllowDuplicates(t *testing.T) {
	t.Run(
		"Using Instance",
		func (t *testing.T) {

			// Arrange
			instance1 := &testInstance{fmt.Sprintf("%s-1", t.Name())}
			instance2 := &testInstance{fmt.Sprintf("%s-2", t.Name())}
			r := &Registrar{}

			// Act
			err1 := r.RegisterInstance(instance1)
			err2 := r.RegisterInstance(instance2)

			// Assert
			assert.NoError(t, err1)
			assert.Error(t, err2)
		})
	t.Run(
		"Using Factory",
		func (t *testing.T) {

			// Arrange
			instance1 := &testInstance{fmt.Sprintf("%s-1", t.Name())}
			instance2 := &testInstance{fmt.Sprintf("%s-2", t.Name())}
			r := &Registrar{}

			// Act
			err1 := r.RegisterFactory(func () *testInstance { return instance1 }, SingletonLifetime)
			err2 := r.RegisterFactory(func () *testInstance { return instance2 }, SingletonLifetime)

			// Assert
			assert.NoError(t, err1)
			assert.Error(t, err2)
		})
}

func TestResolveType(t *testing.T) {
	t.Run(
		"From Module",
		func (t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			module := &testModule{instancesToRegister: []interface{} { instance }}
			c := New()
			err = c.RegisterModule(module)
			assert.NoError(t, err)

			// Act
			receivedInstance, err := c.ResolveType(reflect.TypeOf(instance))
			assert.NoError(t, err)

			// Assert
			assert.NotNil(t, receivedInstance)
			assert.Equal(t, instance.stringValue, receivedInstance.(*testInstance).stringValue)
		})
	t.Run(
		"From Func",
		func (t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			c := New()
			err = c.Register(func(r *Registrar) error {
				r.RegisterInstance(instance)
				return nil
			})
			assert.NoError(t, err)

			// Act
			receivedInstance, err := c.ResolveType(reflect.TypeOf(instance))
			assert.NoError(t, err)

			// Assert
			assert.NotNil(t, receivedInstance)
			assert.Equal(t, instance.stringValue, receivedInstance.(*testInstance).stringValue)
		})
}

func TestResolveInScope(t *testing.T) {
	t.Run(
		"From Module",
		func (t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			module := &testModule{instancesToRegister: []interface{} { instance }}
			c := New()
			err = c.RegisterModule(module)
			assert.NoError(t, err)

			// Act
			err = c.ResolveInScope(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
	t.Run(
		"From Func",
		func (t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			c := New()
			err = c.Register(func(r *Registrar) error {
				r.RegisterInstance(instance)
				return nil
			})
			assert.NoError(t, err)

			// Act
			err = c.ResolveInScope(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
}

func TestResolveFromInterface(t *testing.T) {

	// Arrange
	instance := &testInstance{t.Name()}
	c := New()
	err := c.Register(func (r *Registrar) error {
		return r.RegisterFactory(func () testInterface {
			return instance
		},
		SingletonLifetime)
	})
	assert.NoError(t, err)

	// Act
	err = c.ResolveInScope(func (res testInterface) {

		// Assert
		assert.Equal(t, instance.GetValue(), res.GetValue())
	})

	assert.NoError(t, err)
}

func TestResolveInScopeReturnsError(t *testing.T) {
	t.Run(
		"From provided func",
		func (t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			module := &testModule{instancesToRegister: []interface{} { instance }}
			c := New()
			err = c.RegisterModule(module)
			assert.NoError(t, err)

			// Act
			errorToReturn := fmt.Errorf(t.Name())
			foundError := c.ResolveInScope(func(receivedInstance *testInstance) error {
				return errorToReturn
			})

			assert.NotNil(t, foundError)
			assert.Equal(t, errorToReturn, foundError)
		})
	t.Run(
		"When unable to resolve",
		func (t *testing.T) {

			// Arrange
			c := New()

			// Act
			foundError := c.ResolveInScope(func(receivedInstance *testInstance) error {
				return nil
			})

			assert.NotNil(t, foundError)
		})
}

func TestFactoryIsValidated(t *testing.T) {

	// Valid
	t.Run("Returns one thing", func (t *testing.T) { testFactoryIsValidated(t, func() *testInstance { return nil }, true) })
	t.Run("Returns something or error", func (t *testing.T) { testFactoryIsValidated(t, func() (*testInstance, error) { return nil, nil }, true) })

	// Invalid
	t.Run("Returns nothing", func (t *testing.T) { testFactoryIsValidated(t, func() { }, false) })
	t.Run("Returns error", func (t *testing.T) { testFactoryIsValidated(t, func() error { return nil }, false) })
	t.Run("Returns many things", func (t *testing.T) { testFactoryIsValidated(t, func() (*testInstance, *testInstance) { return nil, nil }, false) })
}

func testFactoryIsValidated(t *testing.T, fn interface{}, shouldPass bool) {

	// Arrange
	registrar := &Registrar{}

	// Act
	err := registrar.RegisterFactory(fn, SingletonLifetime)

	// Assert
	if shouldPass {
		assert.NoError(t, err)
	} else {
		assert.NotNil(t, err)
	}
}

func TestResolveInScopeFuncIsValidated(t *testing.T) {

	// valid
	t.Run("Returns nothing", func (t *testing.T) { testResolveInScopeFuncIsValidated(t, func() { }, true) })
	t.Run("Returns error", func (t *testing.T) { testResolveInScopeFuncIsValidated(t, func() error { return nil }, true) })

	// Invalid
	t.Run("Returns something", func (t *testing.T) { testResolveInScopeFuncIsValidated(t, func() *testInstance { return nil }, false) })
	t.Run("Returns many things", func (t *testing.T) { testResolveInScopeFuncIsValidated(t, func() (*testInstance, *testInstance) { return nil, nil }, false) })
}

func testResolveInScopeFuncIsValidated(t *testing.T, fn interface{}, shouldPass bool) {

	// Arrange
	c := New()

	// Act
	err := c.ResolveInScope(fn)

	// Assert
	if shouldPass {
		assert.NoError(t, err)
	} else {
		assert.NotNil(t, err)
	}
}

func TestResolveSingletonResolvesSameInstance(t *testing.T) {

	// Arrange
	c := New()
	err := c.Register(func (r *Registrar) error {
		return r.RegisterFactory(func () *testInstance {
			return &testInstance{t.Name()}
		},
		SingletonLifetime)
	})
	assert.NoError(t, err)

	// Act / Assert
	for i := 0; i < 10; i ++ {
		err = c.ResolveInScope(func (res *testInstance)  {
			assert.Equal(t, t.Name(), res.GetValue())
		})
		assert.NoError(t, err)
	}
}

func TestResolveTransientResolvesNewInstance(t *testing.T) {

	// Arrange
	c := New()
	counter := 0
	err := c.Register(func (r *Registrar) error {
		return r.RegisterFactory(func () *testInstance {
			counter++
			return &testInstance{strconv.Itoa(counter)}
		},
		TransientLifetime)
	})
	assert.NoError(t, err)

	// Act / Assert
	var lastValue string
	for i := 0; i < 10; i ++ {
		err = c.ResolveInScope(func (res *testInstance)  {
			assert.NotEqual(t, lastValue, res.GetValue())
			lastValue = res.GetValue()
		})
		assert.NoError(t, err)
	}
}