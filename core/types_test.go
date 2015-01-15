// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 与types.go中声明的接口有关的测试用例。

package core

type sqlite3 struct{}

func (s *sqlite3) QuoteStr() (l, r string) {
	return "[", "]"
}

// implement Dialect.GetDBName()
func (s *sqlite3) GetDBName(dataSource string) string {
	return "DBname"
}

// implement Dialect.LimitSQL()
func (s *sqlite3) LimitSQL(limit interface{}, offset ...interface{}) string {
	return "LIMIT ..."
}

// implement Dialect.CreateTableSQL()
func (s *sqlite3) CreateTableSQL(model *Model) (string, error) {
	return "CREATE TABLE ...", nil
}

// implement Dialect.TruncateTableSQL()
func (s *sqlite3) TruncateTableSQL(tableName string) string {
	return "DELETE FROM " + tableName
}

/////////////////////////////////////////////////////////////////////
////////////////////////////// dialect.mysql
/////////////////////////////////////////////////////////////////////

type mysql struct{}

// implement Dialect.GetDBName()
func (m *mysql) GetDBName(dataSource string) string {
	return "dbname"
}

// implement Dialect.Quote
func (m *mysql) QuoteStr() (string, string) {
	return "`", "`"
}

// implement Dialect.Limit()
func (m *mysql) LimitSQL(limit interface{}, offset ...interface{}) string {
	return "LIMIT ..."
}

// implement Dialect.CreateTableSQL()
func (m *mysql) CreateTableSQL(model *Model) (string, error) {
	return "CREATE TABLE ...", nil
}

// implement Dialect.TruncateTableSQL()
func (m *mysql) TruncateTableSQL(tableName string) string {
	return "TRUNCATE TABLE " + tableName
}
