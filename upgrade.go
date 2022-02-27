// SPDX-License-Identifier: MIT

package orm

import (
	"fmt"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/sqlbuilder"
)

// Upgrader 更新数据表对象
//
// 主要针对线上数据与本地模型不相同时，可执行的一些操作。
// 并没有准确的方法判断线上字段与本地定义的是否相同，比如：
// varchar(50) 和 text 在 sqlite3 是相同的，但是在其它数据库可能是稍微有差别的，
// 所以 Upgrader 并不会自动对数据表进行更新，所有更新还是要手动调用相关的函数。
type Upgrader struct {
	model *core.Model
	err   error
	ddl   []sqlbuilder.DDLStmt

	e        Engine
	commit   func() error
	rollback func() error
}

// Upgrade 生成 Upgrader 对象
//
// Upgrader 提供了对现有的数据模型 v 与线上数据表之间的操作。
// 删除操作需要保证已经存在于数据表；
// 而添加操作需要保证已经存在于模型 v，又不存在于数据表。
func (db *DB) Upgrade(v TableNamer) (*Upgrader, error) {
	m, err := db.newModel(v)
	if err != nil {
		return nil, err
	}

	var e Engine = db
	commit := func() error { return nil }
	rollback := func() error { return nil }

	if db.Dialect().TransactionalDDL() {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		e = tx

		commit = func() error { return tx.Commit() }
		rollback = func() error { return tx.Rollback() }
	}

	return &Upgrader{
		model: m,
		ddl:   make([]sqlbuilder.DDLStmt, 0, 10),

		e:        e,
		commit:   commit,
		rollback: rollback,
	}, nil
}

func (u *Upgrader) Engine() Engine { return u.e }

// Err 返回执行过程中的错误信息
func (u *Upgrader) Err() error { return u.err }

// AddColumn 添加表中的列
//
// 列名必须存在于表模型中。
func (u *Upgrader) AddColumn(name ...string) *Upgrader {
	for _, n := range name {
		if u.err != nil {
			return u
		}

		col := u.model.FindColumn(n)
		if col == nil {
			u.err = fmt.Errorf("列名 %s 不存在", n)
			return u
		}

		if !col.HasDefault && !col.Nullable {
			u.err = fmt.Errorf("新增列 %s 必须指定默认值或是 Nullable 属性", col.Name)
			return u
		}

		sql := sqlbuilder.AddColumn(u.Engine()).
			Table(u.model.Name).
			Column(col.Name, col.PrimitiveType, col.AI, col.Nullable, col.HasDefault, col.Default, col.Length...)
		u.ddl = append(u.ddl, sql)
	}

	return u
}

// DropColumn 删除表中的列
//
// 列名可以不存在于表模型，只在数据库中的表包含该列名，就会被删除。
func (u *Upgrader) DropColumn(name ...string) *Upgrader {
	if u.err == nil {
		for _, n := range name {
			sql := sqlbuilder.DropColumn(u.Engine()).Table(u.model.Name).Column(n)
			u.ddl = append(u.ddl, sql)
		}
	}
	return u
}

// AddConstraint 添加约束
//
// 约束必须存在于 model.constraints 中
func (u *Upgrader) AddConstraint(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

LOOP:
	for _, c := range name {
		for _, fk := range u.model.ForeignKeys {
			if fk.Name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.Engine())
			sql.Table(u.model.Name)
			sql.FK(fk.Name, fk.Column.Name, fk.RefTableName, fk.RefColName, fk.UpdateRule, fk.DeleteRule)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}

		for _, uu := range u.model.Uniques {
			if uu.Name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.Engine())
			sql.Table(u.model.Name)
			cols := make([]string, 0, len(uu.Columns))
			for _, col := range uu.Columns {
				cols = append(cols, col.Name)
			}
			sql.Unique(uu.Name, cols...)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}

		for name, expr := range u.model.Checks {
			if name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.Engine())
			sql.Table(u.model.Name)
			sql.Check(name, expr)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}

		if u.model.PrimaryKey.Name == c {
			cols := make([]string, 0, len(u.model.PrimaryKey.Columns))
			for _, col := range u.model.PrimaryKey.Columns {
				cols = append(cols, col.Name)
			}

			sql := sqlbuilder.AddConstraint(u.Engine()).Table(u.model.Name).PK(c, cols...)
			u.ddl = append(u.ddl, sql)
		}
	}

	return u
}

// DropConstraint 删除约束
func (u *Upgrader) DropConstraint(conts ...string) *Upgrader {
	if u.err == nil {
		for _, n := range conts {
			sql := sqlbuilder.DropConstraint(u.Engine()).Table(u.model.Name).Constraint(n)
			u.ddl = append(u.ddl, sql)
		}
	}
	return u
}

// DropPK 删除主键约束
func (u *Upgrader) DropPK(name string) *Upgrader {
	if u.err == nil {
		sql := sqlbuilder.DropConstraint(u.Engine()).Table(u.model.Name).PK(name)
		u.ddl = append(u.ddl, sql)
	}
	return u
}

// AddIndex 添加索引信息
func (u *Upgrader) AddIndex(name ...string) *Upgrader {
	if u.Err() != nil {
		return u
	}

	for _, index := range name {
		sql := sqlbuilder.CreateIndex(u.Engine())
		sql.Table(u.model.Name)

		for _, i := range u.model.Indexes {
			if i.Name != index {
				continue
			}

			cs := make([]string, 0, len(i.Columns))
			for _, c := range i.Columns {
				cs = append(cs, c.Name)
			}

			sql.Name(i.Name)
			sql.Columns(cs...)
		}

		u.ddl = append(u.ddl, sql)
	}

	return u
}

// DropIndex 删除索引
func (u *Upgrader) DropIndex(name ...string) *Upgrader {
	if u.Err() == nil {
		for _, index := range name {
			sql := sqlbuilder.DropIndex(u.Engine()).Table(u.model.Name).Name(index)
			u.ddl = append(u.ddl, sql)
		}
	}
	return u
}

// Do 执行操作
func (u *Upgrader) Do() error {
	if u.Err() != nil {
		return u.Err()
	}

	for _, ddl := range u.ddl {
		if err := ddl.Exec(); err != nil {
			if err2 := u.rollback(); err2 != nil {
				err = fmt.Errorf("%w,%s", err, err2)
			}
			return err
		}
	}

	return u.commit()
}
