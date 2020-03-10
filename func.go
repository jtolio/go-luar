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

var (
	refTypeStatePtr = reflect.TypeOf(&lua.State{})
	refTypeInt      = reflect.TypeOf(int(0))
)

func checkFunc(l *lua.State, index int) reflect.Value {
	ref, ok := lua.CheckUserData(l, index, funcName).(reflect.Value)
	if !ok || ref.Kind() != reflect.Func {
		lua.ArgumentError(l, 1, "function expected")
		panic("unreached")
	}
	return ref
}

func pushAndSetupFuncTable(l *lua.State) {
	if !lua.NewMetaTable(l, funcName) {
		return
	}
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__call", Function: func(l *lua.State) int {
			defer fixPanics()
			f := checkFunc(l, 1)
			ft := f.Type()
			if ft.NumIn() == 1 && ft.NumOut() == 1 &&
				ft.In(0) == refTypeStatePtr && ft.Out(0) == refTypeInt {
				return int(f.Call([]reflect.Value{reflect.ValueOf(l)})[0].Int())
			}
			args := make([]reflect.Value, l.Top()-1)
			expected := ft.NumIn()
			variadic := ft.IsVariadic()
			if !variadic {
				if len(args) != expected {
					lua.Errorf(l, "%s", fmt.Sprintf(
						"wrong number of arguments: got %d, expected %d",
						len(args), expected))
					panic("unreached")
				}
			} else {
				if len(args) < expected-1 {
					lua.Errorf(l, "%s", fmt.Sprintf(
						"wrong number of arguments: got %d, expected %d or more",
						len(args), expected-1))
					panic("unreached")
				}
			}
			for i := 0; i < len(args); i++ {
				var hint reflect.Type
				if !variadic || i < expected-1 {
					hint = ft.In(i)
				} else {
					hint = ft.In(expected - 1).Elem()
				}
				args[i] = toReflectedValue(l, i+2, hint)
			}
			rv := f.Call(args)
			if !l.CheckStack(len(rv)) {
				lua.Errorf(l, "failed to increase stack size")
				panic("unreached")
			}
			for _, val := range rv {
				pushReflectedValue(l, val)
			}
			return len(rv)
		}},
		{Name: "__tostring", Function: func(l *lua.State) int {
			defer fixPanics()
			l.PushString(fmt.Sprintf("%#v", checkFunc(l, 1).Interface()))
			return 1
		}},
		{Name: "__eq", Function: func(l *lua.State) int {
			defer fixPanics()
			l.PushBoolean(checkFunc(l, 1) == checkFunc(l, 2))
			return 1
		}},
	}, 0)
}
