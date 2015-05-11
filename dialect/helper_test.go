// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm"
)

var replacer = strings.NewReplacer(")", " ) ", "(", " ( ", ",", " , ")

var spaceReplaceRegexp = regexp.MustCompile("\\s+")

// 检测两条SQL语句是否相等，忽略大小写与多余的空格。
func chkSQLEqual(a *assert.Assertion, s1, s2 string) {
	// 将'(', ')', ',' 等字符的前后空格标准化
	s1 = replacer.Replace(s1)
	s2 = replacer.Replace(s2)

	// 转换成小写，去掉首尾空格
	s1 = strings.TrimSpace(strings.ToLower(s1))
	s2 = strings.TrimSpace(strings.ToLower(s2))

	// 去掉多余的空格。
	s1 = spaceReplaceRegexp.ReplaceAllString(s1, " ")
	s2 = spaceReplaceRegexp.ReplaceAllString(s2, " ")

	a.Equal(s1, s2)
}

func TestCreatColSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col := &orm.Column{}

	col.Name = "id"
	col.GoType = reflect.TypeOf(1)
	col.Len1 = 5
	createColSQL(dialect, buf, col)
	wont := "id BIGINT(5) NOT NULL"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	col.Len1 = 0
	col.GoType = reflect.TypeOf(int8(1))
	col.HasDefault = true
	col.Default = "1"
	createColSQL(dialect, buf, col)
	wont = "id SMALLINT NOT NULL DEFAULT '1'"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	col.HasDefault = false
	col.Nullable = true
	createColSQL(dialect, buf, col)
	wont = "id SMALLINT NULL"
}

func TestCreatePKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col1 := &orm.Column{Name: "id"}
	col2 := &orm.Column{Name: "username"}
	cols := []*orm.Column{col1, col2}

	createPKSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname PRIMARY KEY(id,username)"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	createPKSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname PRIMARY KEY(id)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateUniqueSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col1 := &orm.Column{Name: "id"}
	col2 := &orm.Column{Name: "username"}
	cols := []*orm.Column{col1, col2}

	createUniqueSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname UNIQUE(id,username)"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	createUniqueSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname UNIQUE(id)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateFKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	fk := &orm.ForeignKey{
		Col:          &orm.Column{Name: "id"},
		RefTableName: "refTable",
		RefColName:   "refCol",
		UpdateRule:   "NO ACTION",
	}

	createFKSQL(dialect, buf, fk, "fkname")
	wont := "CONSTRAINT fkname FOREIGN KEY(id) REFERENCES refTable(refCol) ON UPDATE NO ACTION"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateCheckSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")

	createCheckSQL(dialect, buf, "id>5", "chkname")
	wont := "CONSTRAINT chkname CHECK(id>5)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)
	w := new(bytes.Buffer)

	a.NotError(mysqlLimitSQL(w, 5, 0))
	chkSQLEqual(a, w.String(), " LIMIT 5 OFFSET 0 ")

	w.Reset()
	a.NotError(mysqlLimitSQL(w, 5))
	chkSQLEqual(a, w.String(), "LIMIT 5")
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)
	w := new(bytes.Buffer)

	a.NotError(oracleLimitSQL(w, 5, 0))
	chkSQLEqual(a, w.String(), " OFFSET 0 ROWS FETCH NEXT 5 ROWS ONLY ")

	w.Reset()
	a.NotError(oracleLimitSQL(w, 5))
	chkSQLEqual(a, w.String(), "FETCH NEXT 5 ROWS ONLY ")
}
