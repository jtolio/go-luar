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

func checkInterface(l *lua.State, index int) reflect.Value {
	ref, ok := lua.CheckUserData(l, index, interfaceName).(reflect.Value)
	if !ok || ref.Kind() != reflect.Interface {
		lua.ArgumentError(l, 1, "interface expected")
		panic("unreached")
	}
	return ref
}

func setupInterface(l *lua.State) {
	if !lua.NewMetaTable(l, interfaceName) {
		return
	}
	lua.SetFunctions(l, []lua.RegistryFunction{
		{"__index", func(l *lua.State) int {
			defer fixPanics()
			return getField(l, checkInterface(l, 1), 2)
		}},
		{"__tostring", func(l *lua.State) int {
			defer fixPanics()
			l.PushString(fmt.Sprintf("%#v", checkInterface(l, 1).Interface()))
			return 1
		}},
		{"__eq", func(l *lua.State) int {
			defer fixPanics()
			l.PushBoolean(checkInterface(l, 1) == checkInterface(l, 2))
			return 1
		}},
	}, 0)
}
