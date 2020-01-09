// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

var (
	// ErrColumnTypeError 列的类型错误
	//
	// 部分列对其类型有要求，比如自增列和被定义为乐观锁的锁，
	// 其类型必须为数值类型，否则将返回此错误
	ErrColumnTypeError = errors.New("类型必须为数值")

	// ErrAutoIncrementPrimaryKeyConflict 自增和主键不能同时存在
	//
	// 当添加自增时，会自动将其转换为主键，如果此时已经已经存在主键，则会报此错误。
	ErrAutoIncrementPrimaryKeyConflict = errors.New("自增和主键不能同时存在")
)

// ForeignKey 外键
type ForeignKey struct {
	Column                   *Column
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

// ModelType 表示数据模型的类别
type ModelType int8

// 目前支持的数据模型类别
//
// Table 表示为一张普通的数据表，默认的模型即为 Table；
// 如果实现了 Viewer 接口，则该模型改变视图类型，即 View。
//
// 两者的创建方式稍微有点不同：
// Table 类型创建时，会采用列、约束和索引等信息创建表；
// 而 View 创建时，只使用了 Viewer 接口返回的 Select
// 语句作为内容生成语句，像约束等信息，仅作为查询时的依据，
// 当然 select 语句中的列需要和 Columns 中的列要相对应，
// 否则可能出错。
//
// 在视图类型中，唯一约束、主键约束、自增约束依然是可以定义的，
// 虽然不会呈现在视图中，但是在查询时，可作为 orm 的一个判断依据。
const (
	none ModelType = iota
	Table
	View
)

// Model 表示一个数据库的表或视图模型
type Model struct {
	GoType reflect.Type

	// 模型的名称，可以以 # 符号开头，表示该表名带有一个表名前缀。
	// 在生成 SQL 语句时，该符号会被转换成 Engine.TablePrefix()
	// 返回的值。
	Name string

	// 如果当前模型是视图，那么此值表示的是视图的 select 语句，
	// 其它类型下，ViewAs 不启作用。
	ViewAs string

	Type    ModelType
	Columns []*Column
	OCC     *Column             // 乐观锁
	Meta    map[string][]string // 表级别的数据，如存储引擎，表名和字符集等。

	// 索引内容
	//
	// 目前不支持唯一索引，如果需要唯一索引，可以设置成唯一约束。
	Indexes map[string][]*Column

	// 约束
	Uniques       map[string][]*Column
	Checks        map[string]string
	ForeignKeys   map[string]*ForeignKey
	AutoIncrement *Column
	PrimaryKey    []*Column
}

// NewModel 初始化 Model，分其所有变量分配内存。
// 但是变量的内容依然要手动初始化。
//
// cap 表示列的数量，如果指定了，可以提交分配内存。
func NewModel(modelType ModelType, name string, cap int) *Model {
	m := &Model{
		Name:        name,
		Type:        modelType,
		Columns:     make([]*Column, 0, cap),
		Meta:        map[string][]string{},
		Indexes:     map[string][]*Column{},
		Uniques:     map[string][]*Column{},
		Checks:      map[string]string{},
		ForeignKeys: map[string]*ForeignKey{},
		PrimaryKey:  []*Column{},
	}

	return m
}

// Reset 清空模型内容
func (m *Model) Reset() {
	m.GoType = nil
	m.Name = ""
	m.ViewAs = ""
	m.Type = none
	m.Columns = m.Columns[:0]
	m.OCC = nil
	m.Meta = map[string][]string{}
	m.Indexes = map[string][]*Column{}
	m.Uniques = map[string][]*Column{}
	m.Checks = map[string]string{}
	m.ForeignKeys = map[string]*ForeignKey{}
	m.AutoIncrement = nil
	m.PrimaryKey = []*Column{}
}

// AIName 当前模型中自增列的名称
func (m *Model) AIName() string {
	return AIName(m.Name)
}

// PKName 当前模型中主键约束的名称
func (m *Model) PKName() string {
	return PKName(m.Name)
}

// AddColumns 添加新列
func (m *Model) AddColumns(col ...*Column) error {
	for _, c := range col {
		if err := m.AddColumn(c); err != nil {
			return err
		}
	}

	return nil
}

// AddColumn 添加新列
//
// 按添加顺序确定位置，越早添加的越在前。
func (m *Model) AddColumn(col *Column) error {
	if col.Name == "" {
		return errors.New("列必须存在名称")
	}

	if m.FindColumn(col.Name) != nil {
		return errColumnExists(col.Name)
	}

	m.Columns = append(m.Columns, col)

	if col.AI {
		return m.SetAutoIncrement(col)
	}
	return nil
}

// SetAutoIncrement 将 col 列设置为自增列
//
// 如果已经存在自增列或是主键，返回错误。
func (m *Model) SetAutoIncrement(col *Column) error {
	switch col.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return ErrColumnTypeError
	}

	if m.AutoIncrement != nil && m.AutoIncrement != col {
		return ErrConstraintExists(m.AIName())
	}

	if len(m.PrimaryKey) > 0 {
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

	m.PrimaryKey = append(m.PrimaryKey, col)
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

	switch col.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return ErrColumnTypeError
	}

	if !m.columnExists(col) {
		return errColumnNotFound(col.Name)
	}
	m.OCC = col
	return nil
}

