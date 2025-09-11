package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"news/internal/database"
	"news/internal/handler"

	echo "github.com/labstack/echo/v4"
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

	e.GET("/", func(c echo.Context) error {
		return c.File("web/templates/index.html")
	})
	e.GET("/login-page", func(c echo.Context) error {
		return c.File("web/templates/loginpage.html")
	})
	e.GET("/register-page", func(c echo.Context) error {
		return c.File("web/templates/registerpage.html")
	})
	e.GET("/home-page", func(c echo.Context) error {
		return c.File("web/templates/homepage.html")
	})
	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	err := e.Start("127.0.0.1:8080")
	if err != nil {
		log.Println(err)
	}

}
