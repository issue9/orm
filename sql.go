// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/fetch"
)

const (
	Delete = iota
	Insert
	Update
	Select
)

// sql := sqlbuild.New()
// sql.Table("#user").
//     Where("id>?",5).
//     And("username like ?", "%admin%").
type SQL struct {
	db        core.DB
	tableName string
	errors    []error       // 所有的错误缓存
	buf       *bytes.Buffer // 语句缓存

	// where
	cond     *bytes.Buffer
	condArgs []interface{}

	// data
	cols []string
	vals []interface{}

	// select
	join      *bytes.Buffer
	order     *bytes.Buffer
	limitSQL  string
	limitArgs []interface{}
}

// 新建一个SQL实例。
func newSQL(db core.DB) *SQL {
	return &SQL{
		db:     db,
		errors: []error{},
		buf:    bytes.NewBuffer([]byte{}),

		// where
		cond:     bytes.NewBuffer([]byte{}),
		condArgs: []interface{}{},

		// data
		cols: []string{},
		vals: []interface{}{},

		// select
		join:  bytes.NewBuffer([]byte{}),
		order: bytes.NewBuffer([]byte{}),
		// limitArgs: []interface{}{}, // 无需初始化，直接从dialect赋值得到
	}
}

// 重置SQL语句的状态。除了SQL.db以外，
// 其它属性都将被重围为初始状态。
func (s *SQL) Reset() *SQL {
	s.tableName = ""
	s.errors = s.errors[:0]
	s.buf.Reset()

	// where
	s.cond.Reset()
	s.condArgs = s.condArgs[:0]

	// data
	s.cols = s.cols[:0]
	s.vals = s.vals[:0]

	// select
	s.join.Reset()
	s.order.Reset()
	s.limitArgs = s.limitArgs[:0]

	return s
}

// 是否存在错误
func (s *SQL) HasErrors() bool {
	return len(s.errors) > 0
}

// 返回错误内容
func (s *SQL) Errors() []error {
	return s.errors
}

// SQL.And()的别名
func (s *SQL) Where(cond string, args ...interface{}) *SQL {
	return s.And(cond, args...)
}

// WHERE ... AND ...
func (s *SQL) And(cond string, args ...interface{}) *SQL {
	return s.build(0, cond, args...)
}

// WHERE ... OR ...
func (s *SQL) Or(cond string, args ...interface{}) *SQL {
	return s.build(1, cond, args...)
}

// WHERE ... AND col BETWEEN ...
func (s *SQL) AndBetween(col string, start, end interface{}) *SQL {
	return s.between(0, col, start, end)
}

// WHERE ... OR col BETWEEN ...
func (s *SQL) OrBetween(col string, start, end interface{}) *SQL {
	return s.between(1, col, start, end)
}

// SQL.AndBetween()的别名
func (s *SQL) Between(col string, start, end interface{}) *SQL {
	return s.AndBetween(col, start, end)
}

// WHERE ... AND col IN (...)
func (s *SQL) AndIn(col string, args ...interface{}) *SQL {
	return s.in(0, col, args...)
}

// WHERE ... OR col IN (...)
func (s *SQL) OrIn(col string, args ...interface{}) *SQL {
	return s.in(1, col, args...)
}

// SQL.AndIn()的别名
func (s *SQL) In(col string, args ...interface{}) *SQL {
	return s.AndIn(col, args...)
}

// WHERE ... AND col IS NULL
func (s *SQL) AndIsNull(col string) *SQL {
	return s.isNull(0, col)
}

// WHERE ... OR col IN NULL
func (s *SQL) OrIsNull(col string) *SQL {
	return s.isNull(1, col)
}

// SQL.AndIsNull()的别名
func (s *SQL) IsNull(col string) *SQL {
	return s.AndIsNull(col)
}

// WHERE ... AND col IS NOT NULL
func (s *SQL) AndIsNotNull(col string) *SQL {
	return s.isNotNull(0, col)
}

// WHERE ... OR col IS NOT NULL
func (s *SQL) OrIsNotNull(col string) *SQL {
	return s.isNotNull(1, col)
}

// SQL.AndIsNotNull()的别名
func (s *SQL) IsNotNull(col string) *SQL {
	return s.AndIsNull(col)
}

