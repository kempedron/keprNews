package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"news/internal/database"
	"news/internal/handler"
	"news/internal/middleware"

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
	fmt.Println("hello from kepr news!")
	database.InitDB()
	e := echo.New()
	templates := template.Must(template.ParseGlob("web/templates/*.html"))
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}
			return next(c)
		}
	})
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
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

	e.GET("/", func(c echo.Context) error {
		return c.File("web/templates/index.html")
	})
	e.GET("/login-page", func(c echo.Context) error {
		return c.File("web/templates/loginpage.html")
	})
	e.GET("/register-page", func(c echo.Context) error {
		return c.File("web/templates/registerpage.html")
	})
	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	protected := e.Group("")
	protected.Use(middleware.JWTAuth)
	protected.GET("/home", handler.HomePage)
	protected.GET("/popular-news", handler.AllArticle)
	protected.GET("/add-article-page", handler.AddArticlePage)
	protected.POST("/add-article", handler.AddArticle)
	protected.GET("/article/:article_id", handler.GetArticle)
	protected.POST("/article/delete/:article_id", handler.DeleteArticle)
	protected.GET("/article/search", handler.SearchArticles)
	protected.GET("/search", func(c echo.Context) error {
		return c.File("web/templates/search.html")
	})
	e.Logger.Fatal(e.Start("127.0.0.1:8080"))

}
