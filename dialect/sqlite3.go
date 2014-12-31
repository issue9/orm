// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/issue9/orm/core"
)

type Sqlite3 struct{}

// implement core.Dialect.QuoteStr()
func (s *Sqlite3) QuoteStr() (l, r string) {
	return "[", "]"
}

// implement core.Dialect.SupportLastInsertId()
func (s *Sqlite3) SupportLastInsertId() bool {
	return true
}

// implement core.Dialect.GetDBName()
func (s *Sqlite3) GetDBName(dataSource string) string {
	// 取得最后个路径分隔符的位置，无论是否存在分隔符，用++
	// 表达式都正好能表示文件名开始的位置。
	start := strings.LastIndex(dataSource, string(os.PathSeparator))
	if start < 0 && runtime.GOOS == "windows" { // windows下可以使用/
		start = strings.LastIndex(dataSource, "/")
	}
	start++
	end := strings.LastIndex(dataSource, ".")

	if end < start { // 不存在扩展名，取全部文件名
		return dataSource[start:]
	}
	return dataSource[start:end]
}

// implement core.Dialect.LimitSQL()
func (s *Sqlite3) LimitSQL(limit int, offset ...int) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

// implement core.Dialect.CreateTable()
func (s *Sqlite3) CreateTable(db core.DB, m *core.Model) error {
	has, err := s.hasTable(db, m.Name)
	if err != nil {
		return err
	}

	if has {
		return s.upgradeTable(db, m)
	}
	return s.createTable(db, m)
}

// 是否存在指定名称的表
func (s *Sqlite3) hasTable(db core.DB, tableName string) (bool, error) {
	sql := "SELECT * FROM sqlite_master WHERE type='table' AND name=?"
	rows, err := db.Query(sql, tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// implement base.sqlType()
// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *Sqlite3) sqlType(buf *bytes.Buffer, col *core.Column) error {
	if col == nil {
		return errors.New("sqlType:col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的col.GoType值")
	}

	switch col.GoType.Kind() {
	case reflect.String:
		buf.WriteString("TEXT")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf.WriteString("INTEGER")
	case reflect.Float32, reflect.Float64:
		buf.WriteString("REAL")
	case reflect.Array, reflect.Slice:
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return errors.New("不支持数组类型")
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("INTEGER")
		case nullFloat64:
			buf.WriteString("REAL")
		case nullInt64:
			buf.WriteString("INTEGER")
		case nullString:
			buf.WriteString("TEXT")
		case timeType:
			buf.WriteString("DATETIME")
		}
	}

	return nil
}

// 创建表
func (s *Sqlite3) createTable(db core.DB, model *core.Model) error {
	buf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	buf.Grow(300)

	buf.WriteString(model.Name)
	buf.WriteByte('(')

	// 写入字段信息
	for _, col := range model.Cols {
		if err := createColSQL(s, buf, col); err != nil {
			return err
		}

		if col.IsAI() {
			buf.WriteString(" PRIMARY KEY AUTOINCREMENT")
		}
		buf.WriteByte(',')
	}

	// PK，若有自增，则已经在上面指定
	if len(model.PK) > 0 && !model.PK[0].IsAI() {
		createPKSQL(s, buf, model.PK, pkName)
		buf.WriteByte(',')
	}

	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(s, buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(s, buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(s, buf, chk, name)
		buf.WriteByte(',')
	}

	buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
	buf.WriteByte(')')          // end CreateTable

	_, err := db.Exec(buf.String())
	return err
}

// 更新表。Sqlite3并没有更改列类型的方法，只能采取官网说的方法来实现：
// http://www.sqlite.org/lang_altertable.html
func (s *Sqlite3) upgradeTable(db core.DB, model *core.Model) error {
	// 关闭外键
	if _, err := db.Exec("PRAGMA foreign_keys=OFF"); err != nil {
		return err
	}

	tmpName, err := s.rename(db, model.Name)
	if err != nil {
		return err
	}

	if err := s.createTable(db, model); err != nil {
		return err
	}

	// 从tmpName表中导出数据到model.Name表中
	// "INSERT INTO ?(cols...) SELECT cols FROM ?"
	cols := make([]string, 0, len(model.Cols))
	for colName, _ := range model.Cols {
		cols = append(cols, colName)
	}
	colsSQL := strings.Join(cols, ",")
	buf := bytes.NewBufferString("INSERT INTO ?(")
	buf.WriteString(colsSQL)
	buf.WriteString(") SELECT")
	buf.WriteString(colsSQL)
	buf.WriteString("FROM ?")
	if _, err := db.Exec(buf.String(), model.Name, tmpName); err != nil {
		return err
	}

	if _, err := db.Exec("DROP TABLE IF EXISTS ?", tmpName); err != nil {
		return err
	}

	// 打开外键
	_, err = db.Exec("PRAGMA foreign_keys=OFF")

	return err
}

// 将一个表重命名，新名称通过返回值返回。
func (s *Sqlite3) rename(db core.DB, tableName string) (string, error) {
	tmpName := tableName + "_tmp"
	for {
		has, err := s.hasTable(db, tmpName)
		if err != nil {
			return "", err
		}

		if !has {
			break
		}

		tmpName += "_tmp"
	}

	// 将当前表改名
	sql := "ALTER TABLE ? RENAME TO ?"
	if _, err := db.Exec(sql, tableName, tmpName); err != nil {
		return "", err
	}

	return tmpName, nil
}
