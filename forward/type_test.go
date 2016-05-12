// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import "database/sql"

// forward.Engine

type engine struct {
	dialect Dialect
}

func (e *engine) Dialect() Dialect { return e.dialect }

func (e *engine) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (e *engine) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (e *engine) Prepare(replace bool, query string) (*sql.Stmt, error) { return nil, nil }

func (e *engine) Prefix() string { return "engine" }

// forward.Dialect

type dialect struct {
}

func (d *dialect) Name() string { return "dialect_test" }

func (d *dialect) QuoteTuple() (openQuote, closeQuote byte) { return '`', '`' }

func (d *dialect) Quote(w *SQL, colName string) {}

func (d *dialect) ReplaceMarks(*string) error { return nil }

func (d *dialect) LimitSQL(sql *SQL, limit int, offset ...int) []interface{} {
	if len(offset) == 0 {
		sql.WriteString(" LIMIT ? ")
		return []interface{}{limit}
	}

	sql.WriteString("LIMIT ? OFFSET ? ")
	return []interface{}{limit, offset[0]}
}

func (d *dialect) NoAIColSQL(w *SQL, m *Model) error { return nil }

func (d *dialect) AIColSQL(w *SQL, m *Model) error { return nil }

func (d *dialect) ConstraintsSQL(w *SQL, m *Model) {}

func (d *dialect) TruncateTableSQL(w *SQL, tableName, aiColumn string) {}

func (d *dialect) SupportInsertMany() bool { return true }
