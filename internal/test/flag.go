// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

var flags []*flagVar

type flagVar struct {
	DBName, DriverName string
}

// Main 供测试的 TestMain 调用
func Main(m *testing.M) {
	dbString := flag.String("dbs", "sqlite3,sqlite3", "指定需要测试的数据库，格式为 dbName,driverName:dbName,driverName")

	flag.Parse()

	if *dbString == "" || flags != nil {
		return
	}

	flags = make([]*flagVar, 0, 10)

	items := strings.Split(*dbString, ":")
	for _, item := range items {
		i := strings.Split(item, ",")
		if len(i) != 2 {
			panic(fmt.Sprintf("格式错误：%v", *dbString))
		}
		flags = append(flags, &flagVar{DBName: i[0], DriverName: i[1]})
	}

	os.Exit(m.Run())
}
