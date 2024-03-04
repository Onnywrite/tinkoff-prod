package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Authorized() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header["Authorization"]
			if len(auth) == 0 {
				return c.JSON(http.StatusUnauthorized, &crush{
					Reason: "unauthorized user",
				})
			}

			bearerToken := strings.Split(auth[0], " ")
			if bearerToken[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, &crush{
					Reason: "invalid authorization header format",
				})
			}
			tokenString := bearerToken[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}

				return []byte("$my_%SUPER(n0t-so=MUch)_secret123"), nil
			})

			if err != nil {
				c.JSON(http.StatusUnauthorized, &crush{
					Reason: "unexpected signing method on token",
				})
				return err
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				exp, ok2 := claims["exp"].(int64)
				if !ok2 {
					c.JSON(http.StatusUnauthorized, &crush{
						Reason: "token has expired",
					})
					return err
				}

				if time.Now().Unix() > exp {
					c.JSON(http.StatusUnauthorized, &crush{
						Reason: "token has expired",
					})
					return err
				}

				c.Set("login", claims["login"])
				c.Set("email", claims["email"])
				c.Set("phone", claims["phone"])
			} else {
				c.JSON(http.StatusUnauthorized, &crush{
					Reason: "invalid token",
				})
				return err
			}

			return next(c)
		}
	}
}
