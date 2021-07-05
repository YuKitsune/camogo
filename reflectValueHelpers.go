package camogo

import "reflect"

func valueOrNil(value reflect.Value) interface{} {
	if value.IsNil() {
		return nil
	} else {
		return value.Interface()
	}
}

func errorOrNil(value reflect.Value) error {
	err := valueOrNil(value)
	if err == nil {
		return nil
	}

	return err.(error)
}