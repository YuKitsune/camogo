package container

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type dummyInstance struct {
	stringValue string
}
type testModuleWithInstance struct {
	instanceToRegister interface{}
}
func (m *testModuleWithInstance) Register(r *Registrar) error {
	r.RegisterInstance(m.instanceToRegister)
	return nil
}

type testModuleWithFactory struct {
	factoryToRegister interface{}
	lifetime Lifetime
}
func (m *testModuleWithFactory) Register(r *Registrar) error {
	return r.RegisterFactory(m.factoryToRegister, m.lifetime)
}

func TestCanRegisterServices(t *testing.T) {
	t.Run(
		"Using Module",
		func (t *testing.T) {
			t.Run(
				"Instance",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testInstance := &dummyInstance{testValue}
					module := &testModuleWithInstance{testInstance}
					c := New()

					// Act / Assert
					err := c.RegisterModule(module)
					assert.NoError(t, err)
				})
			t.Run(
				"Factory",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testFactory := func () *dummyInstance {
						return &dummyInstance{testValue}
					}
					module := &testModuleWithFactory{testFactory, TransientLifetime}
					c := New()

					// Act / Assert
					err := c.RegisterModule(module)
					assert.NoError(t, err)
				})
			t.Run(
				"Factory with error",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testFactory := func () (*dummyInstance, error) {
						return &dummyInstance{testValue}, nil
					}
					module := &testModuleWithFactory{testFactory, TransientLifetime}
					c := New()

					// Act / Assert
					err := c.RegisterModule(module)
					assert.NoError(t, err)
				})
		})
	t.Run(
		"Using Func",
		func (t *testing.T) {
			t.Run(
				"Instance",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testInstance := &dummyInstance{testValue}
					fn := func (r *Registrar) error {
						r.RegisterInstance(testInstance)
						return nil
					}
					c := New()

					// Act / Assert
					err := c.RegisterWith(fn)
					assert.NoError(t, err)
				})
			t.Run(
				"Factory",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testFactory := func () (*dummyInstance, error) {
						return &dummyInstance{testValue}, nil
					}
					fn := func (r *Registrar) error {
						return r.RegisterFactory(testFactory, TransientLifetime)
					}
					c := New()

					// Act / Assert
					err := c.RegisterWith(fn)
					assert.NoError(t, err)
				})
		})
}

func TestCanResolveServices(t *testing.T) {
	t.Run(
		"Using Type",
		func (t *testing.T) {
			t.Run(
				"Instance",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testInstance := &dummyInstance{testValue}
					module := &testModuleWithInstance{testInstance}
					c := New()
					err := c.RegisterModule(module)
					assert.NoError(t, err)

					// Act
					result, err := c.Resolve(reflect.TypeOf(testInstance))

					// Assert
					assert.NoError(t, err)
					assert.NotNil(t, result)
					assert.Equal(t, testValue, result.(*dummyInstance).stringValue)
				})
			t.Run(
				"Factory",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testFactory := func () *dummyInstance {
						return &dummyInstance{testValue}
					}
					module := &testModuleWithFactory{testFactory, TransientLifetime}
					c := New()
					err := c.RegisterModule(module)
					assert.NoError(t, err)

					// Act
					result, err := c.Resolve(reflect.TypeOf(&dummyInstance{}))

					// Assert
					assert.NoError(t, err)
					assert.NotNil(t, result)
					assert.Equal(t, testValue, result.(*dummyInstance).stringValue)
				})
		})
	t.Run(
		"Using Func",
		func (t *testing.T) {
			t.Run(
				"Instance",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testInstance := &dummyInstance{testValue}
					module := &testModuleWithInstance{testInstance}
					c := New()
					err := c.RegisterModule(module)
					assert.NoError(t, err)

					// Act / Assert
					err = c.ResolveInScope(func (result *dummyInstance) {
						assert.NotNil(t, result)
						assert.Equal(t, testValue, result.stringValue)
					})

					assert.NoError(t, err)
				})
			t.Run(
				"Factory",
				func (t *testing.T) {

					// Arrange
					testValue := t.Name()
					testFactory := func () *dummyInstance {
						return &dummyInstance{testValue}
					}
					module := &testModuleWithFactory{testFactory, TransientLifetime}
					c := New()
					err := c.RegisterModule(module)
					assert.NoError(t, err)

					// Act / Assert
					err = c.ResolveInScope(func (result *dummyInstance) {
						assert.NotNil(t, result)
						assert.Equal(t, testValue, result.stringValue)
					})

					assert.NoError(t, err)
				})
		})
}

func TestFactoryReturnsError(t *testing.T) {
	t.Run(
		"Using Type",
		func (t *testing.T) {

			// Arrange
			c := New()

			// Act
			result, err := c.Resolve(reflect.TypeOf(&dummyInstance{}))

			// Assert
			assert.Nil(t, result)
			assert.NotNil(t, err)
		})
	t.Run(
		"Using Func",
		func (t *testing.T) {

			// Arrange
			testError := fmt.Errorf(t.Name())
			testFactory := func () (*dummyInstance, error) {
				return nil, testError
			}
			module := &testModuleWithFactory{testFactory, TransientLifetime}
			c := New()
			err := c.RegisterModule(module)
			assert.NoError(t, err)

			// Act / Assert
			err = c.ResolveInScope(func (result *dummyInstance) { })

			assert.NotNil(t, err)
			assert.Equal(t, testError, err)
		})
}
