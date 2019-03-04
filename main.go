package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

const URL = "root:admin123@tcp(127.0.0.1:3306)/db_yhui"

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

var bufPool *sync.Pool

func init() {
	bufPool = &sync.Pool{}
	bufPool.New = func() interface{} {
		return bytes.NewBuffer([]byte{})
	}
}

type DB struct {
	conn *sql.DB
}

func (db *DB) Connetion(uri string) {
	conn, err := sql.Open("mysql", uri)
	checkError(err)
	db.conn = conn
}
func (db *DB) Close() {
	db.conn.Close()
}

func getBuf() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}
func delBuf(buf *bytes.Buffer) {
	bufPool.Put(buf)
}

// TODO sql cache
func (db *DB) Find(table string, query M, field string) []M {
	buf := getBuf()
	defer delBuf(buf)

	buf.WriteString("select ")

	// field := "a b c d"
	farr := strings.Split(field, " ")
	field = strings.Join(farr, ",")

	buf.WriteString(field)

	buf.WriteString(" from ")
	buf.WriteString(table)
	param := make([]interface{}, 0)

	addQuery(buf, &param, query)
	return db.query(buf.String(), param)

}

func (db *DB) Insert(table string, data M) []M {
	buf := getBuf()
	defer delBuf(buf)

	buf.WriteString("insert into ")
	buf.WriteString(table)
	buf.WriteString(" ( ")

	field, param := decMod(data, "")
	buf.WriteString(field)

	buf.WriteString(" ) values")

	// 生成占位符(?,?),现在仅支持单行插入
	sign_len := len(data)
	sign := getSign(sign_len)

	buf.WriteString(sign)

	return db.exec(buf.String(), param)
}

func (db *DB) Update(table string, query M, data M) []M {
	buf := getBuf()
	defer delBuf(buf)

	buf.WriteString("update ")
	buf.WriteString(table)
	buf.WriteString(" set ")

	// 解析键值对
	field, param := decMod(data, "=?")
	buf.WriteString(field)

	addQuery(buf, &param, query)
	return db.exec(buf.String(), param)
}

func (db *DB) Remove(table string, query M) []M {
	buf := getBuf()
	defer delBuf(buf)

	buf.WriteString("delete from ")
	buf.WriteString(table)

	param := make([]interface{}, 0)

	addQuery(buf, &param, query)
	return db.exec(buf.String(), param)
}

func (db *DB) Query(sqlstr string, param ...interface{}) []M {
	f_type := ifType(sqlstr)
	if f_type == 1 {
		return db.query(sqlstr, param)
	} else if f_type == 2 {
		return db.exec(sqlstr, param)
	}
	return nil
}

func (db *DB) exec(sqlstr string, param []interface{}) []M {
	fmt.Println("\r\n", sqlstr, param, "\r\n")

	result, err := db.conn.Exec(sqlstr, param...)
	checkError(err)

	lastInsertId, err := result.LastInsertId()
	checkError(err)
	rowsAffected, err := result.RowsAffected()
	checkError(err)

	m1 := make(M)
	m1["lastInserId"] = lastInsertId  // 插入数据主键id
	m1["rowsAffected"] = rowsAffected // 影响行数

	return []M{m1}
}

func (db *DB) query(sqlstr string, param []interface{}) []M {
	fmt.Println("\r", sqlstr, param, "\r")

	rows, err := db.conn.Query(sqlstr, param...)
	defer rows.Close()

	checkError(err)

	fields, err := rows.Columns()
	checkError(err)

	// 提取字段的方法
	vals := make([]sql.RawBytes, len(fields))
	scanArgs := make([]interface{}, len(fields))

	for i := range vals {
		scanArgs[i] = &vals[i]
	}

	marr := make([]M, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		checkError(err)

		m1 := make(M)
		for i, col := range vals {
			if col == nil {
				m1[fields[i]] = "NULL"
			} else {
				m1[fields[i]] = string(col)
			}
		}
		marr = append(marr, m1)
	}
	err = rows.Err()
	checkError(err)

	return marr
}

var sA = []rune("select")

var sa = []rune("SELECT")

// azAZ 92 122,65 90
func ifType(sql string) int {
	str1 := []rune(sql) // rune int32 ,byte uint8
	f_type := 1         // 默认 query
	for i := 0; i < len(str1); i++ {
		if (str1[i] >= 65 && str1[i] <= 90) || (str1[i] >= 97 && str1[i] <= 122) {
			for j := 0; j < len(sA); j++ {
				if str1[i+j] != sA[j] && str1[i+j] != sa[j] {
					f_type = 2 // 未正确匹配到 第一个 为select
					break
				}
			}
			break // 执行一次就退出了, 来匹配第一个非空字符
		}
	}
	return f_type
}

func addQuery(buf *bytes.Buffer, param *[]interface{}, query M) {
	if len(query) != 0 {
		q_sql, q_param := decQuery(query, " and ", true)

		buf.WriteString(" where ")
		buf.WriteString(q_sql)

		*param = append(*param, q_param...)
	}
}

