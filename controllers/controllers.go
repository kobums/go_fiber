package controllers

import (
	"database/sql"
	"net/http"
	"project/global"
	"project/models"
	"time"

	"github.com/CloudyKit/jet/v3"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	Context    *fiber.Ctx
	Vars       jet.VarMap
	Result     fiber.Map
	Connection *sql.DB
	Session    *models.User
	Current    string
	Code       int

	Date string

	Page     int
	Pagesize int
}


func NewController(f *fiber.Ctx) *Controller {
	var ctl Controller
	ctl.Init(f)
	return &ctl
}

func (c *Controller) Init(f *fiber.Ctx) {
	c.Context = f
	c.Vars = make(jet.VarMap)
	c.Result = make(fiber.Map)
	c.Result["code"] = "ok"
	c.Connection = c.NewConnection()
	c.Code = http.StatusOK

	user, ok := f.Locals("user").(*models.User)

	if ok {
		c.Session = user
	} else {
		c.Session = nil
	}

	c.Date = global.GetDate(time.Now())

	c.Set("_t", time.Now().UnixNano())
}

func (c *Controller) Set(name string, value interface{}) {
	c.Result[name] = value
}


func (c *Controller) Get(name string) string {
	if c.Context.Request().Header.String() == "GET" {
		return c.Query(name)
	} else {
		return c.Post(name)
	}
}


func (c *Controller) Post(name string) string {
	return c.Context.FormValue(name)
}

func (c *Controller) NewConnection() *sql.DB {
	if c.Connection != nil {
		return c.Connection
	}

	c.Connection = models.NewConnection()
	return c.Connection
}


func (c *Controller) Query(name string) string {
	return c.Context.Query(name)
}


func (c *Controller) Close() {
	if c.Connection != nil {
		c.Connection.Close()
		c.Connection = nil
	}
}