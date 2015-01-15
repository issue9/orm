// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/issue9/orm/core"
)

// SQL.And()的别名
func (s *SQL) Where(cond string) *SQL {
	return s.And(cond)
}

// WHERE ... AND ...
func (s *SQL) And(cond string) *SQL {
	return s.whereBuild(0, cond)
}

// WHERE ... OR ...
func (s *SQL) Or(cond string) *SQL {
	return s.whereBuild(1, cond)
}

// WHERE ... AND col BETWEEN ...
func (s *SQL) AndBetween(col string, start, end interface{}) *SQL {
	return s.between(0, col, start, end)
}

// WHERE ... OR col BETWEEN ...
func (s *SQL) OrBetween(col string, start, end interface{}) *SQL {
	return s.between(1, col, start, end)
}

// SQL.AndBetween()的别名
func (s *SQL) Between(col string, start, end interface{}) *SQL {
	return s.AndBetween(col, start, end)
}

// WHERE ... AND col IN (...)
func (s *SQL) AndIn(col string, args ...interface{}) *SQL {
	return s.in(0, col, args...)
}

// WHERE ... OR col IN (...)
func (s *SQL) OrIn(col string, args ...interface{}) *SQL {
	return s.in(1, col, args...)
}

// SQL.AndIn()的别名
func (s *SQL) In(col string, args ...interface{}) *SQL {
	return s.AndIn(col, args...)
}

// WHERE ... AND col IS NULL
func (s *SQL) AndIsNull(col string) *SQL {
	return s.isNull(0, col)
}

// WHERE ... OR col IN NULL
func (s *SQL) OrIsNull(col string) *SQL {
	return s.isNull(1, col)
}

// SQL.AndIsNull()的别名
func (s *SQL) IsNull(col string) *SQL {
	return s.AndIsNull(col)
}

// WHERE ... AND col IS NOT NULL
func (s *SQL) AndIsNotNull(col string) *SQL {
	return s.isNotNull(0, col)
}

// WHERE ... OR col IS NOT NULL
func (s *SQL) OrIsNotNull(col string) *SQL {
	return s.isNotNull(1, col)
}

// SQL.AndIsNotNull()的别名
func (s *SQL) IsNotNull(col string) *SQL {
	return s.AndIsNotNull(col)
}

// 所有where子句的构建，最终都调用此方法来写入实例中。
// op 与前一个语句的连接符号，可以是0(and)或是1(or)常量；
// cond where语句部分，不能直接使用?符号，但可以用`@xx`的形式代替；
//  w := NewSQL(...)
//  w.whereBuild(0, "username=='abc'")
//  w.whereBuild(1, "username=@username") // 正确，参数在运行时才给出
func (s *SQL) whereBuild(op int, cond string) *SQL {
	switch {
	case s.cond.Len() == 0:
		s.cond.WriteString(" WHERE(")
	case op == 0:
		s.cond.WriteString(" AND(")
	case op == 1:
		s.cond.WriteString(" OR(")
	default:
		s.errors = append(s.errors, fmt.Errorf("whereBuild:无效的op操作符:[%v]", op))
	}

	s.cond.WriteString(cond)
	s.cond.WriteByte(')')

	return s
}

// WHERE col IN(v1,v2)语句的实现函数，供andIn()和orIn()函数调用。
func (s *SQL) in(op int, col string, args ...interface{}) *SQL {
	if len(args) <= 0 {
		s.errors = append(s.errors, errors.New("in:args参数不能为空"))
		return s
	}

	cond := bytes.NewBufferString(col)
	cond.WriteString(" IN(")
	for _, arg := range args {
		cond.WriteString(core.AsSQLValue(arg))
		cond.WriteByte(',')
	}
	cond.Truncate(cond.Len() - 1) // 去掉最后的逗号
	cond.WriteByte(')')

	return s.whereBuild(op, cond.String())
}

// 供andBetween()和orBetween()调用。
func (s *SQL) between(op int, col string, start, end interface{}) *SQL {
	return s.whereBuild(op, col+" BETWEEN "+core.AsSQLValue(start)+" AND "+core.AsSQLValue(end))
}

// 供andIsNull()和orIsNull()调用。
func (w *SQL) isNull(op int, col string) *SQL {
	return w.whereBuild(op, col+" IS NULL")
}

// 供andIsNotNull()和orIsNotNull()调用。
func (w *SQL) isNotNull(op int, col string) *SQL {
	return w.whereBuild(op, col+" IS NOT NULL")
}