// 所有SQL子句的构建，最终都调用此方法来写入实例中。
// op 与前一个语句的连接符号，可以是and或是or常量；
// cond 条件语句，值只能是占位符，不能直接写值；
// condArgs 占位符对应的值。
//  w := newSQL(...)
//  w.build(0, "username=='abc'") // 错误：不能使用abc，只能使用？占位符。
//  w.build(1, "username=?", "abc") // 正确，将转换成: and username='abc'
func (s *SQL) build(op int, cond string, args ...interface{}) *SQL {
	switch {
	case s.cond.Len() == 0:
		s.cond.WriteString(" WHERE(")
	case op == 0:
		s.cond.WriteString(" AND(")
	case op == 1:
		s.cond.WriteString(" OR(")
	default:
		s.errors = append(s.errors, fmt.Errorf("build:无效的op操作符:[%v]", op))
	}

	s.cond.WriteString(cond)
	s.cond.WriteByte(')')

	s.condArgs = append(s.condArgs, args...)

	return s
}

// SQL col in(v1,v2)语句的实现函数，供andIn()和orIn()函数调用。
func (s *SQL) in(op int, col string, args ...interface{}) *SQL {
	if len(args) <= 0 {
		s.errors = append(s.errors, errors.New("in:args参数不能为空"))
		return s
	}

	cond := bytes.NewBufferString(col)
	cond.WriteString(" IN(")
	cond.WriteString(strings.Repeat("?,", len(s.condArgs)))
	cond.Truncate(cond.Len() - 1) // 去掉最后的逗号
	cond.WriteByte(')')

	return s.build(op, cond.String(), s.condArgs...)
}

// 供andBetween()和orBetween()调用。
func (s *SQL) between(op int, col string, start, end interface{}) *SQL {
	return s.build(op, col+" BETWEEN ? AND ?", start, end)
}

// 供andIsNull()和orIsNull()调用。
func (w *SQL) isNull(op int, col string) *SQL {
	return w.build(op, col+" IS NULL")
}

// 供andIsNotNull()和orIsNotNull()调用。
func (w *SQL) isNotNull(op int, col string) *SQL {
	return w.build(op, col+" IS NOT NULL")
}

// 设置表名。多次调用，只有最后一次启作用。
func (s *SQL) Table(name string) *SQL {
	s.tableName = name
	return s
}

// 指定列名。
// update/insert 语句可以用此方法指定需要更新的列。
// 若需要指定数据，请使用Data()或是Add()方法；
//
// select 语句可以用此方法指定需要获取的列。
func (s *SQL) Columns(cols ...string) *SQL {
	s.cols = append(s.cols, cols...)

	return s
}

// update/insert 语句可以用此方法批量指定需要更新的字段及相应的数据，
// 其它语句，忽略此方法产生的数据。
func (s *SQL) Data(data map[string]interface{}) *SQL {
	for k, v := range data {
		s.cols = append(s.cols, k)
		s.vals = append(s.vals, v)
	}
	return s
}

// update/insert 语句可以用此方法指定一条需要更新的字段及相应的数据。
// 其它语句，忽略此方法产生的数据。
func (s *SQL) Add(col string, val interface{}) *SQL {
	s.cols = append(s.cols, col)
	s.vals = append(s.vals, val)

	return s
}

var joinType = []string{" LEFT JOIN ", " RIGHT JOIN ", " INNER JOIN ", " FULL JOIN "}

// join功能
func (s *SQL) joinOn(typ int, table string, on string) *SQL {
	if typ < 0 && typ > 4 {
		s.errors = append(s.errors, fmt.Errorf("joinOn:错误的typ值:[%v]", typ))
	}

	s.join.WriteString(joinType[typ])
	s.join.WriteString(table)
	s.join.WriteString(" ON ")
	s.join.WriteString(on)

	return s
}

// LEFT JOIN ... ON ...
func (s *SQL) LeftJoin(table, on string) *SQL {
	return s.joinOn(0, table, on)
}

// RIGHT JOIN ... ON ...
func (s *SQL) RightJoin(table, on string) *SQL {
	return s.joinOn(1, table, on)
}

