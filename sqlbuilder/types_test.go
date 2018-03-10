// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

var (
	_ execer  = &DeleteStmt{}
	_ execer  = &UpdateStmt{}
	_ execer  = &InsertStmt{}
	_ execer  = &UpdateStmt{}
	_ execer  = &CreateIndexStmt{}
	_ execer  = &TruncateStmt{}
	_ queryer = &SelectStmt{}
)

type execer interface {
	Exec() (sql.Result, error)
	ExecContext(ctx context.Context) (sql.Result, error)
	Prepare() (*sql.Stmt, error)
	PrepareContext(ctx context.Context) (*sql.Stmt, error)
}

type queryer interface {
	Query() (*sql.Rows, error)
	QueryContext(ctx context.Context) (*sql.Rows, error)
	Prepare() (*sql.Stmt, error)
	PrepareContext(ctx context.Context) (*sql.Stmt, error)
}
