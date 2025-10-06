package middleware

// import (
// 	"net/http"

// 	"github.com/labstack/echo/v4"
// 	"golang.org/x/time/rate"
// )

// func ReqPerSecLimitMiddleware(requestPerSecond int) echo.MiddlewareFunc {
// 	limiter := rate.NewLimiter(rate.Limit(requestPerSecond), 1)

// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			if !limiter.Allow() {
// 				return c.JSON(http.StatusTooManyRequests, map[string]string{
// 					"Error": "Too many requests",
// 				})
// 			}
// 			return next(c)
// 		}
// 	}
// }
