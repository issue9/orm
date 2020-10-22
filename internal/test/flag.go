// SPDX-License-Identifier: MIT

package test

import (
	"flag"
	"fmt"
	"strings"
)

var (
	dbString = flag.String("dbs", "", "指定需要测试的数据库，格式为 dbName,driverName:dbName,driverName")
	dbs      []*dbFlag
)

type dbFlag struct {
	dbName, driverNme string
}

func parseFlag() {
	if dbs != nil {
		return
	}

	dbs = make([]*dbFlag, 0, 10)

	if len(*dbString) == 0 {
		return
	}

	items := strings.Split(*dbString, ":")
	for _, item := range items {
		i := strings.Split(item, ",")
		if len(i) != 2 {
			panic(fmt.Sprintf("格式错误：%v", *dbString))
		}
		dbs = append(dbs, &dbFlag{dbName: i[0], driverNme: i[1]})
	}
}
