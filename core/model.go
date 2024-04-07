// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/issue9/sliceutil"
)

const pkName = "_pk"

var (
	// ErrAutoIncrementPrimaryKeyConflict 自增和主键不能同时存在
	//
	// 当添加自增时，会自动将其转换为主键，如果此时已经已经存在主键，则会报此错误。
	ErrAutoIncrementPrimaryKeyConflict = errors.New("自增和主键不能同时存在")
)

func errColMustNumber(name string) error { return fmt.Errorf("列 %s 必须是数值类型", name) }

type (
	// Viewer 视图必须要实现的接口
	//
	// 当一个模型实现了该接口，会被识别为视图模型，不再在数据库中创建普通的数据表。
	Viewer interface {
		// ViewAs 返回视图所需的 Select 语句
		ViewAs(e Engine) (string, error)
	}

	// TableNamer 表或是视图必须实现的接口
	TableNamer interface {
		// TableName 返回表或是视图的名称
		TableName() string
	}

	// ApplyModeler 加载数据模型
	//
	// 当一个对象实现此接口时，那么在将对象转换成 [Model] 类型时，
	// 会调用 ApplyModel 方法，给予用户修改 [Model] 的机会。
	ApplyModeler interface {
		ApplyModel(*Model) error
	}

	// ForeignKey 外键
	ForeignKey struct {
		Name                     string // 约束名
		Column                   *Column
		RefTableName, RefColName string
		UpdateRule, DeleteRule   string
	}

	Constraint struct {
		Name    string
		Columns []*Column
	}

	// ModelType 表示数据模型的类别
	ModelType int8

	// Model 表示一个数据库的表或视图模型
	Model struct {
		GoType reflect.Type

		// 模型的名称
		Name string

		// 如果当前模型是视图，那么此值表示的是视图的 select 语句，
		// 其它类型下，ViewAs 不启作用。
		ViewAs string

		Type    ModelType
		Columns []*Column

		// 约束与索引
		//
		// NOTE: 如果是视图模式，理论上是不存在约束信息的，
		// 但是依然可以指定约束，这些信息主要是给 ORM 查看，以便构建搜索语句。
		Checks        map[string]string
		ForeignKeys   []*ForeignKey
		Indexes       []*Constraint // 目前不支持唯一索引，如果需要唯一索引，可以设置成唯一约束。
		Uniques       []*Constraint
		AutoIncrement *Column
		PrimaryKey    *Constraint

		OCC *Column // 乐观锁

		// 表级别的数据
		//
		// 如存储引擎，表名和字符集等，在创建表时，可能会用到这此数据。
		// 可以采用 [Dialect.Name] 限定数据库，比如 mysql_charset 限定为 mysql 下的 charset 属性。
		// 具体可参考各个 dialect 实现的介绍。
		Options map[string][]string
	}
)

// 目前支持的数据模型类别
//
// Table 表示为一张普通的数据表，默认的模型即为 [Table]；
// 如果实现了 [Viewer] 接口，则该模型改变视图类型，即 [View]。
//
// 两者的创建方式稍微有点不同：
// Table 类型创建时，会采用列、约束和索引等信息创建表；
// 而 View 创建时，只使用了 Viewer 接口返回的 Select
// 语句作为内容生成语句，像约束等信息，仅作为查询时的依据，
// 当然 select 语句中的列需要和 [Columns] 中的列要相对应，
// 否则可能出错。
//
// 在视图类型中，唯一约束、主键约束、自增约束依然是可以定义的，
// 虽然不会呈现在视图中，但是在查询时，可作为 orm 的一个判断依据。
const (
	none ModelType = iota
	Table
	View
)

// NewModel 初始化 [Model]
//
// cap 表示列的数量，如果指定了，可以提前分配 [Model.Columns] 字段的大小。
func NewModel(modelType ModelType, name string, cap int) *Model {
	return &Model{
		Name:    name,
		Type:    modelType,
		Columns: make([]*Column, 0, cap),
		Options: map[string][]string{},
		Checks:  map[string]string{},
	}
}

