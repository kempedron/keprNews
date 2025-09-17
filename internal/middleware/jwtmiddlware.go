package middleware

import (
	"net/http"
	"news/internal/jwt"
	"time"

	"github.com/labstack/echo/v4"
)

func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("jwt")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login-page")
		}
		tokenstring := cookie.Value
		if tokenstring == "" {
			return c.Redirect(http.StatusSeeOther, "/login-page")
		}
		claims, err := jwt.ValidateToken(tokenstring)
		if err != nil {
			cookie := new(http.Cookie)
			cookie.Name = "jwt"
			cookie.Value = ""
			cookie.Expires = time.Now().Add(24 * time.Hour)
			cookie.Path = "/"
			c.SetCookie(cookie)
			return c.Redirect(http.StatusSeeOther, "/login-page")
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		return next(c)
	}
}
