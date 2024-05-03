// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package dialect

import (
	"os/exec"

	"github.com/issue9/sliceutil"
)

type base struct {
	driverName     string
	name           string
	quoteL, quoteR byte
}

func newBase(name, driverName string, quoteLeft, quoteRight byte) base {
	return base{
		name:       name,
		driverName: driverName,
		quoteL:     quoteLeft,
		quoteR:     quoteRight,
	}
}

func (b *base) Name() string { return b.name }

func (b *base) DriverName() string { return b.driverName }

func (b *base) Quotes() (byte, byte) { return b.quoteL, b.quoteR }

func buildCmdArgs(k, v string) string {
	if v == "" {
		return ""
	}
	return k + "=" + v
}

func newCommand(name string, env, kv []string) *exec.Cmd {
	env = sliceutil.Filter(env, func(i string, _ int) bool { return i != "" })
	kv = sliceutil.Filter(kv, func(i string, _ int) bool { return i != "" })
	cmd := exec.Command(name, kv...)
	cmd.Env = append(cmd.Env, env...)
	return cmd
}
