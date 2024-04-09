package models

import (
	"project/config"

	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)


type Board struct {
	Id					int64 `json:"id"`
	Title				string `json:"title"`
	Content				string `json:"content"`
	User				int64 `json:"user"`
	Date				string `json:"date"`

	Extra				map[string]interface{} `json:"extra"`
}

type BoardManager struct {
	Conn	*sql.DB
	Tx		*sql.Tx
	Result	*sql.Result
	Index	string
}

func (c *Board) AddExtra(key string, value interface{}) {
	c.Extra[key] = value
}

func NewBoardManager(conn interface{}) *BoardManager {
	var item BoardManager

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

func (p *BoardManager) Close() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}

func (p *BoardManager) SetIndex(index string) {
	p.Index = index
}

func (p *BoardManager) Exec(query string, params ...interface{}) (sql.Result, error) {
	if p.Conn != nil {
		return p.Conn.Exec(query, params...)
	} else {
		return p.Tx.Exec(query, params...)
	}
}

func (p *BoardManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
	if p.Conn != nil {
		return p.Conn.Query(query, params...)
	} else {
		return p.Tx.Query(query + " FOR UPDATE", params...)
	}
}

func (p *BoardManager) GetQeury() string {
	ret := ""

	str := "select b_id, b_title, b_content, b_user, b_date, u_id, u_email, u_name, u_passwd, u_date from board_tb, user_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ")"
	}

	ret += "where 1=1 "

	ret += "and b_user = u_id "

	return ret;
}

func (p *BoardManager) GetQeurySelect() string {
	ret := ""

	str := "select count(*) from board_tb, user_tb "

	if p.Index == "" {
		ret = str
	} else {
		ret = str + " use index(" + p.Index + ") "
	}

	ret += "whrer 1=1 "

	ret += "and b_user = u_id "

	return ret;
}

func (p *BoardManager) Truncate() error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "truncate board_tb "
	p.Exec(query)

	return nil
}

func (p *BoardManager) Insert(item *Board) error {
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
		query = "insert into board_tb (b_id, b_title, b_content, b_user, b_date) values (?, ?, ?, ?, ?)"
		res, err = p.Exec(query , item.Id, item.Title, item.Content, item.User, item.Date)
	} else {
		query = "insert into board_tb (b_title, b_content, b_user, b_date) values (?, ?, ?, ?)"
		res, err = p.Exec(query , item.Title, item.Content, item.User, item.Date)
	}

	if err == nil {
		p.Result = &res
	} else {
		log.Println(err)
		p.Result = nil
	}

	return err
}

func (p *BoardManager) Delete(id int64) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "delete from board_tb where b_id = ?"
	_, err := p.Exec(query, id)

	return err
}

func (p *BoardManager) Update(item *Board) error {
	if p.Conn == nil && p.Tx == nil {
		return errors.New("Connection Error")
	}

	query := "update board_tb set b_title = ?, b_content = ?, b_user = ?, b_date = ? where b_id = ?"
	_, err := p.Exec(query, item.Title, item.Content, item.User, item.Date, item.Id)

	return err
}

func (p *BoardManager) GetIdentity() int64 {
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

func (p *Board) InitExtra() {
	p.Extra = map[string]interface{}{

	}
}

func (p *BoardManager) ReadRow(rows *sql.Rows) *Board {
	var item Board
	var err error

	var _user User

	if rows.Next() {
		err = rows.Scan(&item.Id, &item.Title, &item.Content, &item.User, &item.Date, &_user.Id, &_user.Email, &_user.Name, &_user.Passwd, &_user.Date)
	} else {
		return nil
	}
	if err != nil {
		return nil
	} else {
		item.InitExtra()
		_user.InitExtra()
		item.AddExtra("user", _user)
		return &item
	}
}

func (p *BoardManager) ReadRows(rows *sql.Rows) *[]Board {
	var items []Board

	for rows.Next() {
		var item Board
		var _user User

		err := rows.Scan(&item.Id, &item.Title, &item.Content, &item.User, &item.Date, &_user.Id, &_user.Email, &_user.Name, &_user.Passwd, &_user.Date)

		if err != nil {
			log.Printf("ReadRows error : %v\n", err)
			break
		}

		item.InitExtra()
		_user.InitExtra()
		item.AddExtra("user", _user)

		items = append(items, item)
	}
	return &items
}

func (p *BoardManager) Get(id int64) *Board {
	if p.Conn == nil && p.Tx == nil {
		return nil
	}

	query := p.GetQeury() + " and b_id = ?"

	query += " and b_user = u_id "

	rows, err := p.Query(query, id)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		return nil
	}

	defer rows.Close()

	return p.ReadRow(rows)
}

func (p *BoardManager) Count(args []interface{}) int {
	if p.Conn == nil && p.Tx == nil {
		return 0
	}

	var params []interface{}
	query := p.GetQeurySelect() + " where 1=1 "

	for _, arg := range args {
		switch v := arg.(type) {
		case Where:
			item := v

			if item.Compare == "in" {
				query += " and b_id in (" + strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]") + ")"
			} else if item.Compare == "between" {
				query += " and b_" + item.Column + " between ? and ?"

				s := item.Value.([2]string)
				params = append(params, s[0])
				params = append(params, s[1])
			} else {
				query += " and b_" + item.Column + " " + item.Compare + " ?"
				if item.Compare == "like" {
					params = append(params, "%" + item.Value.(string) + "%")
				} else {
					params = append(params, item.Value)
				}
			}
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

func (p *BoardManager) Find(args []interface{}) *[]Board {
	if p.Conn == nil && p.Tx == nil {
		var items []Board
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
			item := v

			if item.Compare == "in" {
				query += " and b_id in (" + strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]") + ")"
			} else if item.Compare == "between" {
				query += " and b_" + item.Column + " between ? and ?"

				s := item.Value.([2]string)
				params = append(params, s[0])
				params = append(params, s[1])
			} else {
				query += " and b_" + item.Column + " " + item.Compare + " ?"
				if item.Compare == "like" {
					params = append(params, "%" + item.Value.(string) + "%")
				} else {
					params = append(params, item.Value)
				}
			}
		}
	}

	startpage := (page -1) * pagesize

	if page > 0 && pagesize > 0 {
		if orderby == "" {
			orderby = "b_id"
		} else {
			orderby = "b_" + orderby
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
			orderby = "b_id"
		} else {
			orderby = "b_" + orderby
		}
		query += " order by " + orderby
	}

	rows, err := p.Query(query, params...)

	if err != nil {
		log.Printf("query error : %v, %v\n", err, query)
		var items []Board
		return &items
	}

	defer rows.Close()

	return p.ReadRows(rows)
}

// func (p *BoardManager) GetByUser(user string, args ...interface{}) *Board {
//     if user != "" {
//         args = append(args, Where{Column:"user", Value:user, Compare:"="})        
//     }
    
//     items := p.Find(args)

//     if items != nil && len(*items) > 0 {
//         return &(*items)[0]
//     } else {
//         return nil
//     }
// }