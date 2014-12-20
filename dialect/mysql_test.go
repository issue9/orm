// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
)

var _ base = &Mysql{}

var m = &Mysql{}

func TestMysqlGetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(m.GetDBName("root:password@/dbname"), "dbname")
	a.Equal(m.GetDBName("root:@/dbname"), "dbname")
	a.Equal(m.GetDBName("root:password@tcp(localhost:3066)/dbname"), "dbname")
	a.Equal(m.GetDBName("root:password@unix(/tmp/mysql.lock)/dbname?loc=Local"), "dbname")
	a.Equal(m.GetDBName("root:/"), "")
}

func TestMysqlSQLType(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString("")
	col := &core.Column{
		GoType: reflect.TypeOf(1),
	}

	m.sqlType(buf, col)

}
