package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"news/internal/database"
	"news/internal/handler"
	"news/internal/middleware"
	"news/internal/models"

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

	templates := template.Must(template.ParseGlob("templates/*.html"))
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
			return c.File("templates/index.html")
		}
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			return c.File("templates/index.html")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"IsAuthorized": true,
			"Username":     user.Username,
		})
	})
	e.GET("/", func(c echo.Context) error {
		return c.File("templates/index.html")
	})
	e.GET("/login-page", func(c echo.Context) error {
		return c.File("templates/loginpage.html")
	})
	e.GET("/register-page", func(c echo.Context) error {
		return c.File("templates/registerpage.html")
	})
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "все работает")
	})
	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	e.POST("/logout", handler.Logout)
	protected := e.Group("")
	protected.Use(middleware.JWTAuth)
	protected.Use(middleware.ReqPerSecLimitMiddleware(5))
	protected.GET("/popular-news", handler.AllArticle)
	protected.GET("/add-article-page", func(c echo.Context) error {
		return c.File("templates/addArticle.html")
	})
	protected.POST("/add-article", handler.AddArticle)
	protected.GET("/article/:article_id", handler.GetArticle)
	protected.POST("/article/delete/:article_id", handler.DeleteArticle)
	protected.GET("/article/search", handler.SearchArticles)
	protected.GET("/search", func(c echo.Context) error {
		return c.File("templates/search.html")
	})
	e.Logger.Fatal(e.Start("0.0.0.0:8080"))

}
