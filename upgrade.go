// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
)

// Upgrader 提供了对更新表结构的一些操作
type Upgrader struct {
	db     *DB
	engine Engine
	model  *Model
	err    error
}

// Upgrade 生成 Upgrader 对象
func (db *DB) Upgrade(v interface{}) (*Upgrader, error) {
	m, err := db.NewModel(v)
	if err != nil {
		return nil, err
	}

	u := &Upgrader{
		db:    db,
		model: m,
	}

	if db.Dialect().TransactionalDDL() {
		u.engine, err = db.Begin()
		if err != nil {
			return nil, err
		}
	} else {
		u.engine = db
	}

	return u, nil
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

// DropColumn 删除表中的列，列名可以不存在于表模型，
// 只在数据库中的表包含该列名，就会被删除。
func (u *Upgrader) DropColumn(name ...string) *Upgrader {
	for _, n := range name {
		u.dropColumn(n)
	}

	return u
}

func (u *Upgrader) dropColumn(name string) {
	// TODO
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
func (u *Upgrader) DropConstraint() {
	//
}

// Do 执行操作
func (u *Upgrader) Do() error {
	if u.Err() != nil {
		return u.Err()
	}

	// TODO
	return nil
}
