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

func checkType(l *lua.State, index int) reflect.Type {
	ref, ok := lua.CheckUserData(l, index, typeName).(reflect.Type)
	if !ok {
		lua.ArgumentError(l, 1, "type expected")
		panic("unreached")
	}
	return ref
}

func pushType(l *lua.State, ref reflect.Type, zero bool) {
	switch ref.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice:
		lua.ArgumentError(l, 1, fmt.Sprintf("type unsupported: %v", ref))
		panic("unreached")
	default:
	}
	if zero {
		pushReflectedValue(l, reflect.Zero(ref))
	} else {
		pushReflectedValue(l, reflect.New(ref))
	}
}

func setupType(l *lua.State) {
	if !lua.NewMetaTable(l, typeName) {
		return
	}
	lua.SetFunctions(l, []lua.RegistryFunction{
		{"__call", func(l *lua.State) int {
			defer fixPanics()
			pushType(l, checkType(l, 1), true)
			return 1
		}},
		{"__index", func(l *lua.State) int {
			defer fixPanics()
			ref := checkType(l, 1)
			if lua.CheckString(l, 2) != "new" {
				l.PushNil()
				return 1
			}
			l.PushGoFunction(func(l *lua.State) int {
				defer fixPanics()
				pushType(l, ref, false)
				return 1
			})
			return 1
		}},
		{"__tostring", func(l *lua.State) int {
			defer fixPanics()
			l.PushString(fmt.Sprintf("Go Type: %v", checkType(l, 1)))
			return 1
		}},
		{"__eq", func(l *lua.State) int {
			defer fixPanics()
			l.PushBoolean(checkType(l, 1) == checkType(l, 2))
			return 1
		}},
	}, 0)
}
