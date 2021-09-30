// SPDX-License-Identifier: MIT

package orm

import (
	"fmt"

	"github.com/issue9/orm/v4/sqlbuilder"
)

// Upgrader 更新数据表对象
//
// 主要针对线上数据与本地模型不相同时，可执行的一些操作。
// 并没有准确的方法判断线上字段与本地定义的是否相同，比如：
// varchar(50) 和 text 在 sqlite3 是相同的，但是在其它数据库可能是稍微有差别的，
// 所以 Upgrader 并不会自动对数据表进行更新，所有更新还是要手动调用相关的函数。
type Upgrader struct {
	db    *DB
	model *Model
	err   error
	ddl   []sqlbuilder.DDLSQLer
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
		db:    db,
		model: m,
		ddl:   make([]sqlbuilder.DDLSQLer, 0, 10),
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

		sql := sqlbuilder.AddColumn(u.DB()).
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
			sql := sqlbuilder.DropColumn(u.DB()).Table(u.model.Name).Column(n)
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
		for name, fk := range u.model.ForeignKeys {
			if name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.DB())
			sql.Table(u.model.Name)
			sql.FK(name, fk.Column.Name, fk.RefTableName, fk.RefColName, fk.UpdateRule, fk.DeleteRule)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}

		for name, unique := range u.model.Uniques {
			if name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.DB())
			sql.Table(u.model.Name)
			cols := make([]string, 0, len(unique))
			for _, col := range unique {
				cols = append(cols, col.Name)
			}
			sql.Unique(name, cols...)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}

		for name, expr := range u.model.Checks {
			if name != c {
				continue
			}

			sql := sqlbuilder.AddConstraint(u.DB())
			sql.Table(u.model.Name)
			sql.Check(name, expr)

			u.ddl = append(u.ddl, sql)
			continue LOOP
		}
	}

	return u
}

// DropConstraint 删除约束
func (u *Upgrader) DropConstraint(conts ...string) *Upgrader {
	if u.err == nil {
		for _, n := range conts {
			sql := sqlbuilder.DropConstraint(u.DB()).Table(u.model.Name).Constraint(n)
			u.ddl = append(u.ddl, sql)
		}
	}
	return u
}

// AddPK 添加主键约束
func (u *Upgrader) AddPK() *Upgrader {
	if u.err != nil {
		return u
	}

	if u.model.PrimaryKey != nil {
		cols := make([]string, 0, len(u.model.PrimaryKey))
		for _, col := range u.model.PrimaryKey {
			cols = append(cols, col.Name)
		}

		sql := sqlbuilder.AddConstraint(u.DB()).Table(u.model.Name).PK(cols...)
		u.ddl = append(u.ddl, sql)
	}

	return u
}

// DropPK 删除主键约束
func (u *Upgrader) DropPK() *Upgrader {
	if u.err == nil {
		sql := sqlbuilder.DropConstraint(u.DB()).Table(u.model.Name).PK()
		u.ddl = append(u.ddl, sql)
	}
	return u
}

// AddIndex 添加索引信息
func (u *Upgrader) AddIndex(name ...string) *Upgrader {
	if u.err != nil {
		return u
	}

	for _, index := range name {
		sql := sqlbuilder.CreateIndex(u.DB())
		sql.Table(u.model.Name)

		for name, cols := range u.model.Indexes {
			if name != index {
				continue
			}

			cs := make([]string, 0, len(cols))
			for _, c := range cols {
				cs = append(cs, c.Name)
			}

			sql.Name(name)
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
			sql := sqlbuilder.DropIndex(u.DB()).Table(u.model.Name).Name(index)
			u.ddl = append(u.ddl, sql)
		}
	}
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

	for _, ddl := range u.ddl {
		query, err := ddl.DDLSQL()
		if err != nil {
			return err
		}

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

		for _, q := range query {
			if _, err := e.Exec(q); err != nil {
				if err2 := rollback(); err2 != nil {
					err = fmt.Errorf("返回错误 %w 时再次发生错误 %s", err, err2.Error())
				}
				return err
			}
		}

		if err := commit(); err != nil {
			return err
		}
	}

	return nil
}
