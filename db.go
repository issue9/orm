// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

// 两个实现了core.DB接口的结构：Engine和Tx。
// 一个实现普通的数据库操作，一个用于事务的操作。

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/issue9/orm/core"
)

type Engine struct {
	name       string // 数据库的名称
	prefix     string // 表名前缀
	driverName string
	d          core.Dialect
	db         *sql.DB
	stmts      *core.Stmts
	sql        *SQL // 内置的SQL引擎，用于执行Update等操作
	replacer   *strings.Replacer
}

func newEngine(driverName, dataSourceName, prefix string) (*Engine, error) {
	d, found := getDialect(driverName)
	if !found {
		return nil, fmt.Errorf("newEngine:未找到与driverName[%v]相同的Dialect", driverName)
	}

	dbInst, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	l, r := d.QuoteStr()
	inst := &Engine{
		d:          d,
		prefix:     prefix,
		driverName: driverName,
		db:         dbInst,
		name:       d.GetDBName(dataSourceName),
		replacer: strings.NewReplacer(
			core.QuoteLeft, l,
			core.QuoteRight, r,
			core.TableNamePrefix, prefix,
		),
	}
	inst.stmts = core.NewStmts(inst)
	inst.sql = inst.SQL()

	return inst, nil
}

// 对orm/core.DB.Name()的实现，返回当前操作的数据库名称。
func (e *Engine) Name() string {
	return e.name
}

// 对orm/core.DB.GetStmts()的实现，返回当前的sql.Stmt实例缓存容器。
func (e *Engine) GetStmts() *core.Stmts {
	return e.stmts
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (e *Engine) Dialect() core.Dialect {
	return e.d
}

// 去掉占位符。
func (e *Engine) prepareSQL(sql string) string {
	return e.replacer.Replace(sql)
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (e *Engine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	sql = e.prepareSQL(sql)
	r, err := e.db.Exec(sql, args...)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, e.driverName, sql, args...)
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (e *Engine) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	sql = e.prepareSQL(sql)
	r, err := e.db.Query(sql, args...)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, e.driverName, sql, args...)
}

// 对orm/core.DB.QueryRow()的实现。
// 执行一条查询语句，并返回第一条符合条件的记录。
func (e *Engine) QueryRow(sql string, args ...interface{}) *sql.Row {
	return e.db.QueryRow(e.prepareSQL(sql), args...)
}

// 对orm/core.DB.Prepare()的实现。预处理SQL语句成sql.Stmt实例。
func (e *Engine) Prepare(sql string) (*sql.Stmt, error) {
	sql = e.prepareSQL(sql)
	r, err := e.db.Prepare(sql)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, e.driverName, sql)
}

// 关闭当前的Engine，销毁所有的数据。不能再次使用。
// 与之关联的Tx也将不能使用。
func (e *Engine) Close() error {
	e.stmts.Close()
	return e.db.Close()
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
	}
	ret.sql = ret.SQL()
	return ret, nil
}

// 查找缓存的sql.Stmt，在未找到的情况下，第二个参数返回false
func (e *Engine) Stmt(name string) (*sql.Stmt, bool) {
	return e.stmts.Get(name)
}

// 产生一个SQL实例。
func (e *Engine) SQL() *SQL {
	return newSQL(e)
}

// 指定一个条件语句，并返回SQL实例。
// 与SQL()方法稍微有一点不同，Where()会使用已缓存的SQL实例，
// 而SQL()方法会新声明一个SQL实例。
func (e *Engine) Where(cond string, args ...interface{}) *SQL {
	return e.sql.Reset().Where(cond, args...)
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
func (e *Engine) Create(models ...interface{}) (err error) {
	return createMult(e, models...)
}

// 根据models更新或创建表。
func (e *Engine) Upgrade(models ...interface{}) (err error) {
	return upgradeMult(e, models...)
}

// 事务对象
type Tx struct {
	engine *Engine
	tx     *sql.Tx
	sql    *SQL
}

// 对orm/core.DB.Name()的实现，返回当前操作的数据库名称。
func (t *Tx) Name() string {
	return t.engine.Name()
}

// 对orm/core.DB.GetStmts()的实现，返回当前的sql.Stmt实例缓存容器。
func (t *Tx) GetStmts() *core.Stmts {
	return t.engine.GetStmts()
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (t *Tx) Dialect() core.Dialect {
	return t.engine.Dialect()
}

// 去掉占位符。
func (t *Tx) prepareSQL(sql string) string {
	return t.engine.replacer.Replace(sql)
}

// 对orm/core.DB.Exec()的实现。执行一条非查询的SQL语句。
func (t *Tx) Exec(sql string, args ...interface{}) (sql.Result, error) {
	sql = t.prepareSQL(sql)
	r, err := t.tx.Exec(sql, args...)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, t.engine.driverName, sql, args...)
}

// 对orm/core.DB.Query()的实现，执行一条查询语句。
func (t *Tx) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	sql = t.prepareSQL(sql)
	r, err := t.tx.Query(sql, args...)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, t.engine.driverName, sql, args...)
}

// 对orm/core.DB.QueryRow()的实现。
// 执行一条查询语句，并返回第一条符合条件的记录。
func (t *Tx) QueryRow(sql string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(t.prepareSQL(sql), args...)
}

