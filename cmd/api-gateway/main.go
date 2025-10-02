package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Проксирование запросов к сервисам
	e.POST("/api/register", proxyToAuthService)
	e.POST("/api/login", proxyToAuthService)
	e.GET("/api/articles", proxyToArticleService)
	e.POST("/api/articles", proxyToArticleService, authMiddleware) // Требуется аутентификация
	e.DELETE("/api/articles/:id", proxyToArticleService, authMiddleware)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	})

	log.Println("API Gateway запущен на :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

// Middleware аутентификации
func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		// Валидация JWT токена
		userID, err := validateToken(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		// Передача userID внутренним сервисам
		c.Set("userID", userID)
		return next(c)
	}
}

// Функция прокси (реализация на основе httputil.ReverseProxy или HTTP-клиента)
func proxyToAuthService(c echo.Context) error {
	// Логика перенаправления запроса в auth-service
	return c.String(http.StatusOK, "Proxied to Auth Service")
}
