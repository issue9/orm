// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/dialect"
	"github.com/issue9/orm/v6/internal/model"
	"github.com/issue9/orm/v6/internal/model/testdata"
	"github.com/issue9/orm/v6/internal/test"
)

func TestMain(m *testing.M) { test.Main(m) }

func newModules(a *assert.Assertion) *model.Models {
	const testDB = "./test.db"

	db, err := sql.Open("sqlite3", testDB)
	a.NotError(err).NotNil(db)

	ms, e, err := model.NewModels(db, dialect.Sqlite3("sqlite"), "")
	a.NotError(err).
		NotNil(ms).
		NotNil(e).
		NotEmpty(ms.Version()).
		NotNil(ms.DB())

	a.TB().Cleanup(func() {
		a.NotError(os.Remove(testDB))
	})

	return ms
}

func TestModels(t *testing.T) {
	a := assert.New(t, false)
	ms := newModules(a)

	m, err := ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, ms.Length())

	// 相同的 model 实例，不会增加数量
	m, err = ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, ms.Length())

	// 相同的表名，但是类型不同
	m, err = ms.New(&testdata.User{})
	a.NotError(err).
		NotNil(m).
		Equal(2, ms.Length())

	// 添加新的 model
	m, err = ms.New(&Admin{})
	a.NotError(err).
		NotNil(m).
		Equal(3, ms.Length())

	a.NotError(ms.Close())
	a.Equal(0, ms.Length())
}
