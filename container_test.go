package camogo_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yukitsune/camogo"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Run(
		"From Module",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}
			module := &testModule{instancesToRegister: []interface{}{instance}}

			cb := camogo.NewBuilder()
			err = cb.RegisterModule(module)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			err = c.Resolve(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
	t.Run(
		"From Instance",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterInstance(instance)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			err = c.Resolve(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
	t.Run(
		"From Factory",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterFactory(func() *testInstance {
				return instance
			}, camogo.TransientLifetime)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			err = c.Resolve(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
	t.Run(
		"From Factory with error",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterFactory(func() (*testInstance, error) {
				return instance, nil
			}, camogo.TransientLifetime)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			err = c.Resolve(func(receivedInstance *testInstance) {

				// Assert
				assert.NotNil(t, receivedInstance)
				assert.Equal(t, instance.stringValue, receivedInstance.stringValue)
			})

			assert.NoError(t, err)
		})
}

func TestResolveReturnsError(t *testing.T) {
	t.Run(
		"From provided func",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterInstance(instance)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			errorToReturn := fmt.Errorf(t.Name())
			foundError := c.Resolve(func(receivedInstance *testInstance) error {
				return errorToReturn
			})

			assert.NotNil(t, foundError)
			assert.Equal(t, errorToReturn, foundError)
		})
	t.Run(
		"When unable to resolve",
		func(t *testing.T) {

			// Arrange
			cb := camogo.NewBuilder()
			c := cb.Build()

			// Act
			foundError := c.Resolve(func(receivedInstance *testInstance) error {
				return nil
			})

			assert.NotNil(t, foundError)
		})
}

func TestResolveReturnsResult(t *testing.T) {
	t.Run(
		"With error",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterInstance(instance)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			errorToReturn := fmt.Errorf(t.Name())
			testName, foundError := c.ResolveWithResult(func(receivedInstance *testInstance) (interface{}, error) {
				return t.Name(), errorToReturn
			})

			assert.NotNil(t, foundError)
			assert.Equal(t, t.Name(), testName)
			assert.Equal(t, errorToReturn, foundError)
		})
	t.Run(
		"Without error",
		func(t *testing.T) {

			// Arrange
			var err error
			instance := &testInstance{t.Name()}

			cb := camogo.NewBuilder()
			err = cb.RegisterInstance(instance)
			assert.NoError(t, err)

			c := cb.Build()

			// Act
			testName, foundError := c.ResolveWithResult(func(receivedInstance *testInstance) interface{} {
				return t.Name()
			})

			assert.Nil(t, foundError)
			assert.Equal(t, t.Name(), testName)
		})
}

func TestResolveFuncIsValidated(t *testing.T) {

	testResolveFuncIsValidated := func(t *testing.T, fn interface{}, shouldPass bool) {

		// Arrange
		cb := camogo.NewBuilder()
		c := cb.Build()

		// Act
		err := c.Resolve(fn)

		// Assert
		if shouldPass {
			assert.NoError(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}

	// valid
	t.Run("Returns nothing", func(t *testing.T) {
		testResolveFuncIsValidated(t, func() {}, true)
	})
	t.Run("Returns error", func(t *testing.T) {
		testResolveFuncIsValidated(t, func() error { return nil }, true)
	})

	// Invalid
	t.Run("Returns something", func(t *testing.T) {
		testResolveFuncIsValidated(t, func() *testInstance { return nil }, false)
	})
	t.Run("Returns many things", func(t *testing.T) {
		testResolveFuncIsValidated(t, func() (*testInstance, *testInstance) { return nil, nil }, false)
	})
}

func TestResolveWithResultFuncIsValidated(t *testing.T) {

	testResolveWithResultFuncIsValidated := func(t *testing.T, fn interface{}, shouldPass bool) {

		// Arrange
		cb := camogo.NewBuilder()
		c := cb.Build()

		// Act
		_, err := c.ResolveWithResult(fn)

		// Assert
		if shouldPass {
			assert.NoError(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}

	// valid
	t.Run("Returns something", func(t *testing.T) {
		testResolveWithResultFuncIsValidated(t, func() *testInstance { return nil }, true)
	})
	t.Run("Returns something with error", func(t *testing.T) {
		testResolveWithResultFuncIsValidated(t, func() (*testInstance, error) { return nil, nil }, true)
	})

	// Invalid
	t.Run("Returns nothing", func(t *testing.T) {
		testResolveWithResultFuncIsValidated(t, func() {}, false)
	})
	t.Run("Returns only error", func(t *testing.T) {
		testResolveWithResultFuncIsValidated(t, func() error { return nil }, false)
	})
	t.Run("Returns many things", func(t *testing.T) {
		testResolveWithResultFuncIsValidated(t, func() (*testInstance, *testInstance) { return nil, nil }, false)
	})
}

func TestResolveSingletonResolvesSameInstance(t *testing.T) {

	// Arrange
	var err error
	cb := camogo.NewBuilder()
	err = cb.RegisterFactory(func() *testInstance {
		return &testInstance{strconv.Itoa(rand.Int())}
	},
		camogo.SingletonLifetime)
	assert.NoError(t, err)

	c := cb.Build()

	// Act / Assert
	var firstValue string
	for i := 0; i < 10; i++ {
		res, err := c.ResolveWithResult(func(res *testInstance) string {
			return res.GetValue()
		})

		assert.NotNil(t, res)
		assert.NoError(t, err)

		resString := res.(string)
		if i == 0 {
			firstValue = resString
			continue
		}

		assert.Equal(t, firstValue, resString)
	}
}

func TestResolveTransientResolvesNewInstance(t *testing.T) {

	// Arrange
	cb := camogo.NewBuilder()
	counter := 0
	err := cb.RegisterFactory(func() *testInstance {
		counter++
		return &testInstance{strconv.Itoa(counter)}
	},
		camogo.TransientLifetime)
	assert.NoError(t, err)

	c := cb.Build()

	// Act / Assert
	var lastValue string
	for i := 0; i < 10; i++ {
		res, err := c.ResolveWithResult(func(res *testInstance) string {
			return res.GetValue()
		})

		assert.NotNil(t, res)
		assert.NoError(t, err)

		resString := res.(string)
		assert.NotEqual(t, lastValue, resString)

		lastValue = resString
	}
}

func TestResolveScopedResolvedSameInstancePerScope(t *testing.T) {

	// Arrange
	var err error
	cb := camogo.NewBuilder()
	err = cb.RegisterFactory(func() *testInstance {
		return &testInstance{strconv.Itoa(rand.Int())}
	},
		camogo.ScopedLifetime)
	assert.NoError(t, err)

	c0 := cb.Build()
	c1 := c0.NewChild()
	c2 := c0.NewChild()

	// Act / Assert
	var firstValue0 string
	var firstValue1 string
	var firstValue2 string
	for i := 0; i < 10; i++ {
		fn := func(ctr camogo.Container) (string, error) {
			res, err := ctr.ResolveWithResult(func(res *testInstance) string {
				return res.GetValue()
			})

			resString := res.(string)
			return resString, err
		}

		res0, err := fn(c0)
		assert.NoError(t, err)
		assert.NotNil(t, res0)

		res1, err := fn(c1)
		assert.NoError(t, err)
		assert.NotNil(t, res1)

		res2, err := fn(c2)
		assert.NoError(t, err)
		assert.NotNil(t, res2)

		assert.NotEqual(t, res0, res1, res2)

		if i == 0 {
			firstValue0 = res0
			firstValue1 = res1
			firstValue2 = res2
			continue
		}

		assert.Equal(t, firstValue0, res0)
		assert.Equal(t, firstValue1, res1)
		assert.Equal(t, firstValue2, res2)
	}
}

func TestResolveSingletonResolvesSameInstanceFromParent(t *testing.T) {

	const nestLevel int = 10

	// Arrange
	var err error
	cb := camogo.NewBuilder()
	err = cb.RegisterFactory(func() *testInstance {
		return &testInstance{strconv.Itoa(rand.Int())}
	},
		camogo.SingletonLifetime)
	assert.NoError(t, err)

	cr := cb.Build()
	c := cr.NewChild()

	// Act / Assert
	var firstValue string
	for n := 0; n < nestLevel; n++ {
		c = c.NewChild()

		for i := 0; i < 10; i++ {
			res, err := c.ResolveWithResult(func(res *testInstance) string {
				return res.GetValue()
			})

			assert.NotNil(t, res)
			assert.NoError(t, err)

			resString := res.(string)
			if n == 0 && i == 0 {
				firstValue = resString
				continue
			}

			assert.Equal(t, firstValue, resString)
		}
	}
}

func TestResolveTransientResolvesNewInstanceFromParent(t *testing.T) {

	const nestLevel int = 10

	// Arrange
	cb := camogo.NewBuilder()
	counter := 0
	err := cb.RegisterFactory(func() *testInstance {
		counter++
		return &testInstance{strconv.Itoa(counter)}
	},
		camogo.TransientLifetime)
	assert.NoError(t, err)

	cr := cb.Build()
	c := cr.NewChild()

	// Act / Assert
	for n := 0; n < nestLevel; n++ {
		c = c.NewChild()

		var lastValue string
		for i := 0; i < 10; i++ {
			res, err := c.ResolveWithResult(func(res *testInstance) string {
				return res.GetValue()
			})

			assert.NotNil(t, res)
			assert.NoError(t, err)

			resString := res.(string)
			assert.NotEqual(t, lastValue, resString)

			lastValue = resString
		}
	}
}

func TestResolveMatchingSingletonResolvesSameInstance(t *testing.T) {

	// Arrange
	var err error
	cb := camogo.NewBuilder()
	err = cb.RegisterFactory(func() *testInstance {
		return &testInstance{strconv.Itoa(rand.Int())}
	},
		camogo.SingletonLifetime)
	assert.NoError(t, err)

	c := cb.Build()

	// Act / Assert
	var firstValue string
	for i := 0; i < 10; i++ {
		svc, err := c.ResolveMatchingType(func(typ reflect.Type) bool {
			return typ.Name() == testInstanceName()
		})

		assert.NotNil(t, svc)
		assert.NoError(t, err)

		svcInstance := svc.(*testInstance)
		resString := svcInstance.GetValue()

		if i == 0 {
			firstValue = resString
			continue
		}

		assert.Equal(t, firstValue, resString)
	}

}

func TestResolveMatchingTransientResolvesNewInstance(t *testing.T) {

	// Arrange
	cb := camogo.NewBuilder()
	counter := 0
	err := cb.RegisterFactory(func() *testInstance {
		counter++
		return &testInstance{strconv.Itoa(counter)}
	},
		camogo.TransientLifetime)
	assert.NoError(t, err)

	c := cb.Build()

	// Act / Assert
	var lastValue string
	for i := 0; i < 10; i++ {
		svc, err := c.ResolveMatchingType(func(typ reflect.Type) bool {
			return typ.Name() == testInstanceName()
		})

		assert.NotNil(t, svc)
		assert.NoError(t, err)

		svcInstance := svc.(*testInstance)
		resString := svcInstance.GetValue()
		assert.NotEqual(t, lastValue, resString)

		lastValue = resString
	}
}

func TestResolveMatchingSingletonResolvesSameInstanceFromParent(t *testing.T) {

	const nestLevel int = 10

	// Arrange
	var err error
	cb := camogo.NewBuilder()
	err = cb.RegisterFactory(func() *testInstance {
		return &testInstance{strconv.Itoa(rand.Int())}
	},
		camogo.SingletonLifetime)
	assert.NoError(t, err)

	cr := cb.Build()
	c := cr.NewChild()

	// Act / Assert
	var firstValue string
	for n := 0; n < nestLevel; n++ {
		c = c.NewChild()

		for i := 0; i < 10; i++ {
			svc, err := c.ResolveMatchingType(func(typ reflect.Type) bool {
				return typ.Name() == testInstanceName()
			})

			assert.NotNil(t, svc)
			assert.NoError(t, err)

			svcInstance := svc.(*testInstance)
			resString := svcInstance.GetValue()
			if n == 0 && i == 0 {
				firstValue = resString
				continue
			}

			assert.Equal(t, firstValue, resString)
		}
	}

}

func TestResolveMatchingTransientResolvesNewInstanceFromParent(t *testing.T) {

	const nestLevel int = 10

	// Arrange
	cb := camogo.NewBuilder()
	counter := 0
	err := cb.RegisterFactory(func() *testInstance {
		counter++
		return &testInstance{strconv.Itoa(counter)}
	},
		camogo.TransientLifetime)
	assert.NoError(t, err)

	cr := cb.Build()
	c := cr.NewChild()

	// Act / Assert
	for n := 0; n < nestLevel; n++ {
		c = c.NewChild()

		var lastValue string
		for i := 0; i < 10; i++ {
			svc, err := c.ResolveMatchingType(func(typ reflect.Type) bool {
				return typ.Name() == testInstanceName()
			})

			assert.NotNil(t, svc)
			assert.NoError(t, err)

			svcInstance := svc.(*testInstance)
			resString := svcInstance.GetValue()
			assert.NotEqual(t, lastValue, resString)

			lastValue = resString
		}
	}
}

func TestResolveTypeCanReturnContainer(t *testing.T) {

	// Arrange
	cb := camogo.NewBuilder()
	c := cb.Build()

	err := c.Resolve(func (ctr camogo.Container) error {
		assert.NotNil(t, ctr)
		assert.Same(t, c, ctr)
		return nil
	})

	assert.NoError(t, err)
}

func BenchmarkResolve(b *testing.B) {
	instance := &testInstance{b.Name()}

	cb := camogo.NewBuilder()
	err := cb.RegisterInstance(instance)
	assert.NoError(b, err)

	c := cb.Build()
	for i := 0; i < b.N; i++ {
		err = c.Resolve(func(instance *testInstance) {})
		assert.NoError(b, err)
	}
}

func BenchmarkResolveWithResult(b *testing.B) {
	instance := &testInstance{b.Name()}

	cb := camogo.NewBuilder()
	err := cb.RegisterInstance(instance)
	assert.NoError(b, err)

	c := cb.Build()
	for i := 0; i < b.N; i++ {
		_, err = c.ResolveWithResult(func(instance *testInstance) string {
			return instance.GetValue()
		})
		assert.NoError(b, err)
	}
}
