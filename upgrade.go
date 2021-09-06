// SPDX-License-Identifier: MIT

package orm

import (
	"fmt"

	"github.com/issue9/orm/v4/sqlbuilder"
)

// Upgrader 更新数据库对象
//
// 主要针对线上数据与本地模型不相同时，可执行的一些操作。
type Upgrader struct {
	db    *DB
	model *Model
	err   error

	addCols  []*Column
	dropCols []string

	addConts  []string
	dropConts []string

	addIdxs  []string
	dropIdxs []string
}

// Upgrade 生成 Upgrader 对象
//
// Upgrader 提供了对现有的数据模型 v 与线上数据表之间的操作。
// 删除操作需要保证已经存在于数据表；
// 而添加操作需要保证已经存在于模型 v，又不存在于数据表。
func (db *DB) Upgrade(v TableNamer) (*Upgrader, error) {
	m, err := db.NewModel(v)
	if err != nil {
		return nil, err
	}

	return &Upgrader{
		db:        db,
		model:     m,
		addCols:   []*Column{},
		dropCols:  []string{},
		addConts:  []string{},
		dropConts: []string{},
		addIdxs:   []string{},
		dropIdxs:  []string{},
	}, nil
}

// DB 返回关联的 DB 实例
func (u *Upgrader) DB() *DB { return u.db }

// Err 返回执行过程中的错误信息
func (u *Upgrader) Err() error { return u.err }

// AddColumn 添加表中的列
//
// 列名必须存在于表模型中。
func (u *Upgrader) AddColumn(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	for _, n := range name {
		col := u.model.FindColumn(n)
		if col == nil {
			u.err = fmt.Errorf("列名 %s 不存在", n)
		}

		u.addCols = append(u.addCols, col)
	}

	return u
}

// DropColumn 删除表中的列
//
// 列名可以不存在于表模型，只在数据库中的表包含该列名，就会被删除。
func (u *Upgrader) DropColumn(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	if u.dropCols == nil {
		u.dropCols = name
		return u
	}

	u.dropCols = append(u.dropCols, name...)
	return u
}

// AddConstraint 添加约束
//
// 约束必须存在于 model.constraints 中
func (u *Upgrader) AddConstraint(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	if u.addConts == nil {
		u.addConts = name
		return u
	}

	u.addConts = append(u.addConts, name...)

	return u
}

// DropConstraint 删除约束
func (u *Upgrader) DropConstraint(conts ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	if u.dropConts == nil {
		u.dropConts = conts
		return u
	}

	u.dropConts = append(u.dropConts, conts...)
	return u
}

// AddIndex 添加索引信息
func (u *Upgrader) AddIndex(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	if u.addIdxs == nil {
		u.addIdxs = name
		return u
	}

	u.addIdxs = append(u.addIdxs, name...)
	return u
}

// DropIndex 删除索引
func (u *Upgrader) DropIndex(name ...string) *Upgrader {
	if u.Err() != nil {
		return u
	}

	if u.dropIdxs == nil {
		u.dropIdxs = name
		return u
	}

	u.dropIdxs = append(u.dropIdxs, name...)
	return u
}

// Do 执行操作
//
// 此操作会尽量在一个事务中完成，但是如果数据不支持 DDL 模式，则可能是多次提交。
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

	if len(u.dropConts) > 0 {
		if err := u.dropConstraints(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	if len(u.dropIdxs) > 0 {
		if err := u.dropIndexes(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	// 约束可能正好依赖被删除的列。
	// 所以要在删除约束之后，再删除列信息。
	if len(u.dropCols) > 0 {
		if err := u.dropColumns(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	// 先添加列，再添加约束和索引。后者可能依赖添加的列信息。
	if len(u.addCols) > 0 {
		if err := u.addColumns(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	if len(u.addConts) > 0 {
		if err := u.addConstraints(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	if len(u.addIdxs) > 0 {
		if err := u.addIndexes(e); err != nil {
			if err1 := rollback(); err1 != nil {
				return fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
			}
			return err
		}
	}

	return commit()
}

func (u *Upgrader) addColumns(e Engine) error {
	sql := sqlbuilder.AddColumn(e)

	for _, col := range u.addCols {
		if !col.HasDefault && !col.Nullable {
			return fmt.Errorf("新增列 %s 必须指定默认值或是 Nullable 属性", col.Name)
		}
		sql.Reset()

		err := sql.Table(u.model.Name).
			Column(col.Name, col.PrimitiveType, col.AI, col.Nullable, col.HasDefault, col.Default, col.Length...).
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
		err := sql.Table(u.model.Name).
			Column(n).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Upgrader) dropConstraints(e Engine) error {
	sql := sqlbuilder.DropConstraint(e)

	for _, n := range u.dropConts {
		sql.Reset()
		err := sql.Table(u.model.Name).
			Constraint(n).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Upgrader) addConstraints(e Engine) error {
	stmt := sqlbuilder.AddConstraint(e)

LOOP:
	for _, c := range u.addConts {
		stmt.Reset().Table(u.model.Name)

		for name, fk := range u.model.ForeignKeys {
			if name != c {
				continue
			}

			stmt.FK(name, fk.Column.Name, fk.RefTableName, fk.RefColName, fk.UpdateRule, fk.DeleteRule)
			if err := stmt.Exec(); err != nil {
				return err
			}
			continue LOOP
		}

		for name, u := range u.model.Uniques {
			if name != c {
				continue
			}

			cols := make([]string, 0, len(u))
			for _, col := range u {
				cols = append(cols, col.Name)
			}
			stmt.Unique(name, cols...)

			if err := stmt.Exec(); err != nil {
				return err
			}

			continue LOOP
		}

		if u.model.PrimaryKey != nil {
			cols := make([]string, 0, len(u.model.PrimaryKey))
			for _, col := range u.model.PrimaryKey {
				cols = append(cols, col.Name)
			}
			stmt.PK(cols...)

			if err := stmt.Exec(); err != nil {
				return err
			}
			continue LOOP
		}

		for name, expr := range u.model.Checks {
			if name != c {
				continue
			}

			stmt.Check(name, expr)
			if err := stmt.Exec(); err != nil {
				return err
			}
			continue LOOP
		}
	}

	return nil
}

func (u *Upgrader) addIndexes(e Engine) error {
	stmt := sqlbuilder.CreateIndex(e)

	for _, index := range u.addIdxs {
		stmt.Reset().Table(u.model.Name)

		for name, cols := range u.model.Indexes {
			if name != index {
				continue
			}

			cs := make([]string, 0, len(cols))
			for _, c := range cols {
				cs = append(cs, c.Name)
			}

			stmt.Name(name)
			stmt.Columns(cs...)
		}
		if err := stmt.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (u *Upgrader) dropIndexes(e Engine) error {
	stmt := sqlbuilder.DropIndex(e)

	for _, index := range u.dropIdxs {
		stmt.Reset().Table(u.model.Name).Name(index)
		if err := stmt.Exec(); err != nil {
			return err
		}
	}

	return nil
}
