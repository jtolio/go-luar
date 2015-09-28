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
	"unicode"
	"unicode/utf8"

	"github.com/Shopify/go-lua"
)

type Options struct {
	AllowUnexportedAccess bool
}

func SetOptions(l *lua.State, opts Options) {
	l.PushUserData(opts)
	l.SetField(lua.RegistryIndex, optsName)
}

func GetOptions(l *lua.State) Options {
	l.Field(lua.RegistryIndex, optsName)
	opts, ok := l.ToUserData(-1).(Options)
	l.Pop(1)
	if !ok {
		return Options{}
	}
	return opts
}

func (o Options) canAccess(fieldName string) bool {
	if o.AllowUnexportedAccess {
		return true
	}
	buf := []byte(fieldName)
	first, n := utf8.DecodeRune(buf)
	if n == 0 {
		return true
	}
	return string(unicode.ToUpper(first)) == string(buf[:n])
}
