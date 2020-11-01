package inject

import (
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
		c := &Container{}
		a := &TypeA{}
		b := &TypeB{}
		require.NoError(t, c.Register("mocked-int", 10))
		require.NoError(t, c.Register("type-a", a))
		require.NoError(t, c.Register("type-b", b))
		require.EqualValues(t, 10, a.Field)
		require.EqualValues(t, a, b.Field)
	})

	t.Run("not-assignable", func(t *testing.T) {
		c := &Container{}
		a := &TypeA{}
		require.NoError(t, c.Register("mocked-int", "1000"))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: mocked-int is not assignable")
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := &Container{}
		b := &TypeB{}
		require.NoError(t, c.Register("mocked-int", "1000"))
		err := c.Register("type-b", b)
		require.EqualError(t, err, "inject: type-a is not registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := &Container{}
		a := &TypeA{}
		require.NoError(t, c.Register("type-a", 10))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: type-a is already registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := &Container{}
		a := &TypeA{}
		require.NoError(t, c.Register("type-a", 10))
		err := c.Register("type-a", a)
		require.EqualError(t, err, "inject: type-a is already registered")
	})
}

func Test_Get(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := &Container{}
		a := &TypeA{}
		require.NoError(t, c.Register("mocked-int", 10))
		require.NoError(t, c.Register("type-a", a))
		retrievedA, err := c.Get("type-a")
		require.NoError(t, err)
		require.IsType(t, &TypeA{}, retrievedA)
		require.EqualValues(t, 10, retrievedA.(*TypeA).Field)
	})

	t.Run("not-found", func(t *testing.T) {
		c := &Container{}
		dep, err := c.Get("some-dep")
		require.EqualError(t, err, "inject: the requested dependency couldn't be found")
		require.Nil(t, dep)
	})
}