// Reset 清空模型内容
func (m *Model) Reset() {
	m.GoType = nil
	m.Name = ""
	m.ViewAs = ""
	m.Type = none
	m.Columns = m.Columns[:0]
	m.OCC = nil
	m.Options = map[string][]string{}
	m.Checks = map[string]string{}
	m.ForeignKeys = m.ForeignKeys[:0]
	m.AutoIncrement = nil
	m.PrimaryKey = nil
	m.Indexes = m.Indexes[:0]
	m.Uniques = m.Uniques[:0]
}

// SetAutoIncrement 将 col 列设置为自增列
//
// 如果已经存在自增列或是主键，返回错误。
func (m *Model) SetAutoIncrement(col *Column) error {
	switch col.PrimitiveType {
	case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64:
	default:
		return errColMustNumber(col.Name)
	}

	if m.AutoIncrement != nil {
		return ErrConstraintExists(m.AutoIncrement.Name)
	}

	if m.PrimaryKey != nil {
		return ErrAutoIncrementPrimaryKeyConflict
	}

	if !m.columnExists(col) {
		return errConstraintColumnNotExists("AutoIncrement", col.Name)
	}

	col.AI = true
	m.AutoIncrement = col
	return nil
}

// AddPrimaryKey 指定主键约束的列
//
// 自增会自动转换为主键。
// 多次调用，则多列形成一个多列主键。
func (m *Model) AddPrimaryKey(col *Column) error {
	if m.AutoIncrement != nil {
		return ErrAutoIncrementPrimaryKeyConflict
	}

	if !m.columnExists(col) {
		return errConstraintColumnNotExists("PrimaryKey", col.Name)
	}

	if m.PrimaryKey == nil {
		m.PrimaryKey = &Constraint{Name: pkName, Columns: make([]*Column, 0, 5)}
	}
	m.PrimaryKey.append(col)

	return nil
}

// SetOCC 设置该列为乐观锁
func (m *Model) SetOCC(col *Column) error {
	if col.AI || col.Nullable {
		return fmt.Errorf("乐观锁列 %s 不能为同时为自增或 NULL", col.Name)
	}

	if m.OCC != nil {
		return fmt.Errorf("已经存在乐观锁 %s", m.OCC.Name)
	}

	switch col.PrimitiveType {
	case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64:
	default:
		return errColMustNumber(col.Name)
	}

	if !m.columnExists(col) {
		return fmt.Errorf("列 %s 未找到", col.Name)
	}
	m.OCC = col
	return nil
}

// AddIndex 添加索引列
//
// 如果 name 不存在，则创建新的索引
//
// NOTE: 如果 typ == IndexUnique，则等同于调用 AddUnique。
func (m *Model) AddIndex(typ IndexType, name string, col *Column) error {
	if typ == IndexUnique { // 唯一索引直接转为唯一约束
		return m.AddUnique(name, col)
	}

	if !m.columnExists(col) {
		return errConstraintColumnNotExists("Index", col.Name)
	}

	if m.Indexes == nil {
		m.Indexes = make([]*Constraint, 0, 5)
	}
	if index, found := m.Index(name); found {
		index.append(col)
	} else {
		m.Indexes = append(m.Indexes, &Constraint{Name: name, Columns: []*Column{col}})
	}
	return nil
}

// AddUnique 添加唯一约束的列到 name
//
// 如果 name 不存在，则创建新的约束
func (m *Model) AddUnique(name string, col *Column) error {
	if !m.columnExists(col) {
		return errConstraintColumnNotExists("Unique", col.Name)
	}

	if m.Uniques == nil {
		m.Uniques = make([]*Constraint, 0, 5)
	}
	if unique, found := m.Unique(name); found {
		unique.append(col)
	} else {
		m.Uniques = append(m.Uniques, &Constraint{Name: name, Columns: []*Column{col}})
	}

	return nil
}

// NewCheck 添加新的 check 约束
func (m *Model) NewCheck(name string, expr string) error {
	if _, found := m.Checks[name]; found {
		return ErrConstraintExists(name)
	}

	m.Checks[name] = expr
	return nil
}

