// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"
)

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk;len(100)"`

	Regdate int64 `orm:"-"`
}

type FetchUser struct {
	FetchEmail
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"index(username_index);len(20)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`
}

type Log struct {
	ID      int        `orm:"name(id);ai"`
	Content string     `orm:"name(content);len(1024)"`
	Created int        `orm:"name(created)"`
	UID     int        `orm:"name(uid)"`
	User    *FetchUser `orm:"name(user)"`
}

func TestParseObject(t *testing.T) {
	a := assert.New(t, false)
	obj := &Log{ID: 5}
	mapped := map[string]reflect.Value{}

	v := reflect.ValueOf(obj).Elem()
	a.True(v.IsValid())

	a.NotError(parseObject(v, &mapped))
	a.Equal(8, len(mapped), "长度不相等，导出元素为:[%v]", mapped)

	// 忽略的字段
	_, found := mapped["user.Regdate"]
	a.False(found)

	// 判断字段是否存在
	vi, found := mapped["id"]
	a.True(found).True(vi.IsValid())

	// 设置字段的值
	mapped["user.id"].Set(reflect.ValueOf(36))
	a.Equal(36, obj.User.ID)
	mapped["user.Email"].SetString("email")
	a.Equal("email", obj.User.Email)
	mapped["user.Username"].SetString("username")
	a.Equal("username", obj.User.Username)
	mapped["user.group"].SetInt(1)
	a.Equal(1, obj.User.Group)

	type m struct {
		*FetchEmail
		ID int
	}
	o := &m{ID: 5}
	mapped = map[string]reflect.Value{}
	v = reflect.ValueOf(o).Elem()
	a.NotError(parseObject(v, &mapped))
	a.Equal(2, len(mapped), "长度不相等，导出元素为:[%v]", mapped)

	type mm struct {
		FetchEmail
		ID int
	}
	oo := &mm{ID: 5}
	mapped = map[string]reflect.Value{}
	v = reflect.ValueOf(oo).Elem()
	a.NotError(parseObject(v, &mapped))
	a.Equal(2, len(mapped), "长度不相等，导出元素为:[%v]", mapped)
}

func TestGetColumns(t *testing.T) {
	a := assert.New(t, false)
	obj := &FetchUser{}

	cols, err := getColumns(reflect.ValueOf(obj), []string{"id"})
	a.NotError(err).NotNil(cols)
	a.Equal(len(cols), 1)

	// 当列不存在数据模型时
	cols, err = getColumns(reflect.ValueOf(obj), []string{"id", "not-exists"})
	a.NotError(err).NotNil(cols)
	a.Equal(len(cols), 2)
}
