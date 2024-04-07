// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model

import (
	"sync"

	"github.com/issue9/orm/v6/core"
)

// Models 数据模型管理
type Models struct {
	engine core.Engine
	models *sync.Map
}

// NewModels 声明 [Models] 变量
func NewModels(e core.Engine) *Models {
	return &Models{
		engine: e,
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