// NewForeignKey 添加新的外键
func (m *Model) NewForeignKey(fk *ForeignKey) error {
	if fk.Column == nil || fk.RefColName == "" || fk.RefTableName == "" {
		return fmt.Errorf("约束 %s 的 Column、RefColName 和 RefTableName 都不能为空", fk.Name)
	}

	if _, found := m.ForeignKey(fk.Name); found {
		return ErrConstraintExists(fk.Name)
	}

	if !m.columnExists(fk.Column) {
		return errConstraintColumnNotExists("ForeignKey", fk.Column.Name)
	}
	m.ForeignKeys = append(m.ForeignKeys, fk)
	return nil
}

// Sanitize 对整个对象做一次修正和检测，查看是否合法
//
// NOTE: 必需在 Model 初始化完成之后调用。
func (m *Model) Sanitize() error {
	if m.Name == "" {
		return errors.New("缺少模型名称")
	}

	if m.Type != Table && m.Type != View {
		return errors.New("无效的类型")
	}

	if m.PrimaryKey != nil && len(m.PrimaryKey.Columns) == 1 {
		pk := m.PrimaryKey.Columns[0]
		if pk.HasDefault || pk.Nullable {
			return fmt.Errorf("单一主键约束的列 %s 不能为同时设置为默认值", pk.Name)
		}
	}

	for _, c := range m.Columns {
		if err := c.Check(); err != nil {
			return err
		}
	}

	if err := m.PrimaryKey.sanitize(); err != nil {
		return err
	}

	for _, i := range m.Indexes {
		if i == nil {
			return errors.New("存在空的索引")
		}
		if err := i.sanitize(); err != nil {
			return err
		}
	}

	for _, i := range m.Uniques {
		if i == nil {
			return errors.New("存在空的唯一约束")
		}
		if err := i.sanitize(); err != nil {
			return err
		}
	}

	for _, fk := range m.ForeignKeys {
		if fk == nil {
			return errors.New("存在空的外键约束")
		}
		if err := fk.sanitize(); err != nil {
			return err
		}
	}

	return m.checkNames()
}

func errConstraintColumnNotExists(constraint, col string) error {
	return fmt.Errorf("约束 %s 的列 %s 不存在", constraint, col)
}

func (m *Model) checkNames() error {
	l := 2 + len(m.Indexes) + len(m.Uniques) + len(m.ForeignKeys) + len(m.Checks)
	names := make([]string, 0, l)

	if m.PrimaryKey != nil { // 由 Constraint.sanitize 保证 name 不为空值
		names = append(names, m.PrimaryKey.Name)
	}

	for _, c := range m.Indexes {
		names = append(names, c.Name)
	}

	for _, c := range m.Uniques {
		names = append(names, c.Name)
	}

	for _, fk := range m.ForeignKeys {
		names = append(names, fk.Name)
	}

	for name := range m.Checks {
		names = append(names, name)
	}

	slices.Sort(names)
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			return ErrConstraintExists(names[i])
		}
	}

	return nil
}

func (m *Model) Index(name string) (*Constraint, bool) {
	return sliceutil.At(m.Indexes, func(e *Constraint, _ int) bool { return e.Name == name })
}

func (m *Model) Unique(name string) (*Constraint, bool) {
	return sliceutil.At(m.Uniques, func(e *Constraint, _ int) bool { return e.Name == name })
}

func (m *Model) ForeignKey(name string) (*ForeignKey, bool) {
	return sliceutil.At(m.ForeignKeys, func(e *ForeignKey, _ int) bool { return e.Name == name })
}

func (c *Constraint) append(col *Column) {
	if c.Columns == nil {
		c.Columns = make([]*Column, 0, 5)
	}
	c.Columns = append(c.Columns, col)
}

func (c *Constraint) sanitize() error {
	if c == nil {
		return nil
	}

	if c.Name == "" {
		return errors.New("未指定约束名")
	}

	if len(c.Columns) == 0 {
		return fmt.Errorf("约束 %s 并未指定列", c.Name)
	}
	return nil
}

func (f *ForeignKey) sanitize() error {
	if f == nil {
		return nil
	}

	if f.Name == "" {
		return errors.New("未指定外键的约束名")
	}

	if f.Column == nil {
		return fmt.Errorf("外键约束 %s 并未指定列", f.Name)
	}

	if f.RefTableName == "" || f.RefColName == "" {
		return fmt.Errorf("外键约束 %s 缺少必要的字段 ref", f.Name)
	}

	return nil
}
