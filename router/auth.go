package router

import (
	"errors"
	"log"
	"net/http"
	"project/config"
	"strings"
	"project/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthTokenClaims struct {
	User models.User `json:"user"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserId int64 `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// SetAuthRouter 는 POST /api/jwt 로그인 라우트를 등록한다.
// router.go 는 buildtool-router 가 재생성하며 덮어쓰므로, 재생성 후에도 유지되도록
// 이 파일(수동 관리)에 두고 main.go 에서 SetRouter 보다 먼저 호출한다.
// (JwtAuthRequired 미들웨어보다 먼저 등록되어야 인증 없이 호출된다.)
func SetAuthRouter(app *fiber.App) {
	app.Post("/api/jwt", func(ctx *fiber.Ctx) error {
		var req struct {
			Email  string `json:"email" form:"email"`
			Passwd string `json:"passwd" form:"passwd"`
		}
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "error", "message": "invalid request"})
		}
		return ctx.JSON(JwtAuth(req.Email, req.Passwd))
	})
}

func JwtAuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		if ((c.Method() == fiber.MethodPost) && c.Path() == "/api/user") {
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

	if !verifyPassword(manager, user, passwd) {
		return fiber.Map{
			"code":    "error",
			"message": "wrong password",
		}
	}

	// 토큰(클레임)에 비밀번호가 들어가지 않도록 비운다.
	user.Passwd = ""

	at := AuthTokenClaims{
		User: *user,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 6)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
	}

	authManager := models.NewAuthManager(conn)
	auth := authManager.GetByUser(user.Id)

	rt := RefreshTokenClaims{
		UserId: user.Id,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
		},
	}

	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &at)
	signedAuthToken, _ := atoken.SignedString([]byte(config.SecretCode))

	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &rt)
	signedRefreshToken, _ := rtoken.SignedString([]byte(config.SecretCode))

	refreshTokenItem := models.Auth{
		User: user.Id,
		Token: signedRefreshToken,
	}

	if auth == nil {
		authManager.Insert(&refreshTokenItem)
	} else {
		refreshTokenItem.Id = auth.Id
		authManager.Update(&refreshTokenItem)
	}

	user.Passwd = ""
	return fiber.Map{
		"code":  "ok",
		"accessToken": signedAuthToken,
		"refreshToken": signedRefreshToken,
	}
}

func JwtToken(refreshToken string) fiber.Map {
		values := refreshToken
		if values != "" {
			str := values

			if len(str) > 7 && str[:7] == "Bearer " {
				refreshToken = str[7:]

				claims := RefreshTokenClaims{}
				key := func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("Unexpected Signing Method")
					}
					return []byte(config.SecretCode), nil
				}

				_, err := jwt.ParseWithClaims(refreshToken, &claims, key)
				if err == nil {
					conn := models.NewConnection()

					manager := models.NewUserManager(conn)
					user := manager.GetByEmail((claims.Email))

					authManager := models.NewAuthManager((conn))
					auth := authManager.GetByUser((claims.UserId))

					if auth == nil {
						return fiber.Map{
							"code":    "error",
							"message": "token not found",
						}
					}

					if auth.Token != refreshToken {
						return nil
					}

					if user == nil {
						return fiber.Map{
							"code":    "error",
							"message": "user not found",
						}
					}

					// 토큰(클레임)에 비밀번호가 들어가지 않도록 비운다.
					user.Passwd = ""
					at := AuthTokenClaims{
						User: *user,
						RegisteredClaims: jwt.RegisteredClaims{
							// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 6)),
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
						},
					}

					atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &at)
					signedAuthToken, _ := atoken.SignedString([]byte(config.SecretCode))
					

					user.Passwd = ""
					return fiber.Map{
						"code":  "ok",
						"accessToken": signedAuthToken,
					}

				}
			} else {
				log.Println("Jwt header is broken")
			}
		} else {
			log.Println("Jwt header not found")
		}

		return fiber.Map{
			"code":"error",
			"message":"not auth",
		}
}


func JwtMe(token string) fiber.Map {
		values := token
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

				_, err := jwt.ParseWithClaims(token, &claims, key)
				if err == nil {
					return fiber.Map{
						"code":    "ok",
						"id": claims.User.Name,
						"username": claims.User.Email,
						"imageUrl": "/logo/codefactory_logo.png",
					}
				}
			} else {
				log.Println("Jwt header is broken")
			}
		} else {
			log.Println("Jwt header not found")
		}

		return fiber.Map{
			"code":"error",
			"message":"not auth",
		}
}

// HashPasswd 는 비밀번호를 bcrypt 로 해시한다.
func HashPasswd(passwd string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// verifyPassword 는 저장된 값이 bcrypt 해시면 bcrypt 로 검증한다.
// 기존에 평문으로 저장된 계정은 평문 비교 후 일치하면 bcrypt 로 자동 마이그레이션한다.
func verifyPassword(manager *models.UserManager, user *models.User, raw string) bool {
	stored := user.Passwd
	if strings.HasPrefix(stored, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(raw)) == nil
	}
	if stored != "" && stored == raw {
		if hashed, err := HashPasswd(raw); err == nil {
			user.Passwd = hashed
			manager.Update(user)
		}
		return true
	}
	return false
}
