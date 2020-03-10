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

func checkStruct(l *lua.State, index int) reflect.Value {
	ref, ok := lua.CheckUserData(l, index, structName).(reflect.Value)
	if !ok || ref.Kind() != reflect.Struct {
		lua.ArgumentError(l, 1, "struct expected")
		panic("unreached")
	}
	return ref
}

func getField(l *lua.State, ref reflect.Value, name_idx int) int {
	name := lua.CheckString(l, name_idx)
	if !GetOptions(l).canAccess(name) {
		lua.ArgumentError(l, name_idx, fmt.Sprintf("field %#v missing", name))
		panic("unreached")
	}
	val := ref.MethodByName(name)
	if !val.IsValid() {
		if ref.Kind() == reflect.Ptr {
			ref = ref.Elem()
		}
		if ref.Kind() != reflect.Interface {
			val = ref.FieldByName(name)
		}
		if !val.IsValid() {
			lua.ArgumentError(l, name_idx, fmt.Sprintf("field %#v missing", name))
			panic("unreached")
		}
	}
	pushReflectedValue(l, val)
	return 1
}

func setField(l *lua.State, ref reflect.Value, name_idx int, val_idx int) int {
	name := lua.CheckString(l, name_idx)
	lua.CheckAny(l, val_idx)
	if !GetOptions(l).canAccess(name) {
		lua.ArgumentError(l, name_idx, fmt.Sprintf("field %#v missing", name))
		panic("unreached")
	}
	var lhs reflect.Value
	if ref.Kind() == reflect.Ptr {
		lhs = ref.Elem().FieldByName(name)
	} else {
		lhs = ref.FieldByName(name)
	}
	if !lhs.IsValid() {
		lua.ArgumentError(l, name_idx, fmt.Sprintf("field %#v missing", name))
		panic("unreached")
	}
	if !lhs.CanSet() {
		lua.ArgumentError(l, name_idx, fmt.Sprintf("can't set field %#v", name))
		panic("unreached")
	}
	hint := lhs.Type()
	lhs.Set(toReflectedValue(l, val_idx, hint))
	return 0
}

func pushAndSetupStructTable(l *lua.State) {
	if !lua.NewMetaTable(l, structName) {
		return
	}
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: func(l *lua.State) int {
			defer fixPanics()
			return getField(l, checkStruct(l, 1), 2)
		}},
		{Name: "__newindex", Function: func(l *lua.State) int {
			defer fixPanics()
			return setField(l, checkStruct(l, 1), 2, 3)
		}},
		{Name: "__tostring", Function: func(l *lua.State) int {
			defer fixPanics()
			l.PushString(fmt.Sprintf("%#v", checkStruct(l, 1).Interface()))
			return 1
		}},
		{Name: "__eq", Function: func(l *lua.State) int {
			defer fixPanics()
			l.PushBoolean(checkStruct(l, 1) == checkStruct(l, 2))
			return 1
		}},
	}, 0)
}