// 对orm/core.DB.Prepare()的实现。预处理SQL语句成sql.Stmt实例。
func (t *Tx) Prepare(sql string) (*sql.Stmt, error) {
	sql = t.prepareSQL(sql)
	r, err := t.tx.Prepare(sql)

	if err == nil {
		return r, nil
	}
	return nil, newSQLError(err, t.engine.driverName, sql)
}

// 关闭当前的db。
// 不会关闭与之关联的engine实例，
// 仅是取消了与之的关联。
func (t *Tx) Close() {
	t.engine = nil
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
func (t *Tx) Stmt(name string) (*sql.Stmt, bool) {
	stmt, found := t.engine.Stmt(name)
	if !found {
		return nil, false
	}

	return t.tx.Stmt(stmt), true
}

// 返回一个新的SQL实例。
func (t *Tx) SQL() *SQL {
	return newSQL(t)
}

// 指定一个条件语句，并返回SQL实例。
// 与SQL()方法稍微有一点不同，Where()会使用已缓存的SQL实例，
// 而SQL()方法会新声明一个SQL实例。
func (t *Tx) Where(cond string, args ...interface{}) *SQL {
	return t.sql.Reset().Where(cond, args...)
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

// 更新或是创建数据表。
func (t *Tx) Upgrade(v ...interface{}) error {
	return upgradeMult(t, v...)
}

// sql错误信息
type SQLError struct {
	Err        error
	DriverName string
	SQL        string
	Args       []interface{}
}

func newSQLError(err error, driverName, sql string, args ...interface{}) error {
	return &SQLError{
		Err:        err,
		DriverName: driverName,
		SQL:        sql,
		Args:       args,
	}
}

func (err *SQLError) Error() string {
	format := "SQLError:原始错误信息:%v;\ndriverName:%v;\nsql语句:%v;\n对应参数:%v"
	return fmt.Sprintf(format, err.Err, err.DriverName, err.SQL, err.Args)
}

// 供engine.go和tx.go调用的一系列函数。

// 要怕model中的主键或是唯一索引产生where语句，
// 若两者都不存在，则返回错误信息。
// rval为struct的reflect.Value
func where(sql *SQL, m *core.Model, rval reflect.Value) error {
	switch {
	case len(m.PK) != 0:
		for _, col := range m.PK {
			field := rval.FieldByName(col.GoName)
			if !field.IsValid() {
				return fmt.Errorf("where:未找到该名称[%v]的值", col.GoName)
			}
			sql.Where(col.Name+"=?", field.Interface())
		}
	case len(m.UniqueIndexes) != 0:
		for _, cols := range m.UniqueIndexes {
			for _, col := range cols {
				field := rval.FieldByName(col.GoName)
				if !field.IsValid() {
					return fmt.Errorf("where:未找到该名称[%v]的值", col.GoName)
				}
				sql.Where(col.Name+"=?", field.Interface())
			}
			break // 只取一个UniqueIndex就可以了
		}
	default:
		return errors.New("where:无法产生where部分语句")
	}

	return nil
}

// 创建或是更新一个数据表。
// v为一个结构体或是结构体指针。
func createOne(db core.DB, onlyCreate bool, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("createOne:无效的v.Kind()")
	}

	return db.Dialect().UpgradeTable(db, m, onlyCreate)
}

// 插入一个对象到数据库
// 以v中的主键或是唯一索引作为where条件语句。
// 自增字段，即使指定了值，也不会被添加
func insertOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("insertOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	for name, col := range m.Cols {
		if col.IsAI() { // AI过滤
			continue
		}

		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("insertOne:未找到该名称[%v]的值", col.GoName)
		}
		sql.Add(name, field.Interface())
	}

	_, err = sql.Insert()
	return err
}

// 更新一个对象
// 以v中的主键或是唯一索引作为where条件语句，其它值为更新值
func updateOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("updateOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("updateOne:未找到该名称[%v]的值", col.GoName)
		}
		sql.Add(name, field.Interface())
	}

	_, err = sql.Update()
	return err
}

// 删除v表示的单个对象的内容
// 以v中的主键或是唯一索引作为where条件语句
func deleteOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("deleteOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = sql.Delete()
	return err
}

// 创建一个或多个数据表
func createMult(db core.DB, objs ...interface{}) (err error) {
	for _, obj := range objs {
		if err = createOne(db, true, obj); err != nil {
			return
		}
	}

	return
}

// 创建或是更新一个或多个数据表
func upgradeMult(db core.DB, objs ...interface{}) (err error) {
	for _, obj := range objs {
		if err = createOne(db, false, obj); err != nil {
			return
		}
	}

	return
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func insertMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return insertOne(sql, v)
	case reflect.Slice, reflect.Array:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("insertMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := insertOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("insertMult:v的类型[%v]无效", rval.Kind())
	}

	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func updateMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return updateOne(sql, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("updateMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := updateOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("updateMult:v的类型无效")
	}

	return nil
}

// 删除指定的数据对象。
func deleteMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return deleteOne(sql, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("deleteMult:数组元素类型不正确,只能是指针或是struct的指针")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := deleteOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("deleteMult:v的类型无效")
	}

	return nil
}
