// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package dialect

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestBuildCmdArg(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(buildCmdArgs("-p", ""), "").
		Equal(buildCmdArgs("-p", "123"), "-p=123")
}
