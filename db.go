// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

// 两个实现了core.DB接口的结构：Engine和Tx。
// 相对于database/sql下的DB和Tx，一个实现普通的
// 数据库操作，一个用于事务的操作。

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/dialect"
)

const (
	quoteLeft  = "{"
	quoteRight = "}"
	dbPrefix   = "#"
)

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
	replace := strings.NewReplacer(quoteLeft, l, quoteRight, r, dbPrefix, e.prefix)

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

// 产生一个SQL实例。
func (e *Engine) SQL() *SQL {
	return newSQL(e)
}

// 相当于调用了Engine.SQL().Where(...)
func (e *Engine) Where(cond string, args ...interface{}) *SQL {
	return e.SQL().Where(cond, args...)
}

// 插入一个或多个数据。
// v可以是struct或是相同struct组成的数组。
// 若v中指定了自增字段，则该字段的值在插入数据库时，
// 会被自动忽略。
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

// 事务对象
type Tx struct {
	engine *Engine
	tx     *sql.Tx
	sql    *SQL
}

// 对orm/core.DB.Name()的实现，返回当前操作的数据库名称。
func (t *Tx) Name() string {
	return t.engine.Name()
}

// 对orm/core.DB.GetStmts()的实现，返回当前的sql.Stmt实例缓存容器。
func (t *Tx) GetStmts() *core.Stmts {
	return t.engine.GetStmts()
}

// 对orm/core.DB.PrepareSQL()的实现。替换语句的各种占位符。
func (t *Tx) PrepareSQL(sql string) string {
	return t.engine.PrepareSQL(sql)
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (t *Tx) Dialect() core.Dialect {
	return t.engine.Dialect()
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (t *Tx) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(sql, args...)
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (t *Tx) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(sql, args...)
}

// 对orm/core.DB.QueryRow()的实现。
// 执行一条查询语句，并返回第一条符合条件的记录。
func (t *Tx) QueryRow(sql string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(sql, args...)
}

// 对orm/core.DB.Prepare()的实现。预处理SQL语句成sql.Stmt实例。
func (t *Tx) Prepare(sql string) (*sql.Stmt, error) {
	return t.tx.Prepare(sql)
}

// 关闭当前的db
func (t *Tx) close() {
	// 仅仅取消与engine的关联。
	t.engine = nil
}

// 提交事务
// 提交之后，整个Tx对象将不再有效。
func (t *Tx) Commit() (err error) {
	if err = t.tx.Commit(); err == nil {
		t.close()
	}
	return
}

// 回滚事务
func (t *Tx) Rollback() {
	t.tx.Rollback()
}

// 查找缓存的sql.Stmt，在未找到的情况下，第二个参数返回false
func (t *Tx) Stmt(name string) (*sql.Stmt, bool) {
	stmt, found := t.engine.Stmt(name)
	if !found {
		return nil, false
	}

	return t.tx.Stmt(stmt), true
}

// 返回一个新的SQL实例。
func (t *Tx) SQL() *SQL {
	return newSQL(t)
}

// 相当于调用了Engine.SQL().Where(...)
func (t *Tx) Where(cond string, args ...interface{}) *SQL {
	return t.SQL().Where(cond, args...)
}

// 插入一个或多个数据。
// v可以是struct或是相同struct组成的数组。
// 若v中指定了自增字段，则该字段的值在插入数据库时，
// 会被自动忽略。
func (t *Tx) Insert(v interface{}) error {
	return insertMult(t.sql, v)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func (t *Tx) Update(v interface{}) error {
	return updateMult(t.sql, v)
}

// 删除指定的数据对象。
func (t *Tx) Delete(v interface{}) error {
	return deleteMult(t.sql, v)
}