func decMod(m M, fill string) (string, []interface{}) {
	// 解析键值对
	field := getBuf()
	defer delBuf(field)

	var param = make([]interface{}, 0)
	var one = true
	for k, v := range m {
		if one {
			one = false
		} else {
			field.WriteString(",")
		}
		field.WriteString(k)
		field.WriteString(fill)
		param = append(param, v)
	}
	return field.String(), param
}

func getSign(sign_len int) string {
	sign := getBuf()
	defer delBuf(sign)

	sign.WriteString(" ( ")

	for i := 0; i < sign_len; i++ {
		if i != 0 {
			sign.WriteString(",")
		}
		sign.WriteString("?")
	}
	sign.WriteString(" ) ")
	return sign.String()
}

var oneSymbol = map[string]string{
	"$null": " is null ",
}

var onceSymbol = map[string]string{
	"$not": " not ",
}

var twoSymbol = map[string]string{
	"$eq":    " = ",
	"$ne":    " != ",
	"$lt":    " < ",
	"$lte":   " <=",
	"$gt":    " > ",
	"$gte":   " >= ",
	"$regex": " regexp ",
	"$like":  " like ",
	"$in":    " in ",
	"$nin":   " not in ",
}

var threeSymbol = map[string]bool{
	"$range": true,
}

var threeValues = map[string][]string{
	"$range": []string{" between ", " and "},
}

var arrSymbol = map[string]string{
	"$and": " and ",
	"$or":  " or ",
}

var limitSymbol = map[string]string{
	"$limit": " limit ",
}
var sortSymbol = map[string]string{
	"$sort": " order by ",
}
var sortValues = map[int]string{
	-1: " desc ",
	1:  " asc ",
}

// 键值对解析为sql
type M map[string]interface{} // map

type S []interface{} // slice

func throw(k interface{}, v interface{}, t string) {
	log.Fatalln("sql err: key=", k, ",val=", v, ",err=", t)
}

func decQuery(m M, logic string, end bool) (string, S) {
	buf := getBuf()
	defer delBuf(buf)

	buf.WriteString(" ( ")

	param := make([]interface{}, 0)

	var limit = ""
	var sort = ""

	var one = true
	for k, v1 := range m { // 第一层

		if sortSymbol[k] != "" { // S{s,1}
			switch v1.(type) {
			case S:
			default:
				throw(k, v1, "type")
			}
			v := v1.(S)
			if len(v) != 2 {
				if len(v) != 1 {
					throw(k, v1, "len")
				} else {
					v = append(v, 1) // 默认顺序
				}
			}
			if sort == "" {
				b1 := getBuf()
				defer delBuf(b1)
				b1.WriteString(sortSymbol[k])
				b1.WriteString(v[0].(string))
				t := v[1].(int)
				if sortValues[t] == "" {
					throw(k, v, "val")
				}
				b1.WriteString(sortValues[t])
				sort = b1.String()
			}
		} else if limitSymbol[k] != "" {
			b1 := getBuf()
			defer delBuf(b1)

			switch v1.(type) {
			case S:
			case int:
				b1.WriteString(limitSymbol[k])
				str_i := strconv.Itoa(v1.(int))
				b1.WriteString(str_i)
			default:

				throw(k, v1, "type")
			}
			if b1.String() == "" {
				v := v1.(S)
				if len(v) != 2 {
					throw(k, v1, "len")
				}
				if limit == "" {

					b1.WriteString(limitSymbol[k])

					b1.WriteString(string(v[0].(int)))
					b1.WriteString(" , ")
					b1.WriteString(string(v[1].(int)))

					limit = b1.String()
				}
			} else {
				if limit == "" {
					limit = b1.String()
				}
			}
		} else {
			if one {
				one = false
			} else {
				buf.WriteString(logic)
			}

			buf.WriteString(" ( ")

			if twoSymbol[k] != "" { // 判断词
				switch v1.(type) {
				case S:
				default:
					throw(k, v1, "type")
				}

				v := v1.(S)
				if len(v) != 2 {
					throw(k, v, "len")
				}

				field := v[0].(string)
				buf.WriteString(field)

				buf.WriteString(twoSymbol[k])

				switch v[1].(type) {
				case S:
					q := v[1].(S)
					sign_len := len(q)
					sign := getSign(sign_len)
					buf.WriteString(sign)

					param = append(param, q...)
				default:
					buf.WriteString("?")
					param = append(param, v[1])
				}
			} else if arrSymbol[k] != "" { // 逻辑词
				switch v1.(type) {
				case S:
				default:
					throw(k, v1, "type")
				}

				v := v1.(S)
				if len(v) == 0 {
					throw(k, v, "len")
				}
				var first = true
				for _, m1 := range v {
					switch m1.(type) {
					case M:
					default:
						throw(k, v1, "type")
					}

					n_sql, n_param := decQuery(m1.(M), " and ", false)
					if first {
						first = false
					} else {
						buf.WriteString(arrSymbol[k])
					}
					buf.WriteString(n_sql)
					param = append(param, n_param...)
				}
			} else if threeSymbol[k] { // between ? and ?
				switch v1.(type) {
				case S:
				default:
					throw(k, v1, "type")
				}

				v := v1.(S)
				if len(v) != 2 { // 判断操纵数
					throw(k, v, "len")
				}
				field := v[0].(string)

				buf.WriteString(field)

				switch v[1].(type) {
				case S:
				default:
					throw(k, v1, "type")
				}

				vals := v[1].(S) // 判断范围值
				if len(vals) != 2 {
					throw(k, v, "val")
				}
				buf.WriteString(threeValues[k][0])
				buf.WriteString(" ? ")
				buf.WriteString(threeValues[k][1])
				buf.WriteString(" ? ")
				param = append(param, vals...)
			} else if oneSymbol[k] != "" {
				switch v1.(type) {
				case string:
				default:
					throw(k, v1, "type")
				}

				buf.WriteString(" ? ")
				buf.WriteString(oneSymbol[k])

				param = append(param, v1)
			} else if onceSymbol[k] != "" {
				switch v1.(type) {
				case M:
				default:
					throw(k, v1, "type")
				}

				buf.WriteString(onceSymbol[k])
				n_sql, n_param := decQuery(v1.(M), " and ", false)
				buf.WriteString(n_sql)

				param = append(param, n_param...)
			} else { // 普通 =
				buf.WriteString(k)
				buf.WriteString(" = ?")
				param = append(param, v1)
			}
			buf.WriteString(" ) ")
		}

	}
	buf.WriteString(" ) ")
	if end {
		if sort != "" {
			buf.WriteString(sort)
		}
		if limit != "" {
			buf.WriteString(limit)
		}
	}
	//fmt.Println(buf.String(), param)
	return buf.String(), param
}

