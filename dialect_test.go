// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
)

var _ Dialect = &sqlite3{}

type sqlite3 struct {
}

func (s *sqlite3) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (s *sqlite3) Quote(w *bytes.Buffer, name string) error {
	w.WriteByte('`')
	w.WriteString(name)
	return w.WriteByte('`')
}

func (s *sqlite3) LimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error) {
	w.WriteString(" LIMIT ?")

	if len(offset) == 0 {
		return []int{limit}, nil
	}

	w.WriteString(" OFFSET ?")
	return []int{limit, offset[0]}, nil
}

func (s *sqlite3) CreateTableSQL(m *Model) (string, error) {
	return createdSQL[m.Name], nil
}

func (s *sqlite3) TruncateTableSQL(tableName string) string {
	return "DELETE FROM " + tableName + ";update sqlite_sequence set seq=0 where name='" + tableName + "'"
}
