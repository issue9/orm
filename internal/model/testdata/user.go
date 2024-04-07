// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package testdata

import "github.com/issue9/orm/v6/core"

// User 与 model.User 完全相同，包括表名
//
// 但在不同的目录下，类型不同，应该会被创建不同的 core.Model 对象。
type User struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_user_username);index(index_user_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

func (u *User) TableName() string { return "users" }

func (u *User) ApplyModel(m *core.Model) error {
	m.Options["mysql_engine"] = []string{"innodb"}
	m.Options["mysql_charset"] = []string{"utf8"}
	return nil
}
