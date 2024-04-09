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
	dialect core.Dialect
	models  *sync.Map
}

// NewModels 声明 [Models] 变量
func NewModels(d core.Dialect) *Models {
	return &Models{
		dialect: d,
		models:  &sync.Map{},
	}
}

// Clear 清除所有的 [core.Model] 缓存
func (ms *Models) Clear() {
	ms.models.Range(func(key, _ any) bool {
		ms.models.Delete(key)
		return true
	})
}
