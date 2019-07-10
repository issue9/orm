// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"database/sql"

	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.Engine = &sql.DB{}
	_ sqlbuilder.Engine = &sql.Tx{}
)
