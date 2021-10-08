// SPDX-License-Identifier: MIT

// Package flagtest 用于生成测试用的命令行参数
package flagtest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Flags 解析后的参数值
var Flags []*Flag

// Flag 参数对象
type Flag struct {
	DBName, DriverName string
}

// Main 供测试的 TestMain 调用
func Main(m *testing.M) {
	dbString := flag.String("dbs", "sqlite3,sqlite3", "指定需要测试的数据库，格式为 dbName,driverName:dbName,driverName")

	flag.Parse()

	if *dbString == "" || Flags != nil {
		return
	}

	Flags = make([]*Flag, 0, 10)

	items := strings.Split(*dbString, ":")
	for _, item := range items {
		i := strings.Split(item, ",")
		if len(i) != 2 {
			panic(fmt.Sprintf("格式错误：%v", *dbString))
		}
		Flags = append(Flags, &Flag{DBName: i[0], DriverName: i[1]})
	}

	os.Exit(m.Run())
}
