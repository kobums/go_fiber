package router

import (
	"errors"
	"log"
	"net/http"
	"project/config"
	"project/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthTokenClaims struct {
	User models.User `json:"user"`
	jwt.RegisteredClaims
}

func JwtAuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		if ((c.Method() == fiber.MethodPost) && c.Path() == "/api/user") || c.Path() == "/api/jwt" {
			return c.Next()
		}

		values := c.Get("Authorization")
		if values != "" {
			str := values

			if len(str) > 7 && str[:7] == "Bearer " {
				token = str[7:]

				claims := AuthTokenClaims{}
				key := func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("Unexpected Signing Method")
					}
					return []byte(config.SecretCode), nil
				}

				tok, err := jwt.ParseWithClaims(token, &claims, key)
				if err == nil {
					c.Locals("jwt", tok)
					c.Locals("user", &(claims.User))
					return c.Next()
				}
			} else {
				log.Println("Jwt header is broken")
			}
		} else {
			log.Println("Jwt header not found")
		}

		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"code":"error",
			"message":"not auth",
		})
	}
}

func JwtAuth(email string, passwd string) fiber.Map {
	conn := models.NewConnection()

	manager := models.NewUserManager(conn)
	user := manager.GetByEmail(email)

	if user == nil {
		return fiber.Map{
			"code":    "error",
			"message": "user not found",
		}
	}

	if user.Passwd != passwd {
		return fiber.Map{
			"code":    "error",
			"message": "wrong password",
		}
	}

	at := AuthTokenClaims{
		User: *user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 365 * 10)),
		},
	}

	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &at)
	signedAuthToken, _ := atoken.SignedString([]byte(config.SecretCode))

	user.Passwd = ""
	return fiber.Map{
		"code":  "ok",
		"token": signedAuthToken,
	}
}
