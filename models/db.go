package models

import (
	"database/sql"
	"log"
	"project/config"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type PagingType struct {
	Page		int
	Pagesize	int
}

type OrderingType struct {
	Order 		string
}

type LimitType struct {
	Limit int
}

type OptionType struct {
	Page		int
	Pagesize	int
	Order		string
	Limit		int
}

type Where struct {
	Column		string
	Value		interface{}
	Compare		string
}

func Paging(page int , pagesize int) PagingType {
	return PagingType{Page: page, Pagesize: pagesize}
}

// orderingPattern 은 ORDER BY 절로 허용하는 형식이다: 콤마로 구분된 "컬럼명 [asc|desc]" 목록.
// orderby 는 SQL 에 그대로 이어붙으므로 이 형식이 아니면 버린다(SQL 인젝션 방지).
var orderingPattern = regexp.MustCompile(`(?i)^\s*[a-z_][a-z0-9_]*(\s+(asc|desc))?(\s*,\s*[a-z_][a-z0-9_]*(\s+(asc|desc))?)*\s*$`)

func Ordering(order string) OrderingType {
	if order != "" && !orderingPattern.MatchString(order) {
		log.Printf("rejected invalid orderby: %q\n", order)
		return OrderingType{Order: ""}
	}
	return OrderingType{Order: order}
}

func Limit(limit int) LimitType {
	return LimitType{Limit: limit}
}

func GetConnection() *sql.DB {
	r1, err := sql.Open(config.Database, config.ConnectionString)
	if err != nil {
		log.Println("Database Connect Error")
		return nil
	}

	r1.SetMaxOpenConns(1000)
	r1.SetMaxIdleConns(10)
	r1.SetConnMaxIdleTime(5 * time.Minute)

	return r1
}


func NewConnection() *sql.DB {
	db := GetConnection()

	if db != nil {
		return db
	}

	time.Sleep(100 * time.Millisecond)

	db = GetConnection()

	if db != nil {
		return db
	}

	time.Sleep(500 * time.Millisecond)

	db = GetConnection()

	if db != nil {
		return db
	}

	time.Sleep(1 * time.Second)

	db = GetConnection()

	if db != nil {
		return db
	}

	time.Sleep(2 * time.Second)

	db = GetConnection()

	return db
}

// whereColumnPattern 은 Where.Column 으로 허용하는 형식이다(SQL 식별자).
// 컬럼명과 연산자는 자리표시자로 바인딩할 수 없어 쿼리에 그대로 이어붙으므로,
// 이 형식이 아니면 SQL 에 넣지 않는다.
var whereColumnPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// whereCompares 는 Where.Compare 로 허용하는 연산자 목록이다(소문자 기준).
var whereCompares = map[string]bool{
	"=": true, "!=": true, "<>": true, "<": true, "<=": true, ">": true, ">=": true,
	"like": true, "not like": true, "in": true, "between": true,
}

// whereValues 는 in / between 값을 바인딩 값 목록으로 펼친다.
// 슬라이스·배열은 원소별로, 문자열은 콤마로 나눠서, 그 외는 값 하나로 본다.
func whereValues(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	if s, ok := value.(string); ok {
		var ret []interface{}
		for _, part := range strings.Split(s, ",") {
			if part = strings.TrimSpace(part); part != "" {
				ret = append(ret, part)
			}
		}
		return ret
	}
	if _, ok := value.([]byte); ok {
		return []interface{}{value}
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return []interface{}{value}
	}
	ret := make([]interface{}, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		ret = append(ret, rv.Index(i).Interface())
	}
	return ret
}

// WhereClause 는 Where 하나를 " and <prefix><컬럼> <연산자> ?" 로 만들고 바인딩 값을 돌려준다.
// 기존 생성 코드와 같이 in 조건은 항상 <prefix>id 에 걸고, 값은 원소마다 ? 로 바인딩한다.
// 컬럼/연산자/값 형식이 어긋나면 조건을 버리지 않고 " and 1=0" 을 돌려준다(fail-closed:
// 조건을 빼면 오히려 더 많은 행이 조회될 수 있으므로).
func WhereClause(prefix string, item Where) (string, []interface{}) {
	compare := strings.ToLower(strings.TrimSpace(item.Compare))
	if !whereColumnPattern.MatchString(item.Column) || !whereCompares[compare] {
		log.Printf("rejected invalid where: column=%q compare=%q\n", item.Column, item.Compare)
		return " and 1=0", nil
	}

	switch compare {
	case "in":
		values := whereValues(item.Value)
		if len(values) == 0 {
			return " and 1=0", nil
		}
		marks := strings.TrimSuffix(strings.Repeat("?, ", len(values)), ", ")
		return " and " + prefix + "id in (" + marks + ")", values
	case "between":
		values := whereValues(item.Value)
		if _, isString := item.Value.(string); isString || len(values) != 2 {
			return " and 1=0", nil
		}
		return " and " + prefix + item.Column + " between ? and ?", values
	case "like":
		value := item.Value
		if s, ok := value.(string); ok {
			value = "%" + s + "%"
		}
		return " and " + prefix + item.Column + " like ?", []interface{}{value}
	default:
		return " and " + prefix + item.Column + " " + compare + " ?", []interface{}{item.Value}
	}
}
