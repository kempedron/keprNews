package main

import (
	"log"
	"net/http"
	"net/url"

	"news/pkg/config"
	myMiddleware "news/pkg/middleware"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port              string
	AuthServiceURL    string
	ArticleServiceURL string
	RedisURL          string
	JWTSecret         string
}

type ServiceProxy struct {
	target *url.URL
	proxy  echoMiddleware.ProxyBalancer
}

type APIGateway struct {
	config   *Config
	echo     *echo.Echo
	redis    *redis.Client
	services map[string]*ServiceProxy
}

func (g *APIGateway) initService() {
	services := map[string]string{
		"auth":    g.config.AuthServiceURL,
		"article": g.config.ArticleServiceURL,
	}
	for name, serviceURL := range services {
		target, err := url.Parse(serviceURL)
		if err != nil {
			log.Fatalf("Failed to parse %s service URL: %v", name, err)
		}
		g.services[name] = &ServiceProxy{
			target: target,
			proxy: echoMiddleware.NewRoundRobinBalancer([]*echoMiddleware.ProxyTarget{
				{URL: target},
			}),
		}
	}
}

func (g *APIGateway) setMiddleware() {
	g.echo.Use(echoMiddleware.Logger())
	g.echo.Use(echoMiddleware.Recover())
	g.echo.Use(echoMiddleware.CORS())
	g.echo.Use(echoMiddleware.Gzip())

}

func (g *APIGateway) setRoutes() {
	// Статические страницы
	g.echo.GET("/", g.proxyToAuthService)

	// Группа страниц аутентификации
	authPages := g.echo.Group("")
	authPages.GET("/login-page", g.proxyToAuthService)
	authPages.GET("/register-page", g.proxyToAuthService)

	// Группа страниц статей
	articlePages := g.echo.Group("")
	articlePages.Use(myMiddleware.JWTAuth)
	articlePages.GET("/add-article-page", g.proxyToArticleService)
	articlePages.GET("/search", g.proxyToArticleService)
	articlePages.GET("/article/:article_id", g.proxyToArticleService)

	// Public API routes
	public := g.echo.Group("")
	public.POST("/login", g.proxyToAuthService)
	public.POST("/register", g.proxyToAuthService)
	public.POST("/logout", g.proxyToAuthService)
	public.GET("/get-info/user-info", g.proxyToAuthService)
	public.GET("/popular-news", g.proxyToArticleService)

	// Protected API routes
	protected := g.echo.Group("")
	protected.Use(myMiddleware.JWTAuth)
	protected.POST("/add-article", g.proxyToArticleService)
	protected.POST("/article/delete/:article_id", g.proxyToArticleService)
	protected.POST("/articles", g.proxyToArticleService)
	protected.PUT("/articles/:id", g.proxyToArticleService)
	protected.GET("/popular-news", g.proxyToArticleService)
	protected.GET("/article/search", g.proxyToArticleService)
}

func (g *APIGateway) proxyToArticleService(c echo.Context) error {
	return g.proxyToService("article")(c)
}

func (g *APIGateway) proxyToAuthService(c echo.Context) error {
	return g.proxyToService("auth")(c)
}

func (g *APIGateway) proxyToService(serviceName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Printf("Проксирование запроса к сервису: %s", serviceName) // Добавьте это
		service, exists := g.services[serviceName]
		if !exists {
			log.Printf("Сервис %s не найден", serviceName) // И это
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Service unavailable",
			})
		}

		proxyMiddleware := echoMiddleware.ProxyWithConfig(echoMiddleware.ProxyConfig{
			Balancer: service.proxy,
		})

		handler := proxyMiddleware(func(c echo.Context) error { return nil })
		return handler(c)
	}
}

func main() {
	cfg := &Config{
		Port:              config.GetEnv("PORT", "8080"),
		AuthServiceURL:    config.GetEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		ArticleServiceURL: config.GetEnv("ARTICLE_SERVICE_URL", "http://article-service:8080"),
		RedisURL:          config.GetEnv("REDIS_URL", "redis:6379"),
		JWTSecret:         config.GetEnv("JWT_SECRET", ""),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	gateway := NewAPIGateway(cfg)

	defer gateway.echo.Close()
	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(":8081"); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	log.Printf("API Gateway start on port %s", cfg.Port)
	if err := gateway.echo.Start(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start API Gateway:", err)
	}

}

func NewAPIGateway(cfg *Config) *APIGateway {
	e := echo.New()

	e.Use((echoprometheus.NewMiddleware("api_gateway"))) // Собирает метрики
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to parse redis url:", err)
	}
	redisClient := redis.NewClient(redisOpts)
	gateway := &APIGateway{
		config:   cfg,
		echo:     e,
		redis:    redisClient,
		services: make(map[string]*ServiceProxy),
	}
	gateway.initService()
	gateway.setRoutes()
	gateway.setMiddleware()
	return gateway
}
