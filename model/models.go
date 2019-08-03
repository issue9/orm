// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"sync"

	"github.com/issue9/orm/v2/core"
)

// Models 数据模型管理
type Models struct {
	engine core.Engine
	locker sync.Mutex
	items  map[reflect.Type]*core.Model
	names  map[string]struct{}
}

// NewModels 声明 Models 变量
func NewModels(e core.Engine) *Models {
	return &Models{
		engine: e,
		items:  map[reflect.Type]*core.Model{},
		names:  map[string]struct{}{},
	}
}

// Clear 清除所有的 Model 缓存。
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.items = map[reflect.Type]*core.Model{}
	ms.names = map[string]struct{}{}
}

func (ms *Models) addNames(name ...string) error {
	for _, n := range name {
		if _, found := ms.names[n]; found {
			return core.ErrConstraintExists(n)
		}

		ms.names[n] = struct{}{}
	}

	return nil
}
