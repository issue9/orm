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
	"github.com/issue9/orm/forward"
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
	dialect := &mysql{}
	buf := bytes.NewBufferString("")
	col := &forward.Column{}

	col.Name = "id"
	col.GoType = reflect.TypeOf(1)
	col.Len1 = 5
	createColSQL(dialect, buf, col)
	wont := "`id` BIGINT(5) NOT NULL"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	col.Len1 = 0
	col.GoType = reflect.TypeOf(int8(1))
	col.HasDefault = true
	col.Default = "1"
	createColSQL(dialect, buf, col)
	wont = "`id` SMALLINT NOT NULL DEFAULT '1'"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	col.HasDefault = false
	col.Nullable = true
	createColSQL(dialect, buf, col)
	wont = "`id` SMALLINT NULL"
}

func TestCreatePKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := bytes.NewBufferString("")
	col1 := &forward.Column{Name: "id"}
	col2 := &forward.Column{Name: "username"}
	cols := []*forward.Column{col1, col2}

	createPKSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname PRIMARY KEY(`id`,`username`)"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	createPKSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname PRIMARY KEY(`id`)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateUniqueSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := bytes.NewBufferString("")
	col1 := &forward.Column{Name: "id"}
	col2 := &forward.Column{Name: "username"}
	cols := []*forward.Column{col1, col2}

	createUniqueSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname UNIQUE(`id`,`username`)"
	chkSQLEqual(a, buf.String(), wont)

	buf.Reset()
	createUniqueSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname UNIQUE(`id`)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateFKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := bytes.NewBufferString("")
	fk := &forward.ForeignKey{
		Col:          &forward.Column{Name: "id"},
		RefTableName: "refTable",
		RefColName:   "refCol",
		UpdateRule:   "NO ACTION",
	}

	createFKSQL(dialect, buf, fk, "fkname")
	wont := "CONSTRAINT fkname FOREIGN KEY(`id`) REFERENCES refTable(`refCol`) ON UPDATE NO ACTION"
	chkSQLEqual(a, buf.String(), wont)
}

func TestCreateCheckSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := bytes.NewBufferString("")

	createCheckSQL(dialect, buf, "id>5", "chkname")
	wont := "CONSTRAINT chkname CHECK(id>5)"
	chkSQLEqual(a, buf.String(), wont)
}

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)
	w := new(bytes.Buffer)

	ret, err := mysqlLimitSQL(w, 5, 0)
	a.NotError(err).Equal(ret, []int{5, 0})
	chkSQLEqual(a, w.String(), " LIMIT ? OFFSET ? ")

	w.Reset()
	ret, err = mysqlLimitSQL(w, 5)
	a.NotError(err).Equal(ret, []int{5})
	chkSQLEqual(a, w.String(), "LIMIT ?")
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)
	w := new(bytes.Buffer)

	ret, err := oracleLimitSQL(w, 5, 0)
	a.NotError(err).Equal(ret, []int{0, 5})
	chkSQLEqual(a, w.String(), " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ")

	w.Reset()
	ret, err = oracleLimitSQL(w, 5)
	a.NotError(err).Equal(ret, []int{5})
	chkSQLEqual(a, w.String(), "FETCH NEXT ? ROWS ONLY ")
}
