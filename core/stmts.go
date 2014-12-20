// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
	"fmt"
	"sync"
)

// sql.Stmt的缓存集合。
type Stmts struct {
	sync.Mutex
	items map[string]*sql.Stmt
	db    DB
}

// 声明一个Stmts实例。
func NewStmts(d DB) *Stmts {
	return &Stmts{
		items: map[string]*sql.Stmt{},
		db:    d,
	}
}

// 编译SQL语句成sql.Stmt，并以name为名称缓存。
// 若该name的缓存已经存在，则返回一个错误信息。
func (s *Stmts) AddSQL(name, sql string) (*sql.Stmt, error) {
	s.Lock()
	defer s.Unlock()

	if _, found := s.items[name]; found {
		return nil, fmt.Errorf("该名称[%v]的stmt已经存在", name)
	}

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	s.items[name] = stmt
	return stmt, nil
}

// 修改或是添加一条SQL语句的缓存。
// 功能上大致与AddSQL()相同，只是在相同名称已经的sql.Stmt实例
// 已经存在的情况下，AddSQL()返回错误，而SetSQL()则是替换。
func (s *Stmts) SetSQL(name, sql string) (*sql.Stmt, error) {
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	s.Lock()
	defer s.Unlock()

	s.items[name] = stmt
	return stmt, nil
}

// 添加一个sql.Stmt，若已经存在相同名称的，则返回false。
// 若要修改已存在的sql.Stmt，请使用Set()函数
func (s *Stmts) Add(name string, stmt *sql.Stmt) bool {
	s.Lock()
	defer s.Unlock()

	if _, found := s.items[name]; found {
		return false
	}

	s.items[name] = stmt
	return true
}

// 添加或是修改sql.Stmt
func (s *Stmts) Set(name string, stmt *sql.Stmt) {
	s.Lock()
	defer s.Unlock()

	s.items[name] = stmt
}

// 查找指定名称的sql.Stmt实例。若不存在，返回nil,false
func (s *Stmts) Get(name string) (stmt *sql.Stmt, found bool) {
	stmt, found = s.items[name]
	return
}

// 清除所有的缓存。
func (s *Stmts) Clear() {
	s.free()
	s.items = map[string]*sql.Stmt{}
}

// 释放所有缓存空间。
// 与Clear()的区别在于：Close()之后，不能再次通过AddSQL()
// 等函数添加新的缓存内容。
func (s *Stmts) Close() {
	s.free()
	s.items = nil
}

// 释放所有的sql.Stmt实例。
func (s *Stmts) free() {
	s.Lock()
	defer s.Unlock()

	for _, stmt := range s.items {
		stmt.Close()
	}
}
