package main

import (
	"log"
	"net/http"
	"net/url"

	myMiddleware "news/pkg/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	proxy  middleware.ProxyBalancer
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
			log.Fatal("Failed to parse %s service URL: %v", name, err)
		}
		g.services[name] = &ServiceProxy{
			target: target,
			proxy: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{URL: target},
			}),
		}
	}
}

func (g *APIGateway) setMiddleware() {
	g.echo.Use(middleware.Logger())
	g.echo.Use(middleware.Recover())
	g.echo.Use(middleware.CORS())
	g.echo.Use(middleware.Gzip())

	g.echo.Use(myMiddleware.ReqPerSecLimitMiddleware(5))
}

func (g *APIGateway) setRoutes() {
	protected := g.echo.Group("/api")
	protected.Use(myMiddleware.JWTAuth)

	protected.POST("/articles", g.proxyToArticleService)
	protected.PUT("/articles/:id", g.proxyToArticleService)
	protected.DELETE("/articles/:id", g.proxyToArticleService)

}
func (g *APIGateway) proxyToArticleService(c echo.Context) error {
	return g.proxyToService("article")(c)
}

func (g *APIGateway) proxyToAuthService(c echo.Context) error {
	return g.proxyToService("auth")(c)
}

// Общий метод для проксирования
func (g *APIGateway) proxyToService(serviceName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		service, exists := g.services[serviceName]
		if !exists {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Service unavailable",
			})
		}

		// Используем echoMiddleware
		proxy := echoMiddleware.ProxyWithConfig(echoMiddleware.ProxyConfig{
			Balancer: service.proxy,
		})

		return proxy(c)
	}
}

func main() {
	cfg := &Config{
		Port:              config.GetEnv("PORT", "8080"),
		AuthServiceURL:    config.GetEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		ArticleServiceURL: config.GetEnv("ARTICLE_SERVICE_URL", "http://article-service:8082"),
		RedisURL:          config.GetEnv("REDIS_URL", "redis://redis:6379"),
		JWTSecret:         config.GetEnv("JWT_SECRET", ""),
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	gateway := NewAPIGateway(cfg)

}

func NewAPIGateway(cfg *Config) *APIGateway {
	e := echo.New()
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
	g
}
