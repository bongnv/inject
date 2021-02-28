package inject

import "reflect"

var (
	reflectTypeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func implementsError(t reflect.Type) bool {
	return t.Implements(reflectTypeOfError)
}
