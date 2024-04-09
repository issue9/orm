// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model

import "sync"

// Models 数据模型管理
type Models struct {
	models *sync.Map
}

// NewModels 声明 [Models] 变量
func NewModels() *Models {
	return &Models{
		models: &sync.Map{},
	}
}

// Clear 清除所有的 Model 缓存
func (ms *Models) Clear() {
	ms.models.Range(func(key, _ any) bool {
		ms.models.Delete(key)
		return true
	})
}
