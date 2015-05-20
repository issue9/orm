// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
)

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

func (s *sqlite3) LimitSQL(w *bytes.Buffer, limit interface{}, offset ...interface{}) error {
	w.WriteString(" LIMIT ")
	WriteString(w, limit)

	if len(offset) == 0 {
		return nil
	}

	w.WriteString(" OFFSET ")
	return WriteString(w, offset[0])
}

func (s *sqlite3) CreateTableSQL(m *Model) (string, error) {
	return createdSQL[m.Name], nil
}

func (s *sqlite3) TruncateTableSQL(tableName string) string {
	return "DELETE FROM " + tableName
}
