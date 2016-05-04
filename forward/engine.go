// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import (
	"database/sql"
)

// DB与Tx的共有接口，方便以下方法调用。
type Engine interface {
	Dialect() Dialect

	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)

	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)

	Prepare(replace bool, query string) (*sql.Stmt, error)

	Prefix() string
}
