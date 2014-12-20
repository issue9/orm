// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"sync"
)

type engineMap struct {
	sync.Mutex
	items map[string]*Engine
}

var engines = engineMap{items: make(map[string]*Engine)}

// New 声明一个新的Engine实例。
func New(driverName, dataSourceName, engineName, prefix string) (*Engine, error) {
	engines.Lock()
	defer engines.Unlock()

	if _, found := engines.items[engineName]; found {
		return nil, fmt.Errorf("该名称[%v]的Engine已经存在", engineName)
	}

	e, err := newEngine(driverName, dataSourceName, prefix)
	if err != nil {
		return nil, err
	}

	engines.items[engineName] = e

	return e, nil
}

// 获取指定名称的Engine，若不存在则found值返回false。
func Get(engineName string) (e *Engine, found bool) {
	engines.Lock()
	defer engines.Unlock()

	e, found = engines.items[engineName]
	return
}

// 关闭指定名称的Engine
func Close(engineName string) {
	engines.Lock()
	defer engines.Unlock()

	e, found := engines.items[engineName]
	if !found {
		return
	}

	e.close()
	delete(engines.items, engineName)
}

// 关闭所有的Engine
func CloseAll() {
	engines.Lock()
	defer engines.Unlock()

	for _, v := range engines.items {
		v.close()
	}

	// 重新声明一块内存，而不是直接赋值nil
	// 防止在CloseAll()之后，再New()新的Engine
	engines.items = make(map[string]*Engine)
}
