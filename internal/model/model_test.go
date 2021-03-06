// SPDX-License-Identifier: MIT

package model

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/internal/flagtest"
	"github.com/issue9/orm/v3/sqlbuilder"
)

var (
	_ core.DefaultParser  = &last{}
	_ core.PrimitiveTyper = &last{}
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

type last struct {
	IP      string
	Created int64
}

func (l *last) ParseDefault(v string) error {
	vals := strings.Split(v, ",")
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

// Meta 指定表属性
func (m *User) Meta() string {
	return "mysql_engine(innodb);mysql_charset(utf8);name(users)"
}

// Admin 带自增和两个唯一约束
type Admin struct {
	Admin1
	Email string `orm:"name(email);len(20);unique(unique_admin_email)"`
	Group int64  `orm:"name(group);fk(fk_admin_name,#groups,id,NO ACTION)"`
}

// Meta 指定表属性
func (m *Admin) Meta() string {
	return "check(admin_chk_name,{group}>0);mysql_engine(innodb);mysql_charset(utf8);name(administrators)"
}

type obj struct {
	ID int `orm:"name(id);ai"`
}

// Meta 指定表属性
func (m obj) Meta() string {
	return `name(objs)`
}

type viewObject struct {
	ID       int    `orm:"name(id);ai"`
	Username string `orm:"name(username)"`
}

func (v *viewObject) Meta() string {
	return "name(view_objects)"
}

func (v *viewObject) ViewAs(e core.Engine) (string, error) {
	return sqlbuilder.Select(e).
		Columns("id", "username").
		From("User").
		Where("id>?", 10).
		CombineSQL()
}

func TestModels_New(t *testing.T) {
	a := assert.New(t)
	ms := NewModels(nil)
	a.NotNil(ms)

	m, err := ms.New(&Admin{})
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
	index, found := m.Indexes["index_admin_name"]
	a.True(found).Equal(usernameCol, index[0])

	// ai
	a.Equal(m.AutoIncrement, idCol).
		Empty(m.PrimaryKey) // 有自增，则主键为空

	// unique_name
	unique, found := m.Uniques["unique_admin_username"]
	a.True(found).Equal(unique[0], usernameCol)

	fk := m.ForeignKeys["fk_admin_name"]
	a.True(found).
		Equal(fk.Column, groupCol).
		Equal(fk.RefTableName, "#groups").
		Equal(fk.RefColName, "id").
		Equal(fk.UpdateRule, "NO ACTION").
		Equal(fk.DeleteRule, "")

	// check
	chk, found := m.Checks["admin_chk_name"]
	a.True(found).Equal(chk, "{group}>0")

	// meta
	a.Equal(m.Meta, map[string][]string{
		"mysql_engine":  {"innodb"},
		"mysql_charset": {"utf8"},
	})

	// Meta返回的name属性
	a.Equal(m.Name, "#administrators")

	// 多层指针下的 Receive
	o := obj{}
	m, err = ms.New(o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "#objs")
	m, err = ms.New(&o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "#objs")
	oo := &o
	m, err = ms.New(&oo)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "#objs")

	// 无效的 New
	m, err = ms.New(123)
	a.ErrorType(err, fetch.ErrInvalidKind).Nil(m)

	// view
	m, err = ms.New(&viewObject{})
	a.NotError(err).NotNil(m)
	a.Equal(m.Type, core.View).
		NotNil(m.ViewAs)
}

func TestModel_parseMeta(t *testing.T) {
	a := assert.New(t)
	m := &core.Model{
		Checks: map[string]string{},
	}

	// 空值不算错误
	a.NotError(parseMeta(m, ""))

	// name 属性过多
	a.Error(parseMeta(m, "name(m1,m2)"))

	// check 属性过多或是过少
	a.Error(parseMeta(m, "check(ck,id>0 AND id<10,error)"))

	// check 添加成功
	a.NotError(parseMeta(m, "check(ck,id>0 AND id<10)"))

	// check 与已有 check 名称相同
	a.Error(parseMeta(m, "check(ck,id>0)"))
}

func TestModel_setOCC(t *testing.T) {
	a := assert.New(t)
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
	a := assert.New(t)
	m := core.NewModel(core.Table, "m1", 10)
	a.NotNil(m)
	c, err := core.NewColumn(core.Int8)
	a.NotError(err).NotNil(c)

	// 过多的参数
	col := &column{Column: c}
	a.Error(setPK(m, col, []string{"123"}))
}

func TestModel_setAI(t *testing.T) {
	a := assert.New(t)
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
