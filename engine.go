// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/issue9/orm/builder"
	"github.com/issue9/orm/core"
)

// sql语句中的占位符。
const (
	quoteLeft       = "{"
	quoteRight      = "}"
	tableNamePrefix = "#"
)

type engineMap struct {
	sync.Mutex
	items map[string]*Engine
}

var engines = engineMap{items: make(map[string]*Engine)}

// 数据库操作引擎.
type Engine struct {
	name       string
	dbName     string // 数据库的名称
	prefix     string // 表名前缀
	driverName string
	dialect    core.Dialect
	db         *sql.DB
	sql        *builder.SQL // 内置的SQL引擎，用于执行Update等操作

	// core.Stmt缓存
	stmts    map[string]*core.Stmt
	stmtsMux sync.Mutex

	replacer *strings.Replacer
}

// New 声明一个新的Engine实例。
func New(driverName, dataSourceName, engineName, prefix string) (*Engine, error) {
	if len(engineName) == 0 {
		return nil, errors.New("参数engineName不能为空")
	}

	engines.Lock()
	defer engines.Unlock()

	if _, found := engines.items[engineName]; found {
		return nil, fmt.Errorf("该名称[%v]的Engine已经存在", engineName)
	}

	dialect, found := core.Get(driverName)
	if !found {
		return nil, fmt.Errorf("newEngine:未找到与driverName[%v]相同的Dialect", driverName)
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	left, right := dialect.QuoteStr()
	engine := &Engine{
		name:       engineName,
		dbName:     dialect.GetDBName(dataSourceName),
		prefix:     prefix,
		driverName: driverName,
		dialect:    dialect,
		db:         db,
		stmts:      map[string]*core.Stmt{},
		replacer: strings.NewReplacer(
			quoteLeft, left,
			quoteRight, right,
			tableNamePrefix, prefix,
		),
	}
	engine.sql = engine.SQL()

	engines.items[engineName] = engine

	return engine, nil
}

// 获取指定名称的Engine，若不存在则返回nil。
func Get(engineName string) *Engine {
	engines.Lock()
	defer engines.Unlock()

	e, found := engines.items[engineName]
	if !found {
		return nil
	}
	return e
}

// 关闭所有的Engine
func CloseAll() error {
	engines.Lock()
	defer engines.Unlock()

	for _, v := range engines.items {
		if err := v.close(); err != nil {
			return err
		}
	}

	engines.items = make(map[string]*Engine)

	//core.ClearModels()
	return nil
}

// 对orm/core.DB.DB()的实现。返回当前数据库对应的*sql.DB
func (e *Engine) DB() *sql.DB {
	return e.db
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (e *Engine) Dialect() core.Dialect {
	return e.dialect
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (e *Engine) Exec(sql string, args map[string]interface{}) (sql.Result, error) {
	realSQL, argNames := e.prepareSQL(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := e.db.Exec(realSQL, argList...)
	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, e.driverName, sql, realSQL, argList...)
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (e *Engine) Query(sql string, args map[string]interface{}) (*sql.Rows, error) {
	realSQL, argNames := e.prepareSQL(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := e.db.Query(realSQL, argList...)
	if err != nil {
		return nil, newSQLError(err, e.driverName, sql, realSQL, argList...)
	}
	return r, nil
}

// 对orm/core.DB.Prepare()的实现。
// 预处理SQL语句成core.Stmt实例，若指定了name参数，则以name缓存该实例。
func (e *Engine) Prepare(sql string, name ...string) (*core.Stmt, error) {
	if len(name) > 1 {
		return nil, errors.New("Prepare:name参数长度最大只能为１")
	}

	realSQL, args := e.prepareSQL(sql)
	stmt, err := e.db.Prepare(sql)
	if err != nil {
		return nil, newSQLError(err, e.driverName, sql, realSQL)
	}
	coreStmt := core.NewStmt(stmt, args)

	if len(name) == 1 {
		e.stmtsMux.Lock()
		e.stmts[name[0]] = coreStmt
		e.stmtsMux.Unlock()
	}

	return coreStmt, nil
}

// 查找缓存的sql.Stmt，在未找到的情况下，第二个参数返回false
func (e *Engine) GetStmt(name string) (stmt *core.Stmt, found bool) {
	stmt, found = e.stmts[name]
	return
}

// 替换占位符。
func (e *Engine) prepareSQL(sql string) (string, []string) {
	sql = e.replacer.Replace(sql)
	return core.ExtractArgs(sql)
}

func (e *Engine) close() error {
	if err := e.db.Close(); err != nil {
		return err
	}
	e.stmts = nil
	delete(engines.items, e.name)

	return nil
}

// 关闭当前的Engine，销毁所有的数据。不能再次使用。
// 与之关联的Tx也将不能使用。
func (e *Engine) Close() error {
	engines.Lock()
	defer engines.Unlock()

	return e.close()
}

// 开始一个新的事务。
func (e *Engine) Begin() (*Tx, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	ret := &Tx{
		engine: e,
		tx:     tx,
		stmts:  map[string]*core.Stmt{},
	}
	ret.sql = ret.SQL()
	return ret, nil
}

// 产生一个SQL实例。
func (e *Engine) SQL() *builder.SQL {
	return builder.NewSQL(e)
}

// 指定一个条件语句，并返回SQL实例。
// 与SQL()方法稍微有一点不同，Where()会使用已缓存的SQL实例，
// 而SQL()方法会新声明一个SQL实例。
func (e *Engine) Where(cond string) *builder.SQL {
	return e.sql.Reset().Where(cond)
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

// 根据models创建表。
// 若表已经存在，则返回错误信息。
func (e *Engine) Create(models ...interface{}) error {
	return createMult(e, models...)
}

// 删除表结构及数据。
func (e *Engine) Drop(tableName string) error {
	_, err := e.Exec("DROP TABLE "+tableName, nil)
	return err
}

// 清除表内容，但保留表结构。
func (e *Engine) Truncate(tableName string) error {
	_, err := e.Exec(e.Dialect().TruncateTableSQL(tableName), nil)
	return err
}
