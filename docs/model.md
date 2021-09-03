每一个数据模型都可以定义为 Go 结构体。当通过 DB 实例第一次接触到该对象时
（比如 `Insert`、`Create` 等），会生成模型数据。

```go
type User struct {
    ID       int64          `orm:"name(id);ai"`
    Name     string         `orm:"name(name);len(20);index(user_index_name)"`
    Username string         `orm:"name(username);len(20);unique(user_unique_username)"`
    Nickname sql.NullString `orm:"name(nickname);len(20);nullable"`
    Last     *Last          `orm:"name(last);len(-1);default(192.168.1.1,2019-07-29T17:11:01)"`
}

func(u *User) TableName() string { return "#users" }

func(u *User) ApplyModel(m*core.Model) error {
    m.Meta["mysql_charset"] = []string{"utf8"}
    return m.NewCheck("id_great_zero", "id>0")
}
```

结构体中的字段与数据表中列的关联通过名为 orm 的 struct tag 进行设置。
struct tag 中的格式为 `key(val);key(v1,v2)`，其中 key 属性名，val 等为该属性对应的值列表。

目前支持在 struct tag 支持以下属性：

### 属性

#### name(fieldName)

指定当前字段在数据表中的名称，如果未指定，则和字段名相同。
只有可导出的字段才有效果。

#### len(l1, l2)

指定字段的长度。比如 mysql 中的int(5),varchar(255),double(1,2),
不支持该特性的数据，将会忽略该标签的内容，比如 sqlite3。

NOTE:字符串类型必须指定长度，若长度过大或是将长度设置了 -1，
会使用类似于 TEXT 等不定长的形式表达。

如果是日期类型，则第一个可选参数表示日期精度。

#### nullable(true|false)

相当于定义表结构时的 NULL，建议尽量少用该属性。

#### pk

主键，支持联合主键，给多个字段加上 pk 的 struct tag 即可。

主键约束不能自定义约束名。
如果在某些地方需要用到约束名，可以调用 core.PKName() 生成约束名。

#### ai

自增，若指定了自增列，则将自动取消其它的 pk 设置。无法指定起始值和步长。
可手动设置一个非零值来更改某条数据的 AI 行为。

#### unique(index_name)

唯一约束，支持联合索引，index_name 为约束名，会将 index_name
一样的字段定义为一个联合唯一约束。

#### index(index_name)

普通的关键字索引，同 unique 一样会将名称相同的索引定义为一个联合索引。

#### occ(true|false)

当前列作为乐观锁字段。

作为乐观锁的字段，其值表示的是线上数据的值，在更新时，会自动给线上的值加 1。

#### default(value)

指定默认值。相当于定义表结构时的 DEFAULT。

内置类型的格式，Bool 为 true 和 false，time 为 time.RFC3339

自定义类型，用户可以自已实现 DefaultParser 作为解析方式。

#### fk(fk_name,refTable,refColName,updateRule,deleteRule)

定义物理外键，最少需要指定 fk_name、refTable 和 refColName 三个值。分别对应约束名，
引用的表和引用的字段，updateRule,deleteRule，在不指定的情况下，使用数据库的默认值。
refTable 如果需要表名前缀，需要添加 # 符号。

#### check(chk_name, expr)

check 约束。chk_name 为约束名，expr 为该约束的表达式。
check 约束只能在 `ApplyModeler` 接口中指定，而不是像其它约束一样，
通过字段的 struct tag 指定。
因为 check 约束的表达式可以通过 and 或是 or 等符号连接多条基本表达式，
在字段 struct tag 中指定会显得有点怪异。

### 接口

#### TableNamer

指定表名，视图和数据表都需要实现此接口。

#### ApplyModeler

通过 ApplyModeler 接口可以指定一些表级别的属性值。

#### Viewer

如果需要将模型定义为视图，则需要实现此接口，
Viewer 接口返回一条 `SELECT` 语句，用于指定创建视图时的 `SELECT` 部分语句。
实现都需要保证接口中返回的列与模型中列的定义要对应。

在视图模式下，部分功能会不可用，比如 check 约束、索引等。
但是 AI、PK 和唯一索引，仍然在查询时，被用来当作唯一查询条件。

#### DefaultParser

DefaultParser 用于自定义类型的数据作为列时，如果需要指定默认值，
可以实现该接口。

在解析默认值时，如果不存在 `DefaultParser` 接口，也会尝试采用 `sql.Scanner` 接口，
如果两者都不存在，则会采用 github.com/issue9/conv.Value 进行强转换。

比如以下代码就可以实现将一个对象以 JSON 字符串的形式保存在数据库中，
而默认值的设置，可以是 `ip,time` 的形式，以逗号作简单分隔。

```go
// Last 用户最后一次访问信息
type Last struct {
    IP      string    `orm:"name(ip);len(50)"`
    Created time.Time `orm:"name(creator)"`
}

func(l *Last) ParseDefault(v string) (err error) {
    fields := strings.Split(v, ",")
    if len(fields) != 2 {
        return errors.New("格式不正确")
    }

    l.Created, err = time.Parse(time.RFC3339, fields[1])
    if err != nil {
        return err
    }

    l.IP = fields[0]

    return nil
}

func(l *Last) Value() (driver.Value, error) {
    data, err := json.Marshal(l)
    if err != nil {
        return nil, err
    }

    return string(data), nil
}

func(l *Last) Scan(v interface{}) error {
    str, ok := v.(string)
    if !ok {
        return errors.New("无效的类型")
    }

    return json.Unmarshal([]byte(str), l)
}
```

#### BeforeUpdater/BeforeInserter/AfterFetcher

分别用于在更新和插入数据之前和从数据库获取数据之后被执行的方法。
一般用于特定内容的生成，比如：

```go
type User struct {
    Created  time.Time `orm:"name(created)"`  // 创建时间
    Modified time.Time `orm:"name(modified)"` // 修改时间
    Avatar   string    `orm:"name(avatar);len(1024)"`
}

// 每次插入数据，都将 created 和 modified 设置为当前时间
func(u *User) BeforeInsert() error {
    u.Created = time.Now()
    u.Modified = u.Created
    return nil
}

// 每次更新前，都修改 modified 的值为当前时间
func(u *User) BeforeUpdate() error {
    u.Modified = time.Now()
    return nil
}

// 如果不存在头像信息，则给定一个默认图片地址
func(u *User) AfterFetch() error {
    if u.Avatar == "" {
        u.Avatar = "/assets/default-avatar.png"
    }

    return nil
}

```
