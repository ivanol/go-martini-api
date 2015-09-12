package api

import (
	"reflect"
)

//Reflection utilities

//getReflectedSlicePtr takes a Type such as User, and returns an
//interface containing &MakeSlice([]User, 0, 0).
func getReflectedSlicePtr(sliceType reflect.Type) interface{} {
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(slice.Type())
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}
