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

    email := c.Query("email")
    if email != "" {
        args = append(args, models.Where{Column:"email", Value:email, Compare:"like"})
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
	// 응답에 비밀번호(해시)를 노출하지 않는다.
	if items != nil {
		for i := range *items {
			(*items)[i].Passwd = ""
		}
	}
	c.Set("items", items)

    total := manager.Count(args)
	c.Set("total", total)
}

func (c *UserController) Read(id int64) {
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	item := manager.Get(id)
	// 응답에 비밀번호(해시)를 노출하지 않는다.
	if item != nil {
		item.Passwd = ""
	}

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