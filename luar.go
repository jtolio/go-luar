// Copyright 2015 JT Olds
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package luar

import (
	"fmt"
	"reflect"

	"github.com/Shopify/go-lua"
)

// PushValue pushes a Go value mapped to the appropriate Lua binding onto the
// Lua stack. Usually used like:
// if err := PushValue(l, x); err == nil {
//   l.SetGlobal("x")
// }
// Will return an error if the conversion is not possible.
func PushValue(l *lua.State, val interface{}) error {
	if val == nil {
		l.PushNil()
		return nil
	}
	return PushReflectedValue(l, reflect.ValueOf(val))
}

// PushReflectedValue is like PushValue, but works on already reflected values.
func PushReflectedValue(l *lua.State, val reflect.Value) (err error) {
	switch val.Kind() {
	case reflect.Bool:
		l.PushBoolean(val.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		l.PushNumber(float64(val.Int()))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		l.PushNumber(float64(val.Uint()))
		return nil
	case reflect.Float32, reflect.Float64:
		l.PushNumber(val.Float())
		return nil
	case reflect.String:
		l.PushString(val.String())
		return nil

	case reflect.Struct:
		l.PushUserData(val)
		pushAndSetupStructTable(l)
		l.SetMetaTable(-2)

		return nil

	case reflect.Ptr:
		if val.IsNil() {
			l.PushNil()
			return nil
		}

		l.PushUserData(val)
		pushAndSetupPtrTable(l)
		l.SetMetaTable(-2)

		return nil

	case reflect.Func:
		if val.IsNil() {
			l.PushNil()
			return nil
		}

		l.PushUserData(val)
		pushAndSetupFuncTable(l)
		l.SetMetaTable(-2)

		return nil

	case reflect.Interface:
		if val.IsNil() {
			l.PushNil()
			return nil
		}

		l.PushUserData(val)
		pushAndSetupInterfaceTable(l)
		l.SetMetaTable(-2)

		return nil

	case reflect.Array: // TODO
	case reflect.Slice: // TODO
	case reflect.Chan: // TODO
	case reflect.Map: // TODO

	case reflect.Complex64, reflect.Complex128:
	case reflect.UnsafePointer:
	}

	return fmt.Errorf("unsupported value type: %v", val)
}

func pushReflectedValue(l *lua.State, val reflect.Value) {
	err := PushReflectedValue(l, val)
	if err != nil {
		lua.Errorf(l, "%s", err.Error())
		panic("unreachable")
	}
}

// PushType pushes a constructor for the given example's type onto the Lua
// stack. Usually used like:
// if err := PushType(l, Type{}); err == nil {
//   l.SetGlobal("Type")
// }
func PushType(l *lua.State, example interface{}) error {
	setupType(l)
	l.PushUserData(reflect.TypeOf(example))
	lua.SetMetaTableNamed(l, typeName)
	return nil
}

func toReflectedValue(l *lua.State, index int,
	hint reflect.Type) reflect.Value {
	val, err := ToReflectedValue(l, index, hint)
	if err != nil {
		lua.Errorf(l, "%s", err.Error())
		panic("unreachable")
	}
	return val
}

// ToReflectedValue is like ToValue, but leaves the type as a reflect.Value.
// If hint is not nil, will use hint if possible to try and better inform the
// conversion (if hint is a slice and the lua value is a table, will simply
// get the numeric indexes).
func ToReflectedValue(l *lua.State, index int, hint reflect.Type) (
	result reflect.Value, err error) {
	result, err = func() (reflect.Value, error) {
		switch l.TypeOf(index) {
		case lua.TypeNil:
			return reflect.ValueOf(nil), nil
		case lua.TypeBoolean:
			return reflect.ValueOf(l.ToBoolean(index)), nil
		case lua.TypeNumber:
			val, ok := l.ToNumber(index)
			if !ok {
				return reflect.Value{}, fmt.Errorf("unable to cast to number")
			}
			return reflect.ValueOf(val), nil
		case lua.TypeString:
			val, ok := l.ToString(index)
			if !ok {
				return reflect.Value{}, fmt.Errorf("unable to cast to string")
			}
			return reflect.ValueOf(val), nil

		case lua.TypeUserData:
			ud, ok := l.ToUserData(index).(reflect.Value)
			if !ok {
				return reflect.Value{}, fmt.Errorf("unable to cast type")
			}
			return ud, nil

		case lua.TypeTable:
			if hint == nil {
				return reflect.Value{},
					fmt.Errorf("no hint provided for lua table conversion")
			}
			switch hint.Kind() {
			case reflect.Array: // TODO
				return reflect.Value{}, fmt.Errorf("TODO: table to array conversion")
			case reflect.Slice: // TODO
				return reflect.Value{}, fmt.Errorf("TODO: table to slice conversion")
			case reflect.Map: // TODO
				return reflect.Value{}, fmt.Errorf("TODO: table to map conversion")
			default:
				return reflect.Value{},
					fmt.Errorf("not enough information exists to know how to " +
						"convert lua table to go object")
			}
		case lua.TypeLightUserData:
		case lua.TypeFunction:
		case lua.TypeThread:
		}
		return reflect.Value{}, fmt.Errorf(
			"unable to cast value to appropriate Go type")
	}()

	if err != nil || hint == nil {
		return result, err
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return result.Convert(hint), nil
}

// ToValue returns the reverse-mapped Go value from the Lua stack at index
// `index`, and an error if the conversion isn't yet possible.
// If hint is not nil, will use hint if possible to try and better inform the
// conversion (if hint is a slice and the lua value is a table, will simply
// get the numeric indexes).
func ToValue(l *lua.State, index int, hint reflect.Type) (interface{}, error) {
	val, err := ToReflectedValue(l, index, hint)
	if err != nil {
		return nil, err
	}
	if !val.CanInterface() {
		return nil, fmt.Errorf("value cannot be read")
	}
	return val.Interface(), nil
}

func fixPanics() {
	r := recover()
	if r == nil {
		return
	}
	_, ok := r.(error)
	if !ok {
		panic(fmt.Errorf("panic: %v", r))
	}
	panic(r)
}
