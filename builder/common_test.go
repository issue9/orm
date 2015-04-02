// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 公共测试用代码。

package builder

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"

	_ "github.com/mattn/go-sqlite3"
)

var dbFile = "./test.db"

// 用于chkSQLEqual
var replacer = strings.NewReplacer(")", " ) ", "(", " ( ", ",", " , ")

// 用于chkSQLEqual
var spaceReplaceRegexp = regexp.MustCompile("\\s+")

// 检测两条SQL语句是否相等，忽略大小写与多余的空格。
func chkSQLEqual(a *assert.Assertion, s1, s2 string) {
	// 将'(', ')', ',' 等字符的前后空格标准化
	s1 = replacer.Replace(s1)
	s2 = replacer.Replace(s2)

	// 转换成小写，去掉首尾空格
	s1 = strings.TrimSpace(strings.ToLower(s1))
	s2 = strings.TrimSpace(strings.ToLower(s2))

	// 去掉多余的空格。
	s1 = spaceReplaceRegexp.ReplaceAllString(s1, " ")
	s2 = spaceReplaceRegexp.ReplaceAllString(s2, " ")

	a.Equal(s1, s2)
}

//////////////////////////////////////////////////////////////////////
/////////////////////// core.Dialect
//////////////////////////////////////////////////////////////////////

type sqlite3 struct{}

var _ core.Dialect = &sqlite3{}

func (s *sqlite3) QuoteStr() (l, r string) {
	return "[", "]"
}

// implement core.Dialect.GetDBName()
func (s *sqlite3) GetDBName(dataSource string) string {
	return "test"
}

// implement core.Dialect.LimitSQL()
func (s *sqlite3) LimitSQL(limit interface{}, offset ...interface{}) string {
	if len(offset) == 0 {
		return " LIMIT " + core.AsSQLValue(limit)
	}

	return " LIMIT " + core.AsSQLValue(limit) + " OFFSET " + core.AsSQLValue(offset[0])
}

// implement core.Dialect.CreateTableSQL()
func (s *sqlite3) CreateTableSQL(model *core.Model) (string, error) {
	fmt.Errorf("未实现CreateTableSQL")
	os.Exit(1)

	return "CREATE TABLE ...", nil
}

// implement core.Dialect.TruncateTableSQL()
func (s *sqlite3) TruncateTableSQL(tableName string) string {
	return "DELETE FROM " + tableName
}

func init() {
	if err := core.Register("sqlite3", &sqlite3{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//////////////////////////////////////////////////////////////////////
///////////////////////  core.DB
//////////////////////////////////////////////////////////////////////

// 对应DB结构初始化时插入的数据
type user struct {
	ID      int    `orm:"name(id)"`
	Account string `orm:"name(account)"`
}

type DB struct {
	db      *sql.DB
	stmts   map[string]*core.Stmt
	dialect core.Dialect
}

var _ core.DB = &DB{}

// 声明一个DB实例，并带上一个10条记录的表，表内容如下：
//  id(主键)     account
//  1            account-1
//  2            account-2
//  3            account-3
// 依次类推
func newDB(a *assert.Assertion) *DB {
	dialect, found := core.Get("sqlite3")
	a.True(found)

	// 存在文件，则先删除
	_, err := os.Stat(dbFile)
	if err == nil || os.IsExist(err) {
		a.NotError(os.Remove(dbFile))
	}

	db, err := sql.Open("sqlite3", dbFile)
	a.NotError(err).NotNil(db)

	// 创建表
	sql := "CREATE TABLE IF NOT EXISTS user(id INTEGER PRIMARY KEY AUTOINCREMENT,account text)"
	_, err = db.Exec(sql)
	a.NotError(err)

	// 插入数据
	stmt, err := db.Prepare("INSERT INTO user (account) values(?)")
	a.NotError(err)
	for i := 1; i < 11; i++ {
		_, err = stmt.Exec("account-" + strconv.Itoa(i))
		a.NotError(err)
	}

	return &DB{
		db:      db,
		stmts:   map[string]*core.Stmt{},
		dialect: dialect,
	}
}

// 关闭db
func (db *DB) Close(a *assert.Assertion) {
	a.NotError(db.db.Close())

	_, err := os.Stat(dbFile)
	if err == nil || os.IsExist(err) {
		a.NotError(os.Remove(dbFile))
	}

	db.stmts = nil
}

func (db *DB) DB() *sql.DB {
	return db.db
}

func (db *DB) Dialect() core.Dialect {
	return db.dialect
}

func (db *DB) Exec(sql string, args map[string]interface{}) (sql.Result, error) {
	realSQL, argNames := core.ExtractArgs(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := db.db.Exec(realSQL, argList...)
	if err != nil {
		return nil, fmt.Errorf("执行[%v]时出错", realSQL)
	}
	return r, nil
}

func (db *DB) Query(sql string, args map[string]interface{}) (*sql.Rows, error) {
	realSQL, argNames := core.ExtractArgs(sql)

	argList, err := core.ConvArgs(argNames, args)
	if err != nil {
		return nil, err
	}

	r, err := db.db.Query(realSQL, argList...)
	if err != nil {
		return nil, fmt.Errorf("Query:执行[%v]时出错", realSQL)
	}
	return r, nil
}

func (db *DB) Prepare(sql string, name ...string) (*core.Stmt, error) {
	if len(name) > 1 {
		return nil, errors.New("Prepare:name参数长度最大只能为１")
	}

	realSQL, args := core.ExtractArgs(sql)
	stmt, err := db.db.Prepare(sql)
	if err != nil {
		return nil, fmt.Errorf("Prepare:执行[%v]时出错", realSQL)
	}
	coreStmt := core.NewStmt(stmt, args)

	if len(name) == 1 {
		if _, found := db.stmts[name[0]]; found {
			return nil, fmt.Errorf("该名称[%v]已经存在", name[0])
		}
		db.stmts[name[0]] = coreStmt
	}

	return coreStmt, nil
}

func (db *DB) GetStmt(name string) (stmt *core.Stmt, found bool) {
	stmt, found = db.stmts[name]
	return
}
