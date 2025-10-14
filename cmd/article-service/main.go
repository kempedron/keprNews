package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	articleHandler "news/internal/article/handler"
	"news/pkg/database"
	"os"

	"news/pkg/middleware"

	"github.com/labstack/echo-contrib/echoprometheus"
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

	e.Use(echoprometheus.NewMiddleware("article_service"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})
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
	protected := e.Group("")
	e.Use(middleware.CheckTimeForResp)
	protected.Use(middleware.JWTAuth)
	e.GET("/popular-news", articleHandler.AllArticle)
	protected.GET("/add-article-page", func(c echo.Context) error {
		return c.File("/root/web/templates/addArticle.html")
	})
	e.POST("/add-article", articleHandler.AddArticle)
	protected.GET("/article/:article_id", articleHandler.GetArticle)
	protected.POST("/article/delete/:article_id", articleHandler.DeleteArticle)
	protected.GET("/article/search", articleHandler.SearchArticles)
	protected.GET("/search", func(c echo.Context) error {
		return c.File("/root/web/templates/search.html")
	})
	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(":8081"); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))

}
