package handler

import (
	"log"
	"net/http"
	"news/internal/database"
	"news/internal/models"

	"github.com/labstack/echo/v4"
)

type AuthRequest struct {
	Username string `form:"username" validate:"required, min=3"`
	Password string `form:"password" validate:"required, min=6"`
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
	// token, err := jwt.GenerateToken(user.ID, user.Username)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not hash password"})
	// }
	return c.Render(http.StatusPermanentRedirect, "homepage.html", user)
	// return c.JSON(http.StatusCreated, map[string]interface{}{
	// 	"message": "User created successfully",
	// 	"token":   token,
	// 	"user": map[string]interface{}{
	// 		"id":       user.ID,
	// 		"username": user.Username,
	// 	},
	// })
}

func Login(c echo.Context) error {
	log.Println("322")
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
	return c.Render(http.StatusOK, "homepage.html", user)
	// token, err := jwt.GenerateToken(user.ID, user.Username)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate token"})
	// }
	// return c.JSON(http.StatusOK, map[string]interface{}{
	// 	"message": "Login succesful",
	// 	"token":   token,
	// 	"user": map[string]interface{}{
	// 		"id":       user.ID,
	// 		"username": user.Username,
	// 	},
	// })
}
