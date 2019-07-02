// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

var (
	_ execPreparer = &DeleteStmt{}
	_ execPreparer = &UpdateStmt{}
	_ execPreparer = &InsertStmt{}
	_ execPreparer = &UpdateStmt{}
	_ queryer      = &SelectStmt{}
)

type execer interface {
	Exec() (sql.Result, error)
	ExecContext(ctx context.Context) (sql.Result, error)
}

type queryer interface {
	Query() (*sql.Rows, error)
	QueryContext(ctx context.Context) (*sql.Rows, error)
}

type preparer interface {
	Prepare() (*sql.Stmt, error)
	PrepareContext(ctx context.Context) (*sql.Stmt, error)
}

type execPreparer interface {
	execer
	preparer
}

type queryPreparer interface {
	queryer
	preparer
}
