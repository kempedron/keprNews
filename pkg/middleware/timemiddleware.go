package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

func CheckTimeForResp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		log.Printf("Request %s %s completed in %v", c.Request().Method, c.Request().URL.Path, time.Since(start))
		return err
	}
}