// AddIndex 添加索引列
//
// 如果 name 不存在，则创建新的索引
//
// NOTE: 如果是唯一索引，则改为唯一约束
func (m *Model) AddIndex(typ Index, name string, col *Column) error {
	if typ == IndexUnique { // 唯一索引直接转为唯一约束
		return m.AddUnique(name, col)
	}

	if !m.columnExists(col) {
		return errConstraintColumnNotExists("Index", col.Name)
	}
	m.Indexes[name] = append(m.Indexes[name], col)
	return nil
}

// AddUnique 添加唯一约束的列到 name
//
// 如果 name 不存在，则创建新的约束
func (m *Model) AddUnique(name string, col *Column) error {
	if !m.columnExists(col) {
		return errConstraintColumnNotExists("Unique", col.Name)
	}
	m.Uniques[name] = append(m.Uniques[name], col)
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
func (m *Model) NewForeignKey(name string, fk *ForeignKey) error {
	if fk.Column == nil || fk.RefColName == "" || fk.RefTableName == "" {
		return fmt.Errorf("约束 %s 的 Column、RefColName 和 RefTableName 都不能为空", name)
	}

	if _, found := m.ForeignKeys[name]; found {
		return ErrConstraintExists(name)
	}

	if !m.columnExists(fk.Column) {
		return errConstraintColumnNotExists("ForeignKey", fk.Column.Name)
	}
	m.ForeignKeys[name] = fk
	return nil
}

// Sanitize 对整个对象做一次修正和检测，查看是否合法
//
// 必须要在 Model 初始化完成之后调用。
func (m *Model) Sanitize() error {
	if m.Name == "" {
		return errors.New("缺少模型名称")
	}

	if m.Type != Table && m.Type != View {
		return errors.New("无效的类型")
	}

	if len(m.PrimaryKey) == 1 {
		pk := m.PrimaryKey[0]
		if pk.HasDefault || pk.Nullable {
			return fmt.Errorf("单一主键约束的列 %s 不能为同时设置为默认值", pk.Name)
		}
	}

	for _, c := range m.Columns {
		if err := c.Check(); err != nil {
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

	if m.AutoIncrement != nil {
		names = append(names, m.AIName())
	}

	if m.PrimaryKey != nil {
		names = append(names, m.PKName())
	}

	for name := range m.Indexes {
		names = append(names, name)
	}

	for name := range m.Uniques {
		names = append(names, name)
	}

	for name := range m.ForeignKeys {
		names = append(names, name)
	}

	for name := range m.Checks {
		names = append(names, name)
	}

	sort.Strings(names)
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			return ErrConstraintExists(names[i])
		}
	}

	return nil
}
