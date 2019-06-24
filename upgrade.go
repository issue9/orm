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

	addCols  []*Column
	dropCols []string

	addConts  map[string][]string
	dropConts []string

	addIdxs  map[string][]string
	dropIdxs []string

	// TODO 修改列信息
}

// Upgrade 生成 Upgrader 对象
func (db *DB) Upgrade(v interface{}) (*Upgrader, error) {
	m, err := db.NewModel(v)
	if err != nil {
		return nil, err
	}

	return &Upgrader{
		db:        db,
		model:     m,
		addCols:   []*Column{},
		dropCols:  []string{},
		dropConts: []string{},
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
	if u.err != nil {
		return u
	}

	for _, n := range name {
		col := u.model.FindColumn(n)
		if col == nil {
			panic(fmt.Sprintf("列名 %s 不存在", n))
		}

		u.addCols = append(u.addCols, col)
	}

	return u
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
			if err1 := rollback(); err1 != nil {
				return errors.New(err1.Error() + err.Error())
			}
			return err
		}
	}

	if len(u.dropConts) > 0 {
		if err := u.dropConstraints(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return errors.New(err1.Error() + err.Error())
			}
			return err
		}
	}

	if len(u.addCols) > 0 {
		if err := u.addColumns(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return errors.New(err1.Error() + err.Error())
			}
			return err
		}
	}

	// TODO

	return commit()
}

func (u *Upgrader) addColumns(e Engine) error {
	sql := sqlbuilder.AddColumn(e)

	for _, col := range u.addCols {
		sql.Reset()
		typ, err := u.DB().Dialect().SQLType(col)
		if err != nil {
			return err
		}

		_, err = sql.Table(u.model.Name).
			Column(col.Name, typ).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
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