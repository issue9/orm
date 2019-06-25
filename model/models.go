// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"sync"
)

// Models 数据模型管理
type Models struct {
	locker sync.Mutex
	items  map[reflect.Type]*Model
}

// NewModels 声明 Models 变量
func NewModels() *Models {
	return &Models{
		items: map[reflect.Type]*Model{},
	}
}

// Clear 清除所有的 Model 缓存。
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.items = map[reflect.Type]*Model{}
}
