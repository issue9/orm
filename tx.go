// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/issue9/orm/builder"
	"github.com/issue9/orm/core"
)

// 事务对象
type Tx struct {
	engine *Engine
	tx     *sql.Tx
	sql    *builder.SQL

	stmtsMux sync.Mutex
	stmts    map[string]*core.Stmt
}

// 对orm/core.DB.DB()的实现。返回当前数据库对应的*sql.DB
func (t *Tx) DB() *sql.DB {
	return t.engine.DB()
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (t *Tx) Dialect() core.Dialect {
	return t.engine.Dialect()
}

// 去掉占位符。
func (t *Tx) prepareSQL(sql string) (string, []string) {
	sql = t.engine.replacer.Replace(sql)
	return core.ExtractArgs(sql)
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (t *Tx) Exec(sql string, args map[string]interface{}) (sql.Result, error) {
	realSQL, argNames := t.prepareSQL(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := t.tx.Exec(sql, args)
	if err != nil {
		return nil, newSQLError(err, t.engine.driverName, sql, realSQL, argList)
	}
	return r, nil
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (t *Tx) Query(sql string, args map[string]interface{}) (*sql.Rows, error) {
	realSQL, argNames := t.prepareSQL(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := t.tx.Query(sql, args)
	if err != nil {
		return nil, newSQLError(err, t.engine.driverName, sql, realSQL, argList)
	}
	return r, nil
}

// 对orm/core.DB.Prepare()的实现。预处理SQL语句成sql.Stmt实例。
func (t *Tx) Prepare(sql string, name ...string) (*core.Stmt, error) {
	if len(name) > 1 {
		return nil, errors.New("Prepare:name参数长度最大只能为１")
	}

	realSQL, argNames := t.prepareSQL(sql)
	stmt, err := t.tx.Prepare(sql)
	if err != nil {
		return nil, newSQLError(err, t.engine.driverName, sql, realSQL)
	}
	coreStmt := core.NewStmt(stmt, argNames)

	if len(name) == 1 {
		t.stmtsMux.Lock()
		if _, found := t.stmts[name[0]]; found {
			t.stmtsMux.Unlock()
			return nil, fmt.Errorf("该名称[%v]已经存在", name[0])
		}
		t.stmts[name[0]] = coreStmt
		t.stmtsMux.Unlock()
	}

	return coreStmt, nil
}

// 关闭当前的db。
// 不会关闭与之关联的engine实例，
// 仅是取消了与之的关联。
func (t *Tx) Close() error {
	t.engine = nil
	return nil
}

// 提交事务
// 提交之后，整个Tx对象将不再有效。
func (t *Tx) Commit() (err error) {
	if err = t.tx.Commit(); err == nil {
		t.Close()
	}
	return
}

// 回滚事务
func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// 查找缓存的sql.Stmt，在未找到的情况下，第二个参数返回false
func (t *Tx) GetStmt(name string) (stmt *core.Stmt, found bool) {
	stmt, found = t.stmts[name]
	return
}

// 返回一个新的SQL实例。
func (t *Tx) SQL() *builder.SQL {
	return builder.NewSQL(t)
}

// 指定一个条件语句，并返回SQL实例。
// 与SQL()方法稍微有一点不同，Where()会使用已缓存的SQL实例，
// 而SQL()方法会新声明一个SQL实例。
func (t *Tx) Where(cond string) *builder.SQL {
	return t.sql.Reset().Where(cond)
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

// 创建数据表。
func (t *Tx) Create(v ...interface{}) error {
	return createMult(t, v...)
}

// 删除表结构及数据。
func (t *Tx) Drop(tableName string) error {
	_, err := t.Exec("DROP TABLE "+tableName, nil)
	return err
}

// 清除表内容，但保留表结构。
func (t *Tx) Truncate(tableName string) error {
	_, err := t.Exec(t.Dialect().TruncateTableSQL(tableName), nil)
	return err
}
