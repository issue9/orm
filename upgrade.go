// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"errors"
	"fmt"

	"github.com/issue9/orm/v2/sqlbuilder"
)

// Upgrader 提供了对更新表结构的一些操作
type Upgrader struct {
	db    *DB
	model *Model
	err   error

	dropCols  []string
	dropConts []string
}

// Upgrade 生成 Upgrader 对象
func (db *DB) Upgrade(v interface{}) (*Upgrader, error) {
	m, err := db.NewModel(v)
	if err != nil {
		return nil, err
	}

	return &Upgrader{
		db:    db,
		model: m,
	}, nil
}

// DB 返回关联的 DB 实例
func (u *Upgrader) DB() *DB {
	return u.DB()
}

// Err 返回执行过程中的错误信息
func (u *Upgrader) Err() error {
	return u.err
}

// AddColumn 添加表中的列，列名必须存在于表模型中
func (u *Upgrader) AddColumn(name ...string) *Upgrader {
	for _, n := range name {
		u.AddColumn(n)
	}

	return u
}

func (u *Upgrader) addColumn(name string) {
	if u.err != nil {
		return
	}

	col := u.model.FindColumn(name)
	if col == nil {
		panic(fmt.Sprintf("列名 %s 不存在", name))
	}

	// TODO
}

// DropColumns 删除表中的列，列名可以不存在于表模型，
// 只在数据库中的表包含该列名，就会被删除。
func (u *Upgrader) DropColumns(name ...string) *Upgrader {
	u.dropCols = append(u.dropCols, name...)
	return u
}

// RenameColumn 修改列名
func (u *Upgrader) RenameColumn(old, new string) {
	// TODO
}

// ChangeColumn 改变列属性
func (u *Upgrader) ChangeColumn(old, new string, conv func(interface{}) interface{}) {

}

// AddConstraint 添加约束
//
// 约束必须存在于 model.constraints 中
func (u *Upgrader) AddConstraint() {
	//
}

// DropConstraint 删除约束
func (u *Upgrader) DropConstraint(conts ...string) *Upgrader {
	u.dropConts = append(u.dropConts, conts...)
	return u
}

// Do 执行操作
func (u *Upgrader) Do() error {
	if u.Err() != nil {
		return u.Err()
	}

	rollback := func() error {
		return nil
	}

	commit := func() error {
		return nil
	}

	var e Engine = u.DB()
	if u.DB().Dialect().TransactionalDDL() {
		tx, err := u.DB().Begin()
		if err != nil {
			return err
		}

		rollback = func() error {
			return tx.Rollback()
		}
		commit = func() error {
			return tx.Commit()
		}
		e = tx
	}

	if len(u.dropCols) > 0 {
		if err := u.dropColumns(e); err != nil {
			err1 := rollback()
			return errors.New(err1.Error() + err.Error())
		}
	}

	// TODO

	return commit()
}

func (u *Upgrader) dropConstraints(e Engine) error {
	sql := sqlbuilder.DropConstraint(e)

	for _, n := range u.dropCols {
		sql.Reset()
		_, err := sql.Table(u.model.Name).
			Constraint(n).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Upgrader) dropColumns(e Engine) error {
	sql := sqlbuilder.DropColumn(e)

	for _, n := range u.dropCols {
		sql.Reset()
		_, err := sql.Table(u.model.Name).
			Column(n).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
