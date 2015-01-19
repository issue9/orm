// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
	"errors"
)

// 通用但又没有统一标准的数据库功能接口。
//
// 有可能一个Dialect实例会被多个实例引用，
// 不应该在Dialect实例中保存状态值等内容。
type Dialect interface {
	// 对字段或是表名的引用字符。
	QuoteStr() (left, right string)

	// 从dataSourceName变量中获取数据库的名称。
	GetDBName(dataSourceName string) string

	// 生成LIMIT N OFFSET M 或是相同的语意的语句。
	// offset值为一个可选参数，若不指定，则表示LIMIT N语句。
	// 返回的是对应数据库的limit语句以及语句中占位符对应的值。
	LimitSQL(limit interface{}, offset ...interface{}) (sql string)

	// 根据数据模型，创建表。
	CreateTableSQL(m *Model) (sql string, err error)

	// 清空表内容，重置AI。
	TruncateTableSQL(tableName string) (sql string)
}

// 操作数据库的接口，用于统一普通数据库操作和事务操作。
type DB interface {
	DB() *sql.DB

	// 返回Dialect接口。
	Dialect() Dialect

	Exec(sql string, args map[string]interface{}) (sql.Result, error)

	// 相当于sql.DB.Query()。
	Query(sql string, args map[string]interface{}) (*sql.Rows, error)

	// 相当于sql.DB.Prepare()。
	// 若存在name参数，则以name为名称缓存此条Stmt。
	// 若已存在相同名称的，则覆盖原内容。
	Prepare(sql string, name ...string) (*Stmt, error)

	// 获取缓存的Stmt，若不存在，found返回false值。
	GetStmt(name string) (stmt *Stmt, found bool)
}

// 包装sql.Stmt，使其可以指定命名参数。
type Stmt struct {
	stmt     *sql.Stmt
	argNames []string
}

// 声明一个新的Stmt
func NewStmt(stmt *sql.Stmt, argNames []string) *Stmt {
	return &Stmt{stmt: stmt, argNames: argNames}
}

func (s *Stmt) GetStmt() *sql.Stmt {
	return s.stmt
}

// 关闭当前的Stmt实例。
func (s *Stmt) Close() error {
	s.argNames = s.argNames[:0]
	return s.stmt.Close()
}

// 执行非查询语句。
// 参数将根据命名参数进行排序。
func (s *Stmt) Exec(args map[string]interface{}) (sql.Result, error) {
	if len(s.argNames) != len(args) {
		return nil, errors.New("ExecMap:参数长度不一样")
	}

	if len(s.argNames) == 0 && len(args) == 0 {
		return s.stmt.Exec()
	}

	argList, err := ConvArgs(s.argNames, args)
	if err != nil {
		return nil, err
	}

	return s.stmt.Exec(argList...)
}

// 执行查询语句
func (s *Stmt) Query(args map[string]interface{}) (*sql.Rows, error) {
	if len(s.argNames) != len(args) {
		return nil, errors.New("ExecMap:参数长度不一样")
	}

	if len(s.argNames) == 0 && len(args) == 0 {
		return s.stmt.Query()
	}

	argList, err := ConvArgs(s.argNames, args)
	if err != nil {
		return nil, err
	}

	return s.stmt.Query(argList...)
}
