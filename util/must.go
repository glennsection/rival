package util

import (
	"reflect"
)

func Must(err error) {
	// simply panic if error exists
	if err != nil {
		panic(err)
	}
}

func MustIgnoreNotFound(err error) {
	// simply panic if error exists, and isn't "not found"
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
}

func IsNil(value interface{}) bool {
	defer func() { recover() }()
	return value == nil || reflect.ValueOf(value).IsNil()
}