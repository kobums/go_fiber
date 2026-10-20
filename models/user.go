package models

import (
	"project/config"

	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)


type User struct {
	Id					int64 `json:"id"`
	Passwd				string `json:"passwd"`
	Name				string `json:"name"`
	Email				string `json:"email"`
	Date				string `json:"date"`

	Extra				map[string]interface{} `json:"extra"`
}

type UserManager struct {
	Conn	*sql.DB
	Tx		*sql.Tx
	Result	*sql.Result
	Index	string
}

func (c *User) AddExtra(key string, value interface{}) {
	c.Extra[key] = value
}

func NewUserManager(conn interface{}) *UserManager {
	var item UserManager

	if conn == nil {
		item.Conn = NewConnection()
	} else {
		if v, ok := conn.(*sql.DB); ok {
			item.Conn = v
			item.Tx = nil
		} else {
			item.Tx = conn.(*sql.Tx)
			item.Conn = nil
		}
	}
	
	item.Index = ""
	return &item
}

func (p *UserManager) Close() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}

func (p *UserManager) SetIndex(index string) {
	p.Index = index
}

func (p *UserManager) Exec(query string, params ...interface{}) (sql.Result, error) {
	if p.Conn != nil {
		return p.Conn.Exec(query, params...)
	} else {
		return p.Tx.Exec(query, params...)
	}
}

func (p *UserManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
	if p.Conn != nil {
		return p.Conn.Query(query, params...)
	} else {
		return p.Tx.Query(query + " FOR UPDATE", params...)
	}
}

func (p *UserManager) GetQeury() string {
	ret := ""

	str := "select u_id, u_passwd, u_name, u_email, u_date from user_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ")"
	}

	ret += "where 1=1 "

	return ret;
}

func (p *UserManager) GetQeurySelect() string {
	ret := ""

	str := "select count(*) from user_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ") "
	}

	return ret;
}

func (p *UserManager) Truncate() error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "truncate user_tb "
	p.Exec(query)

	return nil
}

func (p *UserManager) Insert(item *User) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	if item.Date == "" {
		t := time.Now()
		item.Date = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	}

	query := ""
	var res sql.Result
	var err error
	if item.Id > 0 {
		query = "insert into user_tb (u_id, u_passwd, u_name, u_email, u_date) values (?, ?, ?, ?, ?)"
		res, err = p.Exec(query , item.Id, item.Passwd, item.Name, item.Email, item.Date)
	} else {
		query = "insert into user_tb (u_passwd, u_name, u_email, u_date) values (?, ?, ?, ?)"
		res, err = p.Exec(query , item.Passwd, item.Name, item.Email, item.Date)
	}

	if err == nil {
		p.Result = &res
	} else {
		log.Println(err)
		p.Result = nil
	}

	return err
}

func (p *UserManager) Delete(id int64) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "delete from user_tb where u_id = ?"
	_, err := p.Exec(query, id)

	return err
}

func (p *UserManager) Update(item *User) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "update user_tb set u_passwd = ?, u_name = ?, u_email = ?, u_date = ? where u_id = ?"
	_, err := p.Exec(query, item.Passwd, item.Name, item.Email, item.Date, item.Id)

	return err
}

func (p *UserManager) GetIdentity() int64 {
	if p.Result == nil && p.Tx == nil {
		return 0
	}

	id, err := (*p.Result).LastInsertId()

	if err != nil {
		return 0
	} else {
		return id
	}
}

func (p *User) InitExtra() {
	p.Extra = map[string]interface{}{

	}
}

func (p *UserManager) ReadRow(rows *sql.Rows) *User {
	var item User
	var err error

	if rows.Next() {
		err = rows.Scan(&item.Id, &item.Passwd, &item.Name, &item.Email, &item.Date)
	} else {
		return nil
	}
	if err != nil {
		return nil
	} else {
		item.InitExtra()
		return &item
	}
}

func (p *UserManager) ReadRows(rows *sql.Rows) *[]User {
	var items []User

	for rows.Next() {
		var item User

		err := rows.Scan(&item.Id, &item.Passwd, &item.Name, &item.Email, &item.Date)

		if err != nil {
			log.Printf("ReadRows error : %v\n", err)
			break
		}

		item.InitExtra()

		items = append(items, item)
	}
	return &items
}

func (p *UserManager) Get(id int64) *User {
	if p.Conn == nil && p.Tx == nil {
		return nil
	}

	query := p.GetQeury() + " and u_id = ?"

	rows, err := p.Query(query, id)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		return nil
	}

	defer rows.Close()

	return p.ReadRow(rows)
}

func (p *UserManager) Count(args []interface{}) int {
	if p.Conn == nil && p.Tx == nil {
		return 0
	}

	var params []interface{}
	query := p.GetQeurySelect() + " where 1=1 "

	for _, arg := range args {
		switch v := arg.(type) {
		case Where:
			clause, whereParams := WhereClause("u_", v)
			query += clause
			params = append(params, whereParams...)
		}
	}

	rows, err := p.Query(query, params...)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		return 0
	}

	defer rows.Close()

	if !rows.Next() {
		return 0
	}

	cnt := 0
	err = rows.Scan(&cnt)

	if err != nil {
		return 0
	} else {
		return cnt
	}
}

func (p *UserManager) Find(args []interface{}) *[]User {
	if p.Conn == nil && p.Tx == nil {
		var items []User
		return &items
	}

	var params []interface{}
	query := p.GetQeury()

	page := 0
	pagesize := 0
	orderby := ""

	for _, arg := range args {
		switch v := arg.(type) {
		case PagingType:
			item := v
			page = item.Page
			pagesize = item.Pagesize
			break
		case OrderingType:
			item := v
			orderby = item.Order
			break
		case LimitType:
			item := v
			page = 1
			pagesize = item.Limit
			break
		case OptionType:
			item := v
			if item.Limit > 0 {
				page = 1
				pagesize = item.Limit
			} else {
				page = item.Page
				pagesize = item.Pagesize
			}
			orderby = item.Order
			break
		case Where:
			clause, whereParams := WhereClause("u_", v)
			query += clause
			params = append(params, whereParams...)
		}
	}

	startpage := (page -1) * pagesize

	if page > 0 && pagesize > 0 {
		if orderby == "" {
			orderby = "u_id"
		} else {
			orderby = "u_" + orderby
		}
		query += " order by " + orderby
		if config.Database == "mysql" {
			query += " limit ? offset ?"
			params = append(params, pagesize)
			params = append(params, startpage)
		} else if config.Database == "mssql" || config.Database == "sqlserver" {
			query += "OFFSET ? ROWS FITCH NEXT ? ROWS ONLY"
			params = append(params, startpage)
			params = append(params, pagesize)
		}
	} else {
		if orderby == "" {
			orderby = "u_id"
		} else {
			orderby = "u_" + orderby
		}
		query += " order by " + orderby
	}

	rows, err := p.Query(query, params...)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		var items []User
		return &items
	}

	defer rows.Close()

	return p.ReadRows(rows)
}

func (p *UserManager) GetByEmail(email string, args ...interface{}) *User {
    if email != "" {
        args = append(args, Where{Column:"email", Value:email, Compare:"="})        
    }
    
    items := p.Find(args)

    if items != nil && len(*items) > 0 {
        return &(*items)[0]
    } else {
        return nil
    }
}