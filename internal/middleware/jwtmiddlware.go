package middleware

import (
	"net/http"
	"news/internal/jwt"
	"strings"

	"github.com/labstack/echo/v4"
)

func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorizetion")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authtorization header required"})
		}
		tokenstring := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenstring == authHeader {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
		}
		claims, err := jwt.ValidateToken(tokenstring)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		return next(c)
	}
}
