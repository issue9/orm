// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"reflect"
	"sync"

	"github.com/issue9/orm/v2/core"
)

// ErrDuplicateName 诸如 postgres 等数据库，
// 需要约束名、索引名称等全局唯一。
// 如果返回此错误，说明违反了此规定。
//
// NOTE: orm 强制所有数据都需要名称全局唯一。
var ErrDuplicateName = errors.New("重复的约束名")

// Models 数据模型管理
type Models struct {
	engine core.Engine
	locker sync.Mutex
	items  map[reflect.Type]*Model
	names  map[string]struct{}
}

// NewModels 声明 Models 变量
func NewModels(e core.Engine) *Models {
	return &Models{
		engine: e,
		items:  map[reflect.Type]*Model{},
		names:  map[string]struct{}{},
	}
}

// Clear 清除所有的 Model 缓存。
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.items = map[reflect.Type]*Model{}
	ms.names = map[string]struct{}{}
}

func (ms *Models) addNames(name ...string) error {
	for _, n := range name {
		if _, found := ms.names[n]; found {
			return ErrDuplicateName
		}

		ms.names[n] = struct{}{}
	}

	return nil
}
