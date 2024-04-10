// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model

import (
	"database/sql"
	"sync"

	"github.com/issue9/orm/v6/core"
)

// Models 数据模型管理
type Models struct {
	db      *sql.DB
	dialect core.Dialect
	models  *sync.Map
	version string
}

// NewModels 声明 [Models] 变量
//
// 返回对象中除了 [Models] 之外，还包含了一个 core.Engine 对象，
// 该对象的表名前缀由参数 tablePrefix 指定。
func NewModels(db *sql.DB, d core.Dialect, tablePrefix string) (*Models, core.Engine, error) {
	ms := &Models{
		db:      db,
		dialect: d,
		models:  &sync.Map{},
	}

	e := ms.NewEngine(db, tablePrefix)
	if err := e.QueryRow(d.VersionSQL()).Scan(&ms.version); err != nil {
		return nil, nil, err
	}
	return ms, e, nil
}

// Close 清除所有的 [core.Model] 缓存
func (ms *Models) Close() error {
	ms.models.Range(func(key, _ any) bool {
		ms.models.Delete(key)
		return true
	})

	return ms.DB().Close()
}

func (ms *Models) DB() *sql.DB { return ms.db }

func (ms *Models) Version() string { return ms.version }

func (ms *Models) Length() (cnt int) {
	ms.models.Range(func(key, value any) bool {
		cnt++
		return true
	})
	return
}
