package rest

import (
	"project/controllers"
	"project/models"
)

type BoardController struct {
	controllers.Controller
}

func (c *BoardController) Index(page int, pagesize int) {
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)

    var args []interface{}

    title := c.Query("title")
    if title != "" {
        args = append(args, models.Where{Column:"title", Value:title, Compare:"="})
    }

    content := c.Query("content")
    if content != "" {
        args = append(args, models.Where{Column:"content", Value:content, Compare:"="})
    }

    user := c.Query("user")
    if user != "" {
        args = append(args, models.Where{Column:"user", Value:user, Compare:"="})
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

func (c *BoardController) Read(id int64) {
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
	item := manager.Get(id)

    c.Set("item", item)
}

func (c *BoardController) Insert(item *models.Board) {
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
	manager.Insert(item)

    id := manager.GetIdentity()
    c.Result["id"] = id
    item.Id = id
}

func (c *BoardController) Update(item *models.Board) {
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
	manager.Update(item)
}

func (c *BoardController) Delete(item *models.Board) {
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
	manager.Delete(item.Id)
}