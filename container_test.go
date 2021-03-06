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

type TypeC struct {
	Field TypeA `inject:"type-a"`
}

type TypeD struct {
	Field int `inject:"auto"`
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
		require.EqualError(t, err, "inject: int is not assignable from string")
	})

	t.Run("not-assignable-pointer-expected", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.Register("mocked-int", 1000))
		require.NoError(t, c.Register("type-a", a))
		err := c.Register("type-c", TypeC{})
		require.EqualError(t, err, "inject: inject.TypeC is not injectable, a pointer is expected")
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

func Test_Register_auto(t *testing.T) {
	c := New()
	c.MustRegister("mocked-int", 10)
	d := &TypeD{}
	err := c.Register("type-d", d)
	require.NoError(t, err)
	require.Equal(t, 10, d.Field, "data should be injected by type")
}

func Test_Register_reserved_name(t *testing.T) {
	c := New()
	err := c.Register("auto", 10)
	require.EqualError(t, err, "inject: auto is revserved, please use a different name")
}

func Test_Unnamed(t *testing.T) {
	c := New()
	err := c.Unnamed(10)
	require.NoError(t, err, "New dependency shouldn't registered")
	require.Len(t, c.dependencies, 1)
	require.NotNil(t, c.dependencies["unnamed.0"])
	require.Equal(t, 0, c.unnamedCounter)
}

func Test_Unnamed_error(t *testing.T) {
	c := New()
	err := c.Unnamed(&TypeA{})
	require.Error(t, err, "There should be error because of missing dependency")
	require.Equal(t, 0, c.unnamedCounter)
	require.NoError(t, c.Unnamed(10))
	require.Len(t, c.dependencies, 1)
	require.Equal(t, 0, c.unnamedCounter)
	require.NotNil(t, c.dependencies["unnamed.0"], "Existing name should be reused")
	require.NoError(t, c.Unnamed(11))
	require.NotNil(t, c.dependencies["unnamed.1"], "New name should be created")
}

func Test_MustUnnamed(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		c.MustUnnamed(&TypeA{})
	}, "There must be panic because of missing dependency")
}

func Test_Inject(t *testing.T) {
	c := New()
	require.NoError(t, c.Unnamed(10))

	d := &TypeD{}
	require.NoError(t, c.Inject(d), "there should be no error injecting dependencies")
	require.Equal(t, 10, d.Field, "data should be injected")
}
