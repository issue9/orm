// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/dialect"
)

// 实现两个internal.DB接口，分别对应sql包的DB和Tx结构，
// 供SQL和model包使用

type Engine struct {
	name   string // 数据库的名称
	prefix string // 表名前缀
	d      core.Dialect
	db     *sql.DB
	stmts  *core.Stmts
	sql    *SQL // 内置的SQL引擎，用于执行Update等操作
}

func newEngine(driverName, dataSourceName, prefix string) (*Engine, error) {
	d, found := dialect.Get(driverName)
	if !found {
		return nil, fmt.Errorf("未找到与driverName[%v]相同的Dialect", driverName)
	}

	dbInst, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	inst := &Engine{
		db:     dbInst,
		d:      d,
		prefix: prefix,
		name:   d.GetDBName(dataSourceName),
	}
	inst.stmts = core.NewStmts(inst)
	inst.sql = inst.SQL()

	return inst, nil
}

// 对orm/core.DB.Name()的实现，返回当前操作的数据库名称。
func (e *Engine) Name() string {
	return e.name
}

// 对orm/core.DB.GetStmts()的实现，返回当前的sql.Stmt实例缓存容器。
func (e *Engine) GetStmts() *core.Stmts {
	return e.stmts
}

// 对orm/core.DB.PrepareSQL()的实现。替换语句的各种占位符。
func (e *Engine) PrepareSQL(sql string) string {
	// TODO 缓存replace
	l, r := e.Dialect().QuoteStr()
	replace := strings.NewReplacer("{", l, "}", r, "#", e.prefix)

	return replace.Replace(sql)
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (e *Engine) Dialect() core.Dialect {
	return e.d
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (e *Engine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return e.db.Exec(sql, args...)
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (e *Engine) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return e.db.Query(sql, args...)
}

// 对orm/core.DB.QueryRow()的实现。
// 执行一条查询语句，并返回第一条符合条件的记录。
func (e *Engine) QueryRow(sql string, args ...interface{}) *sql.Row {
	return e.db.QueryRow(sql, args...)
}

// 对orm/core.DB.Prepare()的实现。预处理SQL语句成sql.Stmt实例。
func (e *Engine) Prepare(sql string) (*sql.Stmt, error) {
	return e.db.Prepare(sql)
}

// 关闭当前的db，销毁所有的数据。不能再次使用。
func (e *Engine) close() {
	e.stmts.Close()
	e.db.Close()
}

// 开始一个新的事务
func (e *Engine) Begin() (*Tx, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	ret := &Tx{
		engine: e,
		tx:     tx,
	}
	ret.sql = ret.SQL()
	return ret, nil
}

// 查找缓存的sql.Stmt，在未找到的情况下，第二个参数返回false
func (e *Engine) Stmt(name string) (*sql.Stmt, bool) {
	return e.stmts.Get(name)
}

func (e *Engine) SQL() *SQL {
	return newSQL(e)
}

func (e *Engine) Where(cond string, args ...interface{}) *SQL {
	return e.SQL().Where(cond, args...)
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func (e *Engine) Insert(v interface{}) error {
	return insertMult(e.sql, v)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func (e *Engine) Update(v interface{}) error {
	return updateMult(e.sql, v)
}

// 删除指定的数据对象。
func (e *Engine) Delete(v interface{}) error {
	return deleteMult(e.sql, v)
}

// 根据obj创建表
func (e *Engine) Create(obj interface{}) error {
	m, err := core.NewModel(obj)
	if err != nil {
		return err
	}
	return e.Dialect().CreateTable(e, m)
}
