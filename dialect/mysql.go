// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/fetch"
)

type Mysql struct{}

// implement core.Dialect.GetDBName()
func (m *Mysql) GetDBName(dataSource string) string {
	start := strings.LastIndex(dataSource, "/")

	start++
	end := strings.LastIndex(dataSource, "?")
	if start > end { // 不存在参数
		return dataSource[start:]
	}

	return dataSource[start:end]
}

// implement core.Dialect.Quote
func (m *Mysql) QuoteStr() (string, string) {
	return "`", "`"
}

// implement core.Dialect.Limit()
func (m *Mysql) LimitSQL(limit int, offset ...int) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

// implement core.Dialect.SupportLastInsertId()
func (m *Mysql) SupportLastInsertId() bool {
	return true
}

// implement core.Dialect.UpgradeTable()
func (m *Mysql) UpgradeTable(db core.DB, model *core.Model, onlyCreate bool) error {
	sql := "SELECT `TABLE_NAME` FROM `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? and `TABLE_NAME`=?"
	rows, err := db.Query(sql, db.Name(), model.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() { // 不存在相同的表名
		return m.createTable(db, model)
	}

	if onlyCreate {
		return fmt.Errorf("CreateTable:该表名[%v]已经存在", model.Name)
	}

	return m.upgradeTable(db, model)
}

// implement base.sqlType()
func (m *Mysql) sqlType(buf *bytes.Buffer, col *core.Column) error {
	if col == nil {
		return errors.New("col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("无效的col.GoType值")
	}

	addIntLen := func() {
		if col.Len1 > 0 {
			buf.WriteByte('(')
			buf.WriteString(strconv.Itoa(col.Len1))
			buf.WriteByte(')')
		}
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		buf.WriteString("BOOLEAN")
	case reflect.Int8:
		buf.WriteString("SMALLINT")
		addIntLen()
	case reflect.Int16:
		buf.WriteString("MEDIUMINT")
		addIntLen()
	case reflect.Int32:
		buf.WriteString("INT")
		addIntLen()
	case reflect.Int64, reflect.Int: // reflect.Int大小未知，都当作是BIGINT处理
		buf.WriteString("BIGINT")
		addIntLen()
	case reflect.Uint8:
		buf.WriteString("SMALLINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint16:
		buf.WriteString("MEDIUMINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint32:
		buf.WriteString("INT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		buf.WriteString("BIGINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Float32, reflect.Float64:
		buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
	case reflect.String:
		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("LONGTEXT")
		}
	case reflect.Slice, reflect.Array: // []rune,[]byte当作字符串处理
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return fmt.Errorf("不支持[%v]类型的数组", k)
		}

		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("LONGTEXT")
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("BOOLEAN")
		case nullFloat64:
			buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
		case nullInt64:
			buf.WriteString("BIGINT")
			addIntLen()
		case nullString:
			if col.Len1 < 65533 {
				buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
			} else {
				buf.WriteString("LONGTEXT")
			}
		case timeType:
			buf.WriteString("DATETIME")
		}
	default:
		return fmt.Errorf("不支持的类型:[%v]", col.GoType.Name())
	}

	return nil
}

// 创建表
func (m *Mysql) createTable(db core.DB, model *core.Model) error {
	buf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	buf.Grow(300)

	buf.WriteString(model.Name)
	buf.WriteByte('(')

	// 写入字段信息
	for _, col := range model.Cols {
		if err := createColSQL(m, buf, col); err != nil {
			return err
		}

		if col.IsAI() {
			buf.WriteString(" PRIMARY KEY AUTO_INCREMENT")
		}
		buf.WriteByte(',')
	}

	// PK，若有自增，则已经在上面指定
	if len(model.PK) > 0 && !model.PK[0].IsAI() {
		createPKSQL(m, buf, model.PK, pkName)
		buf.WriteByte(',')
	}

	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(m, buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(m, buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(m, buf, chk, name)
		buf.WriteByte(',')
	}

	// key index不存在CONSTRAINT形式的语句
	if len(model.KeyIndexes) == 0 {
		for name, index := range model.KeyIndexes {
			buf.WriteString("INDEX ")
			buf.WriteString(name)
			buf.WriteByte('(')
			for _, col := range index {
				buf.WriteString(col.Name)
				buf.WriteByte(',')
			}
			buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
			buf.WriteString("),")
		}
	}

	buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
	buf.WriteByte(')')          // end CreateTable

	_, err := db.Exec(buf.String())
	return err
}

// 更新表
func (m *Mysql) upgradeTable(db core.DB, model *core.Model) error {
	if err := m.deleteIndexes(db, model); err != nil {
		return err
	}

	if err := m.upgradeCols(db, model); err != nil {
		return err
	}

	if err := addIndexes(m, db, model); err != nil {
		return err
	}

	// key
	buf := bytes.NewBufferString("ALTER TABLE ")
	buf.WriteString(model.Name)
	size := buf.Len()

	for name, index := range model.KeyIndexes {
		buf.Truncate(size)
		buf.WriteString(" ADD INDEX ")
		buf.WriteString(name)
		buf.WriteByte('(')
		for _, col := range index {
			buf.WriteString(col.Name)
			buf.WriteByte(',')
		}
		buf.Truncate(buf.Len() - 1)
		buf.WriteByte(')')

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	if model.AI == nil {
		return nil
	}

	// ALTER TABLE document MODIFY COLUMN document_id INT auto_increment
	buf.Truncate(size)
	buf.WriteString(" MODIFY COLUMN ")
	createColSQL(m, buf, model.AI)
	buf.WriteString(" PRIMARY KEY AUTO_INCREMENT")
	_, err := db.Exec(buf.String())
	return err
}

func (m *Mysql) unQuote(sql string) string {
	sql = strings.Replace(sql, core.QuoteLeft, "", 1)
	return strings.Replace(sql, core.QuoteRight, "", 1)
}

// 更新表的字段信息。
func (m *Mysql) upgradeCols(db core.DB, model *core.Model) error {
	oldCols, err := m.getCols(db, model.Name)
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
		//nameWithQuote := core.QuoteLeft + colName + core.QuoteRight
		_, found1 := oldCols[m.unQuote(colName)]
		//_, found2 := oldCols[core.QuoteLeft+colName+core.QuoteRight]
		if !found1 {
			buf.WriteString(" ADD ")
		} else {
			buf.WriteString(" MODIFY COLUMN ")
			delete(oldCols, colName)
		}

		createColSQL(m, buf, col)

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	if len(oldCols) == 0 {
		return nil
	}

	// 删除已经不存在于model中的字段。
	buf.Truncate(size)
	buf.WriteString(" DROP COLUMN ")
	size = buf.Len()
	for name, _ := range oldCols {
		buf.Truncate(size)
		buf.WriteString("`" + name + "`")
		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	return nil
}

// 获取表的字段信息。
// 返回值：键名为字段名，键值无效。
func (m *Mysql) getCols(db core.DB, tableName string) (map[string]interface{}, error) {
	sql := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
	rows, err := db.Query(sql, db.Name(), tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbCols, err := fetch.ColumnString(false, "COLUMN_NAME", rows)
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

// 删除表中的索引
func (m *Mysql) deleteIndexes(db core.DB, model *core.Model) error {
	// 删除有中的标准约束：pk,fk,unique
	sql := "SELECT CONSTRAINT_NAME, CONSTRAINT_TYPE FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA=? AND TABLE_NAME=?"
	rows, err := db.Query(sql, db.Name(), model.Name)
	if err != nil {
		return err
	}

	mapped, err := fetch.MapString(false, rows)
	if err != nil {
		return err
	}

	for _, record := range mapped {
		switch record["CONSTRAINT_TYPE"] {
		case "PRIMARY KEY":
			//_,err = db.Exec("ALTER TABLE "+model.Name+" MODIFY id INT NOT NULL");
			_, err = db.Exec("ALTER TABLE " + model.Name + " DROP PRIMARY KEY")
		case "FOREIGN KEY":
			_, err = db.Exec("ALTER TABLE " + model.Name + " DROP FOREIGN KEY " + record["CONSTRAINT_NAME"])
		case "UNIQUE":
			_, err = db.Exec("ALTER TABLE " + model.Name + " DROP INDEX " + record["CONSTRAINT_NAME"])
		default:
		}

		if err != nil {
			return err
		}
	}
	rows.Close()

	// 删除表中的非标准索引：key index
	sql = "SELECT `INDEX_NAME` FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA=? AND TABLE_NAME=?"
	rows, err = db.Query(sql, db.Name(), model.Name)
	if err != nil {
		return err
	}

	indexes, err := fetch.ColumnString(false, "INDEX_NAME", rows)
	if err != nil {
		return err
	}
	for _, index := range indexes {
		_, err := db.Exec("ALTER TABLE " + model.Name + " DROP INDEX " + index)
		if err != nil {
			return err
		}
	}
	rows.Close()

	return nil
}
