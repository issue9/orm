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
func NewModels(db *sql.DB, d core.Dialect) *Models {
	return &Models{
		db:      db,
		dialect: d,
		models:  &sync.Map{},
	}
}

// Close 清除所有的 [core.Model] 缓存
func (ms *Models) Close() error {
	ms.models.Range(func(key, _ any) bool {
		ms.models.Delete(key)
		return true
	})

	if ms.db != nil { // 方便测试
		return ms.DB().Close()
	}
	return nil
}

func (ms *Models) SetVersion(v string) { ms.version = v }

func (ms *Models) DB() *sql.DB { return ms.db }

func (ms *Models) Version() string { return ms.version }
