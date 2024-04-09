package rest

import (
	"project/controllers"
	"project/models"
)

type UserController struct {
	controllers.Controller
}

func (c *UserController) Index(page int, pagesize int) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}

    name := c.Query("name")
    if name != "" {
        args = append(args, models.Where{Column:"name", Value:name, Compare:"="})
    }

    passwd := c.Query("passwd")
    if passwd != "" {
        args = append(args, models.Where{Column:"passwd", Value:passwd, Compare:"="})
    }

    email := c.Query("email")
    if email != "" {
        args = append(args, models.Where{Column:"email", Value:email, Compare:"="})
    }
    startdate := c.Query("startdate")
    enddate := c.Query("enddate")
    if startdate != "" && enddate != "" {
        var v [2]string
        v[0] = startdate
        v[1] = enddate
        args = append(args, models.Where{Column:"date", Value:v, Compare:"between"})
    } else if  startdate != "" {
        args = append(args, models.Where{Column:"date", Value:startdate, Compare:">="})
    } else if  enddate != "" {
        args = append(args, models.Where{Column:"date", Value:enddate, Compare:"<="})
    }
    
    if page != 0 && pagesize != 0 {
        args = append(args, models.Paging(page, pagesize))
    }

    orderby := c.Query("orderby")
    if orderby == "desc" {
        // if page != 0 && pagesize != 0 {
            orderby = "id desc"
        // }
    } else {
		orderby = ""
	}

    if orderby != "" {
        args = append(args, models.Ordering(orderby))
    }

	items := manager.Find(args)
	c.Set("items", items)

    total := manager.Count(args)
	c.Set("total", total)
}

func (c *UserController) Read(id int64) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	item := manager.Get(id)

    c.Set("item", item)
}

func (c *UserController) Insert(item *models.User) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	manager.Insert(item)

    id := manager.GetIdentity()
    c.Result["id"] = id
    item.Id = id
}

func (c *UserController) Update(item *models.User) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	manager.Update(item)
}

func (c *UserController) Delete(item *models.User) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	manager.Delete(item.Id)
}