// example
func main() {
	var db DB
	db.Connetion(URL)
	defer db.Close()

	rows := db.Find("tests", M{
		"$or": S{
			M{
				"id":    1,
				"state": 1,
			},
			M{
				"id":    2,
				"state": 1,
			},
		},
		"state":  1,
		"$eq":    S{"state", 1},
		"$sort":  S{"id", -1},
		"$limit": 1,
	}, "id rank state name")
	fmt.Println(rows)

	rows = db.Insert("tests", M{
		"rank":  1,
		"state": 3,
		"name":  "aab",
	})
	fmt.Println(rows)

	rows = db.Remove("tests", M{
		"id": 14,
	})
	fmt.Println(rows)

	rows = db.Update("tests", M{
		"id": 15,
	}, M{
		"name": "aabc",
	})
	fmt.Println(rows)

	rows = db.Query(`
			update tests
			set name="xxxx"
			where id=?
		`, 1)
	fmt.Println(rows)
}

func auto_param(val ...interface{}) {
	fmt.Println(val...)
}
func set(arr *[]interface{}) {
	*arr = append(*arr, 1)
}

func main1() {
	str1 := []rune(" select")
	str2 := []rune("azAZ")
	b1 := []byte("a")
	fmt.Printf("%t", b1[0])
	fmt.Println(str1, str2)
	sA := []rune("select")
	sa := []rune("SELECT")
	NULL := []rune(" ")
	f_type := 1 // 默认 query
	for i := 0; i < len(str1); i++ {
		if str1[i] != NULL[0] { // 第一个非空字符
			for j := 0; j < len(sA); j++ {
				if str1[i+j] != sA[j] && str1[i+j] != sa[j] { // 第一个未正确匹配到 select
					f_type = 2
					break
				}
			}
			break // 执行一次就退出了
		}
	}
	fmt.Println(f_type)

	b := []interface{}{1, 2}
	set(&b)
	fmt.Println(b)

	// auto_param(1, 23, "ab")

	// str := "a,b,c"
	// a := strings.Split(str, ",")
	// var b bytes.Buffer
	// b.WriteString("(")
	// for i := 0; i < len(a); i++ {
	// 	if i != 0 {
	// 		b.WriteString(",")
	// 	}
	// 	b.WriteString("?")
	// }
	// b.WriteString(")")
	// fmt.Println(b.String()
	// var k = "k"
	// var m = map[string]string{
	// 	k: "123",
	// }
	// fmt.Println(m)

	var m = M{
		"$range": S{
			"a", S{"b", "c"},
		},
		"$null": "ax",
		"$eq":   S{"1", "2"},
	}
	decQuery(m, " and ", true)

	var str = "a b c d"
	v := strings.Split(str, " ")
	fmt.Println(v)

	// v := m["$in"].(Q)
	// switch v[1].(type) {
	// case Q:
	// 	fmt.Println("q")
	// default:
	// 	fmt.Println("xx")
	// }

}
