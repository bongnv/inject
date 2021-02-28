package inject

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type TypeA struct {
	Field int `inject:"mocked-int"`
}

type TypeB struct {
	Field *TypeA `inject:"type-a"`
}

func Test_Register(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		b := &TypeB{}
		require.NoError(t, c.Register("mocked-int", 10))
		require.NoError(t, c.Register("type-a", a))
		require.NoError(t, c.Register("type-b", b))
		require.EqualValues(t, 10, a.Field)
		require.EqualValues(t, a, b.Field)
	})

	t.Run("not-assignable", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.Register("mocked-int", "1000"))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: mocked-int is not assignable")
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		b := &TypeB{}
		require.NoError(t, c.Register("mocked-int", "1000"))
		err := c.Register("type-b", b)
		require.EqualError(t, err, "inject: type-a is not registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.Register("type-a", 10))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: type-a is already registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.Register("type-a", 10))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: type-a is already registered")
	})
}

func Test_Get(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.Register("mocked-int", 10))
		require.NoError(t, c.Register("type-a", a))
		retrievedA, err := c.Get("type-a")
		require.NoError(t, err)
		require.IsType(t, &TypeA{}, retrievedA)
		require.EqualValues(t, 10, retrievedA.(*TypeA).Field)
	})

	t.Run("not-found", func(t *testing.T) {
		c := New()
		dep, err := c.Get("some-dep")
		require.EqualError(t, err, "inject: the requested dependency couldn't be found")
		require.Nil(t, dep)
	})
}

func Test_MustRegister_panic(t *testing.T) {
	c := New()
	require.NoError(t, c.Register("mock-int", 10))
	require.Panics(t, func() {
		c.MustRegister("mock-int", 20)
	})
}

func Test_MustGet_panic(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		c.MustGet("some-dep")
	}, "it must panic as there is no request dep")
}

func Test_MustGet_no_panic(t *testing.T) {
	c := New()
	c.MustRegister("mock-int", 10)
	require.NotPanics(t, func() {
		require.EqualValues(t, 10, c.MustGet("mock-int"))
	}, "it must panic as there is no request dep")
}

func Test_Register_factory_function(t *testing.T) {
	t.Run("too-many-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, int, error) {
			return 0, 0, nil
		}

		err := c.Register("new-func", mockFunc)
		require.EqualError(t, err, "inject: unsupported factory function")
	})

	t.Run("no-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() {}

		err := c.Register("new-func", mockFunc)
		require.EqualError(t, err, "inject: unsupported factory function")
	})

	t.Run("not-implements-error", func(t *testing.T) {
		c := New()
		mockFunc := func() (string, int) {
			return "", 0
		}

		err := c.Register("new-func", mockFunc)
		require.EqualError(t, err, "inject: 2nd output param must implement error")
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		err := c.Register("int-dep", mockFunc)
		require.EqualError(t, err, "inject: couldn't find the dependency for string")
	})

	t.Run("conflict-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		c.MustRegister("string-dep-1", "dep-1")
		c.MustRegister("string-dep-2", "dep-2")
		err := c.Register("int-dep", mockFunc)
		require.EqualError(t, err, "inject: there is a conflict when finding the dependency for string")
	})

	t.Run("error-with-factory-func", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, errors.New("random error")
		}

		c.MustRegister("string-dep-1", "dep-1")
		err := c.Register("int-dep", mockFunc)
		require.EqualError(t, err, "random error")
	})

	t.Run("happy-path-with-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 1, nil
		}

		c.MustRegister("string-dep-1", "dep-1")
		err := c.Register("int-dep", mockFunc)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.EqualValues(t, 1, c.MustGet("int-dep"))
		}, "int-dep must be registered")
	})

	t.Run("happy-path-without-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, error) {
			return 1, nil
		}

		err := c.Register("int-dep", mockFunc)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.EqualValues(t, 1, c.MustGet("int-dep"))
		}, "int-dep must be registered")
	})
}
