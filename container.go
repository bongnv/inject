package inject

import (
	"errors"
	"fmt"
	"reflect"
)

// Container contains all dependencies.
type Container struct {
	dependencies map[string]*dependency
}

// Register registers new dependency with a name to the Container. name has to be unique.
// It returns an error if name is not unique or it's unable to inject dependencies.
func (c *Container) Register(name string, dep interface{}) error {
	if _, found := c.dependencies[name]; found {
		return fmt.Errorf("inject: %s is already registered", name)
	}

	newDep := &dependency{
		value:        dep,
		reflectType:  reflect.TypeOf(dep),
		reflectValue: reflect.ValueOf(dep),
	}

	if err := c.populate(newDep); err != nil {
		return err
	}

	c.dependencies[name] = newDep
	return nil
}

// MustRegister is the similar to Register. Instead of returning an error,
// it panics if anything goes wrong.
func (c *Container) MustRegister(name string, dep interface{}) {
	if err := c.Register(name, dep); err != nil {
		panic(err)
	}
}

// Get loads a dependency from the Container using name.
// It returns an error if the requested dependency couldn't be found.
func (c *Container) Get(name string) (interface{}, error) {
	dep, found := c.dependencies[name]
	if !found {
		return nil, errors.New("inject: the requested dependency couldn't be found")
	}

	return dep.value, nil
}

// MustGet is the similar to Get. Instead of returning an error,
// it panics if anything goes wrong.
func (c *Container) MustGet(name string) interface{} {
	dep, err := c.Get(name)
	if err != nil {
		panic(err)
	}

	return dep
}

type dependency struct {
	value        interface{}
	reflectValue reflect.Value
	reflectType  reflect.Type
}

func (c *Container) populate(dep *dependency) error {
	if !isStructPtr(dep.reflectType) {
		return nil
	}

	for i := 0; i < dep.reflectValue.Elem().NumField(); i++ {
		fieldValue := dep.reflectValue.Elem().Field(i)
		fieldType := fieldValue.Type()
		structField := dep.reflectType.Elem().Field(i)
		fieldTag := structField.Tag
		tagValue, ok := fieldTag.Lookup("inject")
		if !ok {
			continue
		}

		loadedValue, found := c.dependencies[tagValue]
		if !found {
			return fmt.Errorf("inject: %s is not registered", tagValue)
		}

		if !loadedValue.reflectType.AssignableTo(fieldType) {
			return fmt.Errorf("inject: %s is not assignable", tagValue)
		}

		fieldValue.Set(loadedValue.reflectValue)
	}

	return nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
