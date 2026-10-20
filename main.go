package main

import (
	"project/router"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	log "github.com/sirupsen/logrus"
)

func main() {
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: false,
	// 	FullTimestamp: true,
	// })

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		// ${url} 은 쿼리스트링(/api/jwt?passwd=...)까지 기록하므로 ${path} 만 남긴다.
		Format: "[${time}] | ${status} | ${latency} | ${ip}:${port} | ${method} | ${path}\n",
		TimeFormat: time.DateTime,
	}))

	app.Use(cors.New(cors.Config{
		AllowHeaders: "Origin, Content-Type, Authorization, Accept",
		AllowCredentials: true,
		AllowOrigins: "http://141.164.56.77:3000, http://localhost:3000",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			// fiber.MethodHead,
			// fiber.MethodPatch,
		}, ","),
	}))

	// POST /api/jwt (로그인). GET /api/jwt?email=&passwd= 는 router.go 에 deprecated 로 남아 있다.
	router.SetAuthRouter(app)
	router.SetRouter(app)

	log.Fatal(app.Listen(":9003"))
}
