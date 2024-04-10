// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v6/core"

// Version 查询数据库服务器的版本信息
func Version(e core.Engine) (version string, err error) {
	err = e.QueryRow(e.Dialect().VersionSQL()).Scan(&version)
	return
}
