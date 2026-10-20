package router

import (
	"project/controllers/rest"
	"project/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func SetRouter(app *fiber.App) {
	// Deprecated: 쿼리스트링 로그인은 기존 클라이언트 호환용. 새 클라이언트는 POST /api/jwt (auth.go SetAuthRouter).
	app.Get("/api/jwt", func(ctx *fiber.Ctx) error {
		email := ctx.Query("email")
		passwd := ctx.Query("passwd")
		return ctx.JSON(JwtAuth(email, passwd))
	})
	app.Get("/api/jwt/token", func(ctx *fiber.Ctx) error {
		token := ctx.Get("Authorization")
		return ctx.JSON(JwtToken(token))
	})
	apiGroup := app.Group("/api")
	apiGroup.Use(JwtAuthRequired())
	{
		apiGroup.Get("/user/:id", func(ctx *fiber.Ctx) error {
			id_, _ := strconv.ParseInt(ctx.Params("id"), 10, 64)
			var controller rest.UserController
			controller.Init(ctx)
			controller.Read(id_)
			controller.Close()
			return ctx.JSON(controller.Result)
		})

		apiGroup.Get("/me", func(ctx *fiber.Ctx) error {
			token := ctx.Get("Authorization")
			return ctx.JSON(JwtMe(token))
		})

		apiGroup.Get("/user", func(ctx *fiber.Ctx) error {
			page_, _ := strconv.Atoi(ctx.Query("page"))
			pagesize_, _ := strconv.Atoi(ctx.Query("pagesize"))
			var controller rest.UserController
			controller.Init(ctx)
			controller.Index(page_, pagesize_)
			controller.Close()
			return ctx.JSON(controller.Result)
		})

		apiGroup.Post("/user", func(ctx *fiber.Ctx) error {
			item_ := &models.User{}
			ctx.BodyParser(item_)
			var controller rest.UserController
			controller.Init(ctx)
			hashed, err := HashPasswd(item_.Passwd)
			if item_.Passwd != "" && err == nil {
				item_.Passwd = hashed
				controller.Insert(item_)
			} else {
			    controller.Result["code"] = "error"
			}
			controller.Close()
			return ctx.JSON(controller.Result)
		})

		apiGroup.Put("/user", func(ctx *fiber.Ctx) error {
			item_ := &models.User{}
			ctx.BodyParser(item_)
			var controller rest.UserController
			controller.Init(ctx)
			// 본인 계정만 수정할 수 있다 (IDOR 방지).
			if controller.Session == nil || item_.Id != controller.Session.Id {
				controller.Close()
				return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": "error", "message": "forbidden"})
			}
			if item_.Passwd == "" {
				// 비밀번호를 보내지 않으면 기존 비밀번호를 유지한다.
				if existing := models.NewUserManager(controller.NewConnection()).Get(item_.Id); existing != nil {
					item_.Passwd = existing.Passwd
				}
			} else {
				hashed, err := HashPasswd(item_.Passwd)
				if err != nil {
					controller.Close()
					return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "error"})
				}
				item_.Passwd = hashed
			}
			if item_ != nil {
				controller.Update(item_)
			} else {
			    controller.Result["code"] = "error"
			}
			controller.Close()
			return ctx.JSON(controller.Result)
		})

		apiGroup.Delete("/user", func(ctx *fiber.Ctx) error {
			item_ := &models.User{}
			ctx.BodyParser(item_)
			var controller rest.UserController
			controller.Init(ctx)
			// 본인 계정만 삭제할 수 있다 (IDOR 방지).
			if controller.Session == nil || item_.Id != controller.Session.Id {
				controller.Close()
				return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": "error", "message": "forbidden"})
			}
			if item_ != nil {
				controller.Delete(item_)
			} else {
			    controller.Result["code"] = "error"
			}
			controller.Close()
			return ctx.JSON(controller.Result)
		})
	}
}