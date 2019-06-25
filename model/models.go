// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"reflect"
	"sync"
)

// ErrDuplicateConstraintName 重复的约束名
//
// 部分数据需要约束名全局唯一，Models 为了通用性，强制要求全局唯一，
// 在出现约束名重复时，会返回此错误信息。
//
// NOTE: 索引名称也会被作约束名称处理。
var ErrDuplicateConstraintName = errors.New("已经存在相同的约束名")

// Models 数据模型管理
type Models struct {
	locker sync.Mutex
	items  map[reflect.Type]*Model
	names  map[string]struct{}
}

// 保存约束名
func (ms Models) addNames(name string) error {
	for key := range ms.names {
		if key == name {
			return ErrDuplicateConstraintName
		}
	}

	ms.names[name] = struct{}{}

	return nil
}

// NewModels 声明 Models 变量
func NewModels() *Models {
	return &Models{
		items: map[reflect.Type]*Model{},
		names: map[string]struct{}{},
	}
}

// Clear 清除所有的 Model 缓存。
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.items = map[reflect.Type]*Model{}
}
