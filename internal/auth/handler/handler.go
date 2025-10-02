package handler

import (
	"log"
	"net/http"
	"news/pkg/database"
	"news/pkg/jwt"
	"news/pkg/models"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

type AuthRequest struct {
	Username string `form:"username" validate:"required, min=3"`
	Password string `form:"password" validate:"required, min=6"`
}

func Login(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("error in getbind: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	var user models.User
	log.Printf("req.Username:%s", req.Username)
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		log.Printf("error in usercheck: %s", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
	if err := user.CheckPassword(req.Password); err != nil {
		log.Printf("error in passcheck: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid autenthecation"})
	}
	log.Println("username:", user.Username)

	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("error generating token: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not generate token"})
	}
	cookie := new(http.Cookie)
	cookie.Name = "jwt"
	cookie.Value = token
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/")
}

func Register(c echo.Context) error {
	var req AuthRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Username aldery exist"})
	}
	user := models.User{
		Username: req.Username,
	}
	if err := user.HashPassword(req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not hash password"})
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("error generating token: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not generate token"})
	}
	cookie := new(http.Cookie)
	cookie.Name = "jwt"
	cookie.Value = token
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/")
}

func Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "jwt"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
