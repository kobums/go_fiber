// 경고: 이 파일과 controllers/rest/*.go, models/*.go 는 생성 도구(buildtool-router /
// buildtool-model)로 만든 코드에 보안 수정을 직접 덧댄 것이다. 다시 생성하면 아래 수정이
// 전부 지워지므로, 재생성 후 반드시 다시 적용하고 확인할 것.
//
//  1. 비밀번호 해시: POST /api/user 는 HashPasswd(bcrypt)로 저장, PUT /api/user 는 비밀번호가
//     비어 있으면 기존 해시 유지·있으면 HashPasswd. 평문 저장 금지.
//  2. 본인 계정 확인(IDOR 방지): PUT / DELETE /api/user 는
//     controller.Session == nil || item_.Id != controller.Session.Id 이면 403.
//  3. 로그인: POST /api/jwt(router/auth.go SetAuthRouter)가 정식. 아래 GET /api/jwt 쿼리스트링
//     로그인은 기존 클라이언트 호환용 deprecated — 쿼리스트링을 로그에 남기지 말 것.
//  4. 응답/JWT 에서 비밀번호 제거: router/auth.go 의 user.Passwd = "",
//     controllers/rest/user.go Index/Read 의 Passwd = "", passwd 쿼리 필터 제거.
//  5. SQL: models/db.go 의 WhereClause(컬럼/연산자 검증, in 값 바인딩)·Ordering 검증을
//     models/user.go, models/auth.go 의 case Where 가 호출해야 한다.
//  6. 설정: config.json 은 git 에서 제외(config.json.example 참고), 비밀값은 환경변수로.
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