// INNER JOIN ... ON ...
func (s *SQL) InnerJoin(table, on string) *SQL {
	return s.joinOn(2, table, on)
}

// FULL JOIN ... ON ...
func (s *SQL) FullJoin(table, on string) *SQL {
	return s.joinOn(3, table, on)
}

var orderType = []string{"ASC ", "DESC "}

// 供Asc()和Desc()使用。
// sort: 0=asc,1=desc，其它值无效
func (s *SQL) orderBy(sort int, col string) *SQL {
	if sort != 1 && sort != 2 {
		s.errors = append(s.errors, fmt.Errorf("orderBy:错误的sort参数:[%v]", sort))
	}

	if s.order.Len() == 0 {
		s.order.WriteString("ORDER BY ")
	} else {
		s.order.WriteString(", ")
	}

	s.order.WriteString(col)
	s.order.WriteString(orderType[sort])

	return s
}

// ORDER BY ... ASC
func (s *SQL) Asc(cols ...string) *SQL {
	for _, c := range cols {
		s.orderBy(0, c)
	}

	return s
}

// ORDER BY ... DESC
func (s *SQL) Desc(cols ...string) *SQL {
	for _, c := range cols {
		s.orderBy(1, c)
	}

	return s
}

// LIMIT ... OFFSET ...
// offset值为0时，相当于limit N的效果。
func (s *SQL) Limit(limit, offset int) *SQL {
	s.limitSQL, s.limitArgs = s.db.Dialect().LimitSQL(limit, offset)
	return s
}

// 分页，调用Limit()实现，即Page()与Limit()方法会相互覆盖。
func (s *SQL) Page(start, size int) *SQL {
	if start < 1 {
		s.errors = append(s.errors, errors.New("Page:start必须大于0"))
	}
	if size < 1 {
		s.errors = append(s.errors, errors.New("Page:size必须大于0"))
	}

	start-- // 转到从0页开始
	return s.Limit(size, start*size)
}

// 产生SELECT语句
func (s *SQL) selectSQL() string {
	s.buf.Reset()

	s.buf.WriteString("SELECT ")
	s.buf.WriteString(strings.Join(s.cols, ","))
	s.buf.WriteString(" FROM ")
	s.buf.WriteString(s.tableName)
	s.buf.WriteString(s.join.String())
	s.buf.WriteString(s.cond.String())  // where
	s.buf.WriteString(s.order.String()) // NOTE(caixw):mysql中若要limit，order字段是必须提供的
	s.buf.WriteString(s.limitSQL)

	return s.db.PrepareSQL(s.buf.String())
}

// 功能同database/sql.DB.Query(...)
func (s *SQL) Query(args ...interface{}) (*sql.Rows, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	if len(args) == 0 {
		// 与selectSQL中添加的顺序相同，where在limit之前
		args = append(s.condArgs, s.limitArgs)
	}

	return s.db.Query(s.selectSQL(), args...)
}

// 功能同data/sql.DB.QueryRow(...)
func (s *SQL) QueryRow(args ...interface{}) *sql.Row {
	if s.HasErrors() {
		panic("构建语句时发生错误信息")
	}

	if len(args) == 0 {
		// 与sqlString中添加的顺序相同，where在limit之前
		args = append(s.condArgs, s.limitArgs)
	}

	return s.db.QueryRow(s.selectSQL(), args...)
}

// 导出数据到map[string]interface{}
func (s *SQL) Fetch2Map(args ...interface{}) (map[string]interface{}, error) {
	rows, err := s.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := fetch.Map(true, rows)
	if err != nil {
		return nil, err
	}

	return data[0], nil
}

// 导出所有数据到[]map[string]interface{}
func (s *SQL) Fetch2Maps(args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Map(false, rows)
}

// 返回指定列的第一行内容
func (s *SQL) FetchColumn(col string, args ...interface{}) (interface{}, error) {
	rows, err := s.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := fetch.Column(true, col, rows)
	if err != nil {
		return nil, err
	}

	return data[0], nil
}

// 返回指定列的所有数据
func (s *SQL) FetchColumns(col string, args ...interface{}) ([]interface{}, error) {
	rows, err := s.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Column(false, col, rows)
}

