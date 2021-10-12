package camogo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type testInterface interface {
	GetValue() string
}

type testInstance struct {
	stringValue string
}

func (v *testInstance) GetValue() string {
	return v.stringValue
}

func testInstanceName() string {
	return reflect.TypeOf(&testInstance{}).Name()
}

type testModule struct {
	instancesToRegister []interface{}
	factoriesToRegister []struct {
		factory  interface{}
		lifetime Lifetime
	}
}

func (m *testModule) Register(cb ContainerBuilder) error {
	for _, i := range m.instancesToRegister {
		err := cb.RegisterInstance(i)
		if err != nil {
			return err
		}
	}

	for _, f := range m.factoriesToRegister {
		err := cb.RegisterFactory(f.factory, f.lifetime)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestBuilderDoesNotAllowDuplicates(t *testing.T) {
	t.Run(
		"Using Instance",
		func(t *testing.T) {

			// Arrange
			instance1 := &testInstance{fmt.Sprintf("%s-1", t.Name())}
			instance2 := &testInstance{fmt.Sprintf("%s-2", t.Name())}
			cb := NewBuilder()

			// Act
			err1 := cb.RegisterInstance(instance1)
			err2 := cb.RegisterInstance(instance2)

			// Assert
			assert.NoError(t, err1)
			assert.Error(t, err2)
		})
	t.Run(
		"Using Factory",
		func(t *testing.T) {

			// Arrange
			instance1 := &testInstance{fmt.Sprintf("%s-1", t.Name())}
			instance2 := &testInstance{fmt.Sprintf("%s-2", t.Name())}
			cb := NewBuilder()

			// Act
			err1 := cb.RegisterFactory(func() *testInstance { return instance1 }, SingletonLifetime)
			err2 := cb.RegisterFactory(func() (*testInstance, error) { return instance2, nil }, TransientLifetime)

			// Assert
			assert.NoError(t, err1)
			assert.Error(t, err2)
		})
}

func TestFactoryIsValidated(t *testing.T) {

	testFactoryIsValidated := func(t *testing.T, fn interface{}, shouldPass bool) {

		// Arrange
		cb := NewBuilder()

		// Act
		err := cb.RegisterFactory(fn, SingletonLifetime)

		// Assert
		if shouldPass {
			assert.NoError(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}

	// Valid
	t.Run("Returns one instance", func(t *testing.T) {
		testFactoryIsValidated(t, func() *testInstance { return nil }, true)
	})
	t.Run("Returns instance or error", func(t *testing.T) {
		testFactoryIsValidated(t, func() (*testInstance, error) { return nil, nil }, true)
	})

	// Invalid
	t.Run("Returns nothing", func(t *testing.T) {
		testFactoryIsValidated(t, func() {}, false)
	})
	t.Run("Only returns error", func(t *testing.T) {
		testFactoryIsValidated(t, func() error { return nil }, false)
	})
	t.Run("Returns more than one instance", func(t *testing.T) {
		testFactoryIsValidated(t, func() (*testInstance, *testInstance) { return nil, nil }, false)
	})
	t.Run("Returns more than one instance and error", func(t *testing.T) {
		testFactoryIsValidated(t, func() (*testInstance, *testInstance, error) { return nil, nil, nil }, false)
	})
}

func BenchmarkRegisterFactory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cb := NewBuilder()
		_ = cb.RegisterFactory(func() (*testInstance, error) {
			return &testInstance{b.Name()}, nil
		}, TransientLifetime)
	}
}
