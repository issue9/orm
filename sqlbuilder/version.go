// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v3/core"

// Version 查询数据库服务器的版本信息
func Version(e core.Engine) (version string, err error) {
	if err := e.QueryRow(e.Dialect().VersionSQL()).Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}
