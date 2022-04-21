// SPDX-License-Identifier: MIT

package model

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/sqlbuilder"
)

var (
	_ core.PrimitiveTyper = &last{}
	_ sql.Scanner         = &last{}
)

type last struct {
	IP      string
	Created int64
}

func (l *last) Scan(v any) error {
	vv, ok := v.(string)
	if !ok {
		return errors.New("无效的类型")
	}

	vals := strings.Split(vv, ",")
	if len(vals) != 2 {
		return errors.New("无效的值格式")
	}

	cc, err := time.Parse(core.TimeFormatLayout, vals[1])
	if err != nil {
		return err
	}

	l.IP = vals[0]
	l.Created = cc.Unix()
	return nil
}

func (l *last) PrimitiveType() core.PrimitiveType {
	return core.String
}

// User 带自增和一个唯一约束
type User struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_user_username);index(index_user_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

type Admin1 struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_admin_username);index(index_admin_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

func (u *User) TableName() string { return "users" }

func (u *User) ApplyModel(m *core.Model) error {
	m.Options["mysql_engine"] = []string{"innodb"}
	m.Options["mysql_charset"] = []string{"utf8"}
	return nil
}

// Admin 带自增和两个唯一约束
type Admin struct {
	Admin1
	Email string `orm:"name(email);len(20);unique(unique_admin_email)"`
	Group int64  `orm:"name(group);fk(fk_admin_name,groups,id,NO ACTION)"`
}

func (a *Admin) TableName() string { return "administrators" }

func (a *Admin) ApplyModel(m *core.Model) error {
	m.Options["mysql_engine"] = []string{"innodb"}
	m.Options["mysql_charset"] = []string{"utf8"}
	return m.NewCheck("admin_chk_name", "{group}>0")
}

type obj struct {
	ID int `orm:"name(id);ai"`
}

func (o obj) TableName() string { return "objs" }

type viewObject struct {
	ID       int    `orm:"name(id);ai"`
	Username string `orm:"name(username)"`
}

func (v *viewObject) TableName() string { return "view_objects" }

func (v *viewObject) ViewAs(e core.Engine) (string, error) {
	return sqlbuilder.Select(e).
		Columns("id", "username").
		From("User").
		Where("id>?", 10).
		CombineSQL()
}

func TestModels_New(t *testing.T) {
	a := assert.New(t, false)
	ms := NewModels(nil)
	a.NotNil(ms)

	m, err := ms.New("", &Admin{})
	a.NotError(err).NotNil(m)
	a.Equal(m.Type, core.Table).
		Empty(m.ViewAs)

	// cols
	idCol := m.FindColumn("id") // 指定名称为小写
	a.NotNil(idCol)

	usernameCol := m.FindColumn("Username") // 未指定别名，与字段名相同
	a.NotNil(usernameCol).False(usernameCol.Nullable)

	// 通过 struct tag 过滤掉的列
	regdate := m.FindColumn("Regdate")
	a.Nil(regdate)

	groupCol := m.FindColumn("group")
	a.NotNil(groupCol)

	// index
	index, found := m.Index("index_admin_name")
	a.True(found).Equal(usernameCol, index.Columns[0])

	// ai
	a.Equal(m.AutoIncrement, idCol).
		Nil(m.PrimaryKey) // 有自增，则主键为空

	// unique_name
	unique, found := m.Unique("unique_admin_username")
	a.True(found).Equal(unique.Columns[0], usernameCol)

	fk, found := m.ForeignKey("fk_admin_name")
	a.True(found).
		Equal(fk.Name, "fk_admin_name").
		Equal(fk.Column, groupCol).
		Equal(fk.RefTableName, "groups").
		Equal(fk.RefColName, "id").
		Equal(fk.UpdateRule, "NO ACTION").
		Equal(fk.DeleteRule, "")

	// check
	chk, found := m.Checks["admin_chk_name"]
	a.True(found).Equal(chk, "{group}>0")

	// options
	a.Equal(m.Options, map[string][]string{
		"mysql_engine":  {"innodb"},
		"mysql_charset": {"utf8"},
	})

	a.Equal(m.Name, "administrators")

	// 多层指针下的 Receive

	o := obj{}
	m, err = ms.New("", o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "objs")
	a.Equal(len(ms.models), 2)

	m, err = ms.New("", &o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "objs")
	a.Equal(len(ms.models), 2)

	m, err = ms.New("p_", &o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "p_objs")
	a.Equal(len(ms.models), 3)

	// view
	m, err = ms.New("", &viewObject{})
	a.NotError(err).NotNil(m)
	a.Equal(m.Type, core.View).
		NotNil(m.ViewAs)
}

func TestModel_setOCC(t *testing.T) {
	a := assert.New(t, false)
	m := core.NewModel(core.Table, "m1", 10)

	c, err := core.NewColumn(core.Int)
	a.NotError(err).NotNil(c)
	col := &column{Column: c}
	col.Name = "occ"
	a.NotError(m.AddColumn(c))

	a.NotError(setOCC(m, col, nil))
	a.Equal(col.Column, m.OCC)

	// m.OCC 已经存在
	a.Error(setOCC(m, col, nil))

	// 太多的值，occ(true)
	m.OCC = nil
	a.Error(setOCC(m, col, []string{"true"}))
}

func TestModel_setPK(t *testing.T) {
	a := assert.New(t, false)
	m := core.NewModel(core.Table, "m1", 10)
	a.NotNil(m)
	c, err := core.NewColumn(core.Int8)
	a.NotError(err).NotNil(c)

	// 过多的参数
	col := &column{Column: c}
	a.Error(setPK(m, col, []string{"123"}))
}

func TestModel_setAI(t *testing.T) {
	a := assert.New(t, false)
	m := core.NewModel(core.Table, "m1", 10)
	a.NotNil(m)

	c, err := core.NewColumn(core.Int)
	a.NotError(err).NotNil(c)
	col := &column{Column: c}

	// 太多的参数
	a.Error(col.setAI(m, []string{"true", "false"}))

	// 列类型只能是整数型
	c, err = core.NewColumn(core.Float32)
	a.NotError(err).NotNil(c)
	col = &column{Column: c}
	a.Error(col.setAI(m, nil))

	c, err = core.NewColumn(core.Int)
	a.NotError(err).NotNil(c)
	col = &column{Column: c}
	col.Name = "ai"
	a.NotError(m.AddColumn(col.Column))
	a.NotError(col.setAI(m, nil))
}
