// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/issue9/orm/core"
)

const (
	Delete = iota
	Insert
	Update
	Select
)

type Errors []error

func (err Errors) Error() string {
	ret := "发生以下错误:"
	for index, msg := range err {
		ret += (strconv.Itoa(index) + ":" + msg.Error())
	}

	return ret
}

// SQL语句的构建和缓存工具。
// 可以通过函数链的形式来写SQL语句，无须关注SQL语句本身的结构顺序。
// 把持使用命名参数的形式传递参数。
//
//  sql := NewSQL(db)
//      Table("user").
//      Where("id>@id").
//      And("username like @username").
//      AndIsNotNull("Email").
//      Desc("id")
//  data := sql.FetchMaps(map[string]interface{}{"id":1,"username":"abc"})
type SQL struct {
	db        core.DB
	tableName string
	errors    []error       // 所有的错误缓存
	buf       *bytes.Buffer // 语句缓存
	cond      *bytes.Buffer
	data      map[string]interface{} // 额外数据，如insert和update中需要插入的数据

	// select
	cols     []string
	join     *bytes.Buffer
	order    *bytes.Buffer
	limitSQL string
}

// 新建一个SQL实例。
func NewSQL(db core.DB) *SQL {
	return &SQL{
		db:     db,
		errors: []error{},
		buf:    bytes.NewBuffer([]byte{}),
		cond:   bytes.NewBuffer([]byte{}),
		data:   map[string]interface{}{},

		// select
		cols:  []string{},
		join:  bytes.NewBuffer([]byte{}),
		order: bytes.NewBuffer([]byte{}),
	}
}

// 重置SQL语句的状态。除了SQL.db以外，
// 其它属性都将被重围为初始状态。
func (s *SQL) Reset() *SQL {
	s.tableName = ""
	s.errors = s.errors[:0]
	s.buf.Reset()
	s.cond.Reset()
	s.data = map[string]interface{}{}

	// select
	s.cols = s.cols[:0]
	s.join.Reset()
	s.order.Reset()
	s.limitSQL = ""

	return s
}

// 是否存在错误
func (s *SQL) HasErrors() bool {
	return len(s.errors) > 0
}

// 返回错误内容
func (s *SQL) GetErrors() []error {
	return s.errors
}

// 设置表名。多次调用，只有最后一次启作用。
func (s *SQL) Table(name string) *SQL {
	s.tableName = name
	return s
}

// update/insert 语句可以用此方法批量指定需要更新的字段及相应的数据，
// 其它语句，忽略此方法产生的数据。
func (s *SQL) Data(data map[string]interface{}) *SQL {
	s.data = data
	return s
}

// update/insert 语句可以用此方法指定一条需要更新的字段及相应的数据。
// 其它语句，忽略此方法产生的数据。
func (s *SQL) Set(col string, val interface{}) *SQL {
	s.data[col] = val
	return s
}

// 将当前语句预编译并缓存到stmts中，方便之后再次使用。
// action用于指定语句的类型，可以是Insert, Delete, Update或是Select。
// 若name与已有的相同，则会覆盖！
func (s *SQL) Prepare(action int, name string) (*core.Stmt, error) {
	var sql string
	switch action {
	case Delete:
		sql = s.deleteSQL()
	case Update:
		sql = s.updateSQL()
	case Insert:
		sql = s.insertSQL()
	case Select:
		sql = s.selectSQL()
	default:
		return nil, fmt.Errorf("Stmt:无效的的action值[%v]", action)
	}

	return s.db.Prepare(sql, name)
}

func (s *SQL) ExecMap(action int, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

// 执行当前语句。
// action指定语句类型，可以是Delete,Insert或Update，但不能是Select
func (s *SQL) Exec(action int, args map[string]interface{}) (sql.Result, error) {
	switch action {
	case Delete:
		return s.Delete(args)
	case Update:
		return s.Update(args)
	case Insert:
		return s.Insert(args)
	case Select:
		return nil, errors.New("Exec:select语句不能使用Exec()方法执行")
	default:
		return nil, fmt.Errorf("Exec:无效的的action值[%v]", action)
	}
}

// 产生delete语句
func (s *SQL) deleteSQL() string {
	s.buf.Reset()
	s.buf.WriteString("DELETE FROM ")
	s.buf.WriteString(s.tableName)

	// where
	s.buf.WriteString(s.cond.String())

	return s.buf.String()
}

// 执行DELETE操作。
// 相当于s.Exec(Delete, args)
func (s *SQL) Delete(args map[string]interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	return s.db.Exec(s.deleteSQL(), args)
}

// 产生update语句
func (s *SQL) updateSQL() string {
	s.buf.Reset()
	s.buf.WriteString("UPDATE ")
	s.buf.WriteString(s.tableName)
	s.buf.WriteString(" SET ")
	for k, v := range s.data {
		s.buf.WriteString(k)
		s.buf.WriteString("=")
		s.buf.WriteString(core.AsSQLValue(v))
		s.buf.WriteByte(',')
	}
	s.buf.Truncate(s.buf.Len() - 1)

	// where
	s.buf.WriteString(s.cond.String())

	return s.buf.String()
}

// 执行UPDATE操作。
// 相当于s.Exec(Update, args)
func (s *SQL) Update(args map[string]interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	return s.db.Exec(s.updateSQL(), args)
}

// 产生insert语句
func (s *SQL) insertSQL() string {
	s.buf.Reset() // 清空之前的内容

	s.buf.WriteString("INSERT INTO ")
	s.buf.WriteString(s.tableName)

	var keys []string
	var vals []string
	for k, v := range s.data {
		keys = append(keys, k)
		vals = append(vals, core.AsSQLValue(v))
	}

	s.buf.WriteByte('(')
	s.buf.WriteString(strings.Join(keys, ","))
	s.buf.WriteString(") VALUES(")
	s.buf.WriteString(strings.Join(vals, ","))
	s.buf.WriteByte(')')

	return s.buf.String()
}

// 执行INSERT操作。
// 相当于s.Exec(Insert, args)
func (s *SQL) Insert(args map[string]interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	return s.db.Exec(s.insertSQL(), args)
}