// 将当前select语句查询的数据导出到v中
// v可以是map[string]interface{}，或是orm/fetch.Obj中允许的类型。
func (s *SQL) Fetch(v interface{}, args ...interface{}) error {
	rows, err := s.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	switch vv.Kind() {
	case reflect.Map:
		if vv.Type().Key().Kind() != reflect.String {
			return errors.New("map的键名类型只能为string")
		}
		if vv.Type().Elem().Kind() != reflect.Interface {
			return errors.New("map的键值类型只能为interface{}")
		}

		mapped, err := s.Fetch2Map(args...)
		if err != nil {
			return err
		}
		vv.Set(reflect.ValueOf(mapped))
	default:
		return fetch.Obj(v, rows)
	}
	return nil
}

// 将当前语句预编译并缓存到stmts中，方便之后再次使用。
// action用于指定语句的类型，可以是Insert, Delete, Update或是Select。
func (s *SQL) Stmt(action int, name string) (*sql.Stmt, error) {
	var sql string
	switch action {
	case Delete:
		sql = s.db.PrepareSQL(s.deleteSQL())
	case Update:
		sql = s.db.PrepareSQL(s.updateSQL())
	case Insert:
		sql = s.db.PrepareSQL(s.insertSQL())
	case Select:
		sql = s.db.PrepareSQL(s.selectSQL())
	default:
		return nil, fmt.Errorf("无效的的action值[%v]", action)
	}

	return s.db.GetStmts().AddSQL(name, sql)
}

// 执行当前语句。
// action指定语句类型，可以是Delete,Insert或Update，但不能是Select
func (s *SQL) Exec(action int, args ...interface{}) (sql.Result, error) {
	switch action {
	case Delete:
		return s.Delete(args...)
	case Update:
		return s.Update(args...)
	case Insert:
		return s.Insert(args...)
	case Select:
		return nil, errors.New("select语句不能使用Exec()方法执行")
	default:
		return nil, fmt.Errorf("无效的的action值[%v]", action)
	}
}

// 产生delete语句
func (s *SQL) deleteSQL() string {
	s.buf.Reset()
	s.buf.WriteString("DELETE FROM ")
	s.buf.WriteString(s.tableName)

	// where
	s.buf.WriteString(s.cond.String())

	return s.db.PrepareSQL(s.buf.String())
}

// 执行DELETE操作。
// 相当于s.Exec(Delete, args...)
func (s *SQL) Delete(args ...interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	if len(args) == 0 {
		args = s.condArgs
	}

	return s.db.Exec(s.deleteSQL(), args...)
}

// 产生update语句
func (s *SQL) updateSQL() string {
	s.buf.Reset()
	s.buf.WriteString("UPDATE ")
	s.buf.WriteString(s.tableName)
	s.buf.WriteString(" SET ")
	for _, v := range s.cols {
		s.buf.WriteString(v)
		s.buf.WriteString("=?,")
	}
	s.buf.Truncate(s.buf.Len() - 1)

	// where
	s.buf.WriteString(s.cond.String())

	return s.db.PrepareSQL(s.buf.String())
}

// 执行UPDATE操作。
// 相当于s.Exec(Update, args...)
func (s *SQL) Update(args ...interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	if len(args) == 0 {
		args = append(s.vals, s.condArgs)
	}

	return s.db.Exec(s.updateSQL(), args...)
}

// 产生insert语句
func (s *SQL) insertSQL() string {
	s.buf.Reset() // 清空之前的内容

	s.buf.WriteString("INSERT INTO ")
	s.buf.WriteString(s.tableName)

	s.buf.WriteByte('(')
	s.buf.WriteString(strings.Join(s.cols, ","))
	s.buf.WriteString(") VALUES(")
	placeholder := strings.Repeat("?,", len(s.cols))
	// 去掉上面的最后一个逗号
	s.buf.WriteString(placeholder[0 : len(placeholder)-1])
	s.buf.WriteByte(')')

	return s.db.PrepareSQL(s.buf.String())
}

// 执行INSERT操作。
// 相当于s.Exec(Insert, args...)
func (s *SQL) Insert(args ...interface{}) (sql.Result, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	if len(args) == 0 {
		args = s.vals
	}

	return s.db.Exec(s.insertSQL(), args...)
}
