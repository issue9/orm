// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/fetch"
)

type Postgres struct{}

// implement core.Dialect.QuoteStr()
func (p *Postgres) QuoteStr() (l, r string) {
	return `"`, `"`
}

// implement core.Dialect.SupportLastInsertId()
func (p *Postgres) SupportLastInsertId() bool {
	return true
}

// 匹配dbname=dbname 或是dbname =dbname等格式
var dbnamePrefix = regexp.MustCompile(`\s*=\s*|\s+`)

// implement core.Dialect.GetDBName()
func (p *Postgres) GetDBName(dataSource string) string {
	// dataSource样式：user=user dbname = db password=
	words := dbnamePrefix.Split(dataSource, -1)
	for index, word := range words {
		if word != "dbname" {
			continue
		}

		if index+1 >= len(words) {
			return ""
		}

		return words[index+1]
	}

	return ""
}

// implement core.Dialect.LimitSQL()
func (p *Postgres) LimitSQL(limit int, offset ...int) (sql string, args []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

// implement core.Dialect.CreateTable()
func (p *Postgres) CreateTable(db core.DB, model *core.Model) error {
	sql := "SELECT * FROM pg_tables where schemaname = 'public' and tablename=?"
	rows, err := db.Query(sql, model.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() { // 表已经存在
		return p.upgradeTable(db, model)
	}
	return p.createTable(db, model)
}

// 创建新表
func (p *Postgres) createTable(db core.DB, model *core.Model) error {
	buf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	buf.Grow(300)

	buf.WriteString(model.Name)
	buf.WriteByte('(')

	// 写入字段信息
	for _, col := range model.Cols {
		createColSQL(p, buf, col)
		buf.WriteByte(',')
	}

	// PK
	if len(model.PK) > 0 {
		createPKSQL(p, buf, model.PK, pkName)
		buf.WriteByte(',')
	}

	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(p, buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(p, buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(p, buf, chk, name)
		buf.WriteByte(',')
	}

	buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
	buf.WriteByte(')')          // end CreateTable

	_, err := db.Exec(buf.String())
	return err
}

// 更新表
func (p *Postgres) upgradeTable(db core.DB, model *core.Model) error {
	if err := p.upgradeCols(db, model); err != nil {
		return err
	}

	if err := p.deleteConstraints(db, model); err != nil {
		return err
	}

	return addIndexes(p, db, model)
}

// 更新表的列信息。
// 将model中的列与表中的列做对比：存在的修改；不存在的添加；只存在于
// 表中的列则直接删除。
func (p *Postgres) upgradeCols(db core.DB, model *core.Model) error {
	dbColsMap, err := p.getCols(db, model)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString("ALTER TABLE ")
	buf.WriteString(model.Name)
	size := buf.Len()

	// 将model中的列信息作用于数据库中的表，
	// 并将过滤dbCols中的列，只剩下不存在于model中的字段。
	for colName, col := range model.Cols {
		buf.Truncate(size)

		if _, found := dbColsMap[colName]; !found {
			buf.WriteString(" ADD ")
		} else {
			buf.WriteString(" MODIFY COLUMN ")
			delete(dbColsMap, colName)
		}

		createColSQL(p, buf, col)

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	if len(dbColsMap) == 0 {
		return nil
	}

	// 删除已经不存在于model中的字段。
	buf.Truncate(size)
	buf.WriteString(" DROP COLUMN ")
	size = buf.Len()
	for name, _ := range dbColsMap {
		buf.Truncate(size)
		buf.WriteString(name)
		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	return nil
}

// 获取表的列信息
func (p *Postgres) getCols(db core.DB, model *core.Model) (map[string]interface{}, error) {
	sql := "SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ?"
	rows, err := db.Query(sql, model.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbCols, err := fetch.ColumnString(false, "column_name", rows)
	if err != nil {
		return nil, nil
	}

	// 转换成map，仅用到键名，键值一律置空
	dbColsMap := make(map[string]interface{}, len(dbCols))
	for _, col := range dbCols {
		dbColsMap[col] = nil
	}

	return dbColsMap, nil
}

// 删除表中所有约束
func (p *Postgres) deleteConstraints(db core.DB, model *core.Model) error {
	sql := "SELECT  con.conname FROM pg_constraint AS con, pg_class AS cls WHERE con.conrelid=c.oid AND c.relname=?"
	rows, err := db.Query(sql, model.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	conts, err := fetch.ColumnString(false, "conname", rows)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString("ALTER TABLE ? DROP CONSTRAINT ?")
	for _, cont := range conts {
		if _, err := db.Exec(buf.String(), model.Name, cont); err != nil {
			return err
		}
	}

	return nil
}

// implement base.sqlType
// 将col转换成sql类型，并写入buf中。
func (p *Postgres) sqlType(buf *bytes.Buffer, col *core.Column) error {
	if col == nil {
		return errors.New("col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("无效的col.GoType值")
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		buf.WriteString("BOOLEAN")
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		if col.IsAI() {
			buf.WriteString("SERIAL")
		} else {
			buf.WriteString("SMALLINT")
		}
	case reflect.Int32, reflect.Uint32:
		if col.IsAI() {
			buf.WriteString("SERIAL")
		} else {
			buf.WriteString("INT")
		}
	case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint:
		if col.IsAI() {
			buf.WriteString("BIGSERIAL")
		} else {
			buf.WriteString("BIGINT")
		}
	case reflect.Float32, reflect.Float64:
		buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
	case reflect.String:
		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Slice, reflect.Array: // []rune,[]byte当作字符串处理
		k := col.GoType.Elem().Kind()
		if (k != reflect.Int8) && (k != reflect.Int32) {
			return errors.New("不支持数组类型")
		}

		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("BOOLEAN")
		case nullFloat64:
			buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
		case nullInt64:
			if col.IsAI() {
				buf.WriteString("BIGSERIAL")
			} else {
				buf.WriteString("BIGINT")
			}
		case nullString:
			if col.Len1 < 65533 {
				buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
			}
			buf.WriteString("TEXT")
		case timeType:
			buf.WriteString("TIME")
		}
	default:
		return fmt.Errorf("不支持的类型:[%v]", col.GoType.Name())
	}

	return nil
}
