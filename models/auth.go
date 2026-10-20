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


type Auth struct {
	Id					int64 `json:"id"`
	User			int64 `json:"user"`
	Token				string `json:"token"`
	Date				string `json:"date"`

	Extra				map[string]interface{} `json:"extra"`
}

type AuthManager struct {
	Conn	*sql.DB
	Tx		*sql.Tx
	Result	*sql.Result
	Index	string
}

func (c *Auth) AddExtra(key string, value interface{}) {
	c.Extra[key] = value
}

func NewAuthManager(conn interface{}) *AuthManager {
	var item AuthManager

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

func (p *AuthManager) Close() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}

func (p *AuthManager) SetIndex(index string) {
	p.Index = index
}

func (p *AuthManager) Exec(query string, params ...interface{}) (sql.Result, error) {
	if p.Conn != nil {
		return p.Conn.Exec(query, params...)
	} else {
		return p.Tx.Exec(query, params...)
	}
}

func (p *AuthManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
	if p.Conn != nil {
		return p.Conn.Query(query, params...)
	} else {
		return p.Tx.Query(query + " FOR UPDATE", params...)
	}
}

func (p *AuthManager) GetQeury() string {
	ret := ""

	str := "select a_id, a_user, a_token, a_date from auth_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ")"
	}

	ret += "where 1=1 "

	return ret;
}

func (p *AuthManager) GetQeurySelect() string {
	ret := ""

	str := "select count(*) from auth_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ") "
	}

	return ret;
}

func (p *AuthManager) Truncate() error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "truncate auth_tb "
	p.Exec(query)

	return nil
}

func (p *AuthManager) Insert(item *Auth) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	if item.Date == "" {
		t := time.Now()
		item.Date = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	}

	log.Println(item.Id, item.User, item.Token, item.Date)

	query := ""
	var res sql.Result
	var err error
	if item.Id > 0 {
		query = "insert into auth_tb (a_id, a_user, a_token, a_date) values (?, ?, ?, ?)"
		res, err = p.Exec(query , item.Id, item.User, item.Token, item.Date)
	} else {
		query = "insert into auth_tb (a_user, a_token, a_date) values (?, ?, ?)"
		res, err = p.Exec(query , item.User, item.Token, item.Date)
	}

	if err == nil {
		p.Result = &res
	} else {
		p.Result = nil
	}

	return err
}

func (p *AuthManager) Delete(id int64) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "delete from auth_tb where a_id = ?"
	_, err := p.Exec(query, id)

	return err
}

func (p *AuthManager) Update(item *Auth) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	if item.Date == "" {
		t := time.Now()
		item.Date = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	}

	query := "update auth_tb set a_user = ?, a_token = ?, a_date = ? where a_id = ?"
	_, err := p.Exec(query, item.User, item.Token, item.Date, item.Id)

	return err
}

func (p *AuthManager) GetIdentity() int64 {
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

func (p *Auth) InitExtra() {
	p.Extra = map[string]interface{}{

	}
}

func (p *AuthManager) ReadRow(rows *sql.Rows) *Auth {
	var item Auth
	var err error

	if rows.Next() {
		err = rows.Scan(&item.Id, &item.User, &item.Token, &item.Date)
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

func (p *AuthManager) ReadRows(rows *sql.Rows) *[]Auth {
	var items []Auth

	for rows.Next() {
		var item Auth

		err := rows.Scan(&item.Id, &item.User, &item.Token, &item.Date)

		if err != nil {
			log.Printf("ReadRows error : %v\n", err)
			break
		}

		item.InitExtra()

		items = append(items, item)
	}
	return &items
}

func (p *AuthManager) Get(id int64) *Auth {
	if p.Conn == nil && p.Tx == nil {
		return nil
	}

	query := p.GetQeury() + " and a_id = ?"

	rows, err := p.Query(query, id)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		return nil
	}

	defer rows.Close()

	return p.ReadRow(rows)
}

func (p *AuthManager) Count(args []interface{}) int {
	if p.Conn == nil && p.Tx == nil {
		return 0
	}

	var params []interface{}
	query := p.GetQeurySelect() + " where 1=1 "

	for _, arg := range args {
		switch v := arg.(type) {
		case Where:
			clause, whereParams := WhereClause("a_", v)
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

func (p *AuthManager) Find(args []interface{}) *[]Auth {
	if p.Conn == nil && p.Tx == nil {
		var items []Auth
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
			clause, whereParams := WhereClause("a_", v)
			query += clause
			params = append(params, whereParams...)
		}
	}

	startpage := (page -1) * pagesize

	if page > 0 && pagesize > 0 {
		if orderby == "" {
			orderby = "a_id"
		} else {
			orderby = "a_" + orderby
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
			orderby = "a_id"
		} else {
			orderby = "a_" + orderby
		}
		query += " order by " + orderby
	}

	rows, err := p.Query(query, params...)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		var items []Auth
		return &items
	}

	defer rows.Close()

	return p.ReadRows(rows)
}

func (p *AuthManager) GetByUser(user int64, args ...interface{}) *Auth {
    if user != 0 {
        args = append(args, Where{Column:"user", Value:user, Compare:"="})        
    }
    
    items := p.Find(args)

    if items != nil && len(*items) > 0 {
        return &(*items)[0]
    } else {
        return nil
    }
}