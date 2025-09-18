package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"news/internal/database"
	"news/internal/jwt"
	"news/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
	return c.Redirect(http.StatusSeeOther, "/home")
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
	return c.Redirect(http.StatusSeeOther, "/home")
}
func AddArticlePage(c echo.Context) error {
	return c.File("web/templates/addArticle.html")
}

func AddArticle(c echo.Context) error {
	type ArticleRequest struct {
		Title   string `json:"article-title"`
		Content string `json:"article-content"`
		Tags    string `json:"tags"`
	}

	var req ArticleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	inputTags := strings.TrimSpace(req.Tags)

	if title == "" || content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Заголовок и содержание статьи обязательны",
		})
	}
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user ID from token"})
	}
	tx := database.DB.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка начала транзакции"})
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	article := models.Article{
		AuthorID:       userID,
		ArticleTitle:   title,
		ArticleContent: content,
	}
	if err := tx.Create(&article).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "не удалось создать статью:" + err.Error()})
	}
	if inputTags != "" {
		tagNames := strings.Split(inputTags, ",")
		uniqueTags := make(map[string]bool)
		var tagsToProcess []string

		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName != "" && !uniqueTags[tagName] {
				uniqueTags[tagName] = true
				tagsToProcess = append(tagsToProcess, tagName)
			}
		}
		if len(tagsToProcess) > 0 {
			var existTags []models.Tag
			if err := tx.Where("tag_content IN (?)", tagsToProcess).Find(&existTags).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при поиске тегов"})
			}
			existingTagMap := make(map[string]models.Tag)
			for _, tag := range existTags {
				existingTagMap[tag.TagContent] = tag
			}
			var newTags []models.Tag
			for _, tagname := range tagsToProcess {
				if _, exists := existingTagMap[tagname]; !exists {
					newTags = append(newTags, models.Tag{TagContent: tagname})
				}
			}
			if len(newTags) > 0 {
				if err := tx.Create(&newTags).Error; err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при создании тегов"})
				}
				for _, tag := range newTags {
					existingTagMap[tag.TagContent] = tag
				}
			}

			var articleTags []models.Tag
			for _, tagname := range tagsToProcess {
				articleTags = append(articleTags, existingTagMap[tagname])
			}
			if err := tx.Model(&article).Association("Tags").Append(articleTags); err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при связывании тега со статьей"})
			}
		}

	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при коммите транзакции"})
	}

	if err := database.DB.Preload("Tags").First(&article, article.ID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при загрузке статьи"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "статья успешно создана",
		"article": article,
	})
}
func AllArticle(c echo.Context) error {
	articles, err := GetArticlesWithDetails(database.DB)
	if err != nil {
		log.Printf("error get articles from DB: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error in get articles from DB"})
	}
	return c.Render(http.StatusOK, "allArticle.html", articles)
}

func GetArticlesWithDetails(db *gorm.DB) ([]models.Article, error) {
	var articles []models.Article
	err := db.Preload("Author").Preload("Tags").Order("RANDOM()").Limit(10).Find(&articles).Error
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func HomePage(c echo.Context) error {
	userIDValue := c.Get("userID")
	if userIDValue == nil {
		return c.Redirect(http.StatusSeeOther, "/login-page")
	}
	userID := userIDValue.(uint)
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "user not found"})
	}
	return c.Render(http.StatusOK, "homepage.html", user)

}
func GetArticle(c echo.Context) error {
	articleID := c.Param("article_id")
	articleIDUint, err := strconv.ParseUint(articleID, 10, 32)
	if err != nil {
		log.Printf("error parse articleId -> uint: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат ID статьи"})
	}
	article, err := GetArticleByID(database.DB, articleIDUint)
	if err != nil {
		log.Printf("error in getting article by ID: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ошибка на стороне сервера"})
	}
	return c.Render(http.StatusOK, "article.html", article)
}

func GetArticleByID(db *gorm.DB, articleID uint64) (models.Article, error) {
	var article models.Article
	err := db.Preload("Author").Preload("Tags").First(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Article{}, fmt.Errorf("статья с ID %d не найдена ", articleID)
		}
		return models.Article{}, fmt.Errorf("ошибка при получении статьи: %s", err)
	}
	return article, nil
}
