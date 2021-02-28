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

	var toAddDep *dependency
	depType := reflect.TypeOf(dep)
	if depType.Kind() == reflect.Func {
		createdDep, err := c.executeFunc(dep, depType)
		if err != nil {
			return err
		}

		toAddDep = createdDep
	} else {
		toAddDep = &dependency{
			value:        dep,
			reflectType:  depType,
			reflectValue: reflect.ValueOf(dep),
		}
	}

	if err := c.populate(toAddDep); err != nil {
		return err
	}

	c.dependencies[name] = toAddDep
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

func (c *Container) executeFunc(fn interface{}, fnType reflect.Type) (*dependency, error) {
	if fnType.NumOut() > 2 || fnType.NumOut() < 1 {
		return nil, errors.New("inject: unsupported factory function")
	}

	if fnType.NumOut() == 2 && !implementsError(fnType.Out(1)) {
		return nil, errors.New("inject: 2nd output param must implement error")
	}

	fnVal := reflect.ValueOf(fn)
	inParams, err := c.generateInParams(fnType)
	if err != nil {
		return nil, err
	}

	out := fnVal.Call(inParams)
	if len(out) == 2 && !out[1].IsNil() {
		return nil, out[1].Interface().(error)
	}

	newDep := &dependency{
		value:        out[0].Interface(),
		reflectValue: out[0],
		reflectType:  out[0].Type(),
	}

	return newDep, nil
}

func (c *Container) generateInParams(fnType reflect.Type) ([]reflect.Value, error) {
	params := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		param, err := c.findByType(fnType.In(i))
		if err != nil {
			return nil, err
		}

		params[i] = param.reflectValue
	}

	return params, nil
}

func (c *Container) findByType(t reflect.Type) (*dependency, error) {
	var foundVal *dependency
	for _, v := range c.dependencies {
		if v.reflectType.AssignableTo(t) {
			if foundVal != nil {
				return nil, fmt.Errorf("inject: there is a conflict when finding the dependency for %s", t.String())
			}

			foundVal = v
		}
	}

	if foundVal == nil {
		return nil, fmt.Errorf("inject: couldn't find the dependency for %s", t.String())
	}

	return foundVal, nil
}

type dependency struct {
	value        interface{}
	reflectValue reflect.Value
	reflectType  reflect.Type
}
