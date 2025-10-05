package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	authHandler "news/internal/auth/handler"
	"news/pkg/database"
	"os"

	"news/pkg/middleware"
	"news/pkg/models"

	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type TemplateRegistry struct {
	templates *template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	err := database.InitRedis()
	if err != nil {
		log.Printf("error init redis: %s", err)
		log.Fatal(err)
	}
	err = database.InitDB()
	if err != nil {
		log.Printf("error init database: %s", err)
		log.Fatal(err)
	}
	e := echo.New()

	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "/root/web/templates/"
	}

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Fatalf("Директория с шаблонами не найдена: %s", templatePath)
	}

	templatePath = "./web/templates/"
	templates := template.Must(template.ParseGlob(templatePath + "*.html"))
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.Use(echomiddleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}
			return next(c)
		}
	})
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://0.0.0.0:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
			echo.HeaderXRequestedWith,
		},
		ExposeHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentLength,
			"X-Total-Count",
		},
		MaxAge: 86400,
	}))

	e.GET("/get-info/user-info", func(c echo.Context) error {
		userID, err := middleware.GetUserIDFromToken(c)
		if err != nil {
			return c.File("web/templates/index.html")
		}
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			return c.File("/root/web/templates/index.html")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"IsAuthorized": true,
			"Username":     user.Username,
		})
	})

	e.GET("/", func(c echo.Context) error {
		return c.File("/root/web/templates/index.html")
	})
	e.GET("/login-page", func(c echo.Context) error {
		return c.File("/root/web/templates/loginpage.html")
	})
	e.GET("/register-page", func(c echo.Context) error {
		return c.File("/root/web/templates/registerpage.html")
	})

	e.POST("/login", authHandler.Login)
	e.POST("/register", authHandler.Register)
	e.POST("/logout", authHandler.Logout)
	protected := e.Group("")
	protected.Use(middleware.JWTAuth)
	protected.Use(middleware.ReqPerSecLimitMiddleware(5))
	e.Logger.Fatal(e.Start("0.0.0.0:8080"))

}
