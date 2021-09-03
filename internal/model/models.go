// SPDX-License-Identifier: MIT

package model

import (
	"sync"

	"github.com/issue9/orm/v3/core"
)

// Models 数据模型管理
type Models struct {
	engine core.Engine
	locker sync.Mutex
	items  map[string]*core.Model
	names  map[string]struct{}
}

// NewModels 声明 Models 变量
func NewModels(e core.Engine) *Models {
	return &Models{
		engine: e,
		items:  map[string]*core.Model{},
		names:  map[string]struct{}{},
	}
}

// Clear 清除所有的 Model 缓存
func (ms *Models) Clear() {
	ms.locker.Lock()
	defer ms.locker.Unlock()

	ms.items = map[string]*core.Model{}
	ms.names = map[string]struct{}{}
}

func (ms *Models) addNames(name string) error {
	if _, found := ms.names[name]; found {
		return core.ErrConstraintExists(name)
	}

	ms.names[name] = struct{}{}

	return nil
}
