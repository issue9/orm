// SPDX-License-Identifier: MIT

package model

import (
	"sync"

	"github.com/issue9/orm/v4/core"
)

// Models 数据模型管理
type Models struct {
	engine core.Engine
	locker sync.Mutex
	models map[string]*core.Model
}

// NewModels 声明 Models 变量
func NewModels(e core.Engine) *Models {
	return &Models{
		engine: e,
		models: map[string]*core.Model{},
	}
}

// Clear 清除所有的 Model 缓存
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.models = map[string]*core.Model{}
}
