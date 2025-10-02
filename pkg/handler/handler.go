// package handler

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"math/rand"
// 	"net/http"
// 	"news/pkg/database"
// 	"news/pkg/models"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/labstack/echo/v4"
// 	"github.com/redis/go-redis/v9"
// 	"gorm.io/gorm"
// )

// var redisClient *redis.Client

// type AuthRequest struct {
// 	Username string `form:"username" validate:"required, min=3"`
// 	Password string `form:"password" validate:"required, min=6"`
// }

// func AddArticle(c echo.Context) error {
// 	type ArticleRequest struct {
// 		Title   string `json:"article-title"`
// 		Content string `json:"article-content"`
// 		Tags    string `json:"tags"`
// 	}

// 	var req ArticleRequest
// 	if err := c.Bind(&req); err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
// 	}

// 	title := strings.TrimSpace(req.Title)
// 	content := strings.TrimSpace(req.Content)
// 	inputTags := strings.TrimSpace(req.Tags)

// 	if title == "" || content == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "Заголовок и содержание статьи обязательны",
// 		})
// 	}
// 	userID, ok := c.Get("userID").(uint)
// 	if !ok {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user ID from token"})
// 	}
// 	tx := database.DB.Begin()
// 	if tx.Error != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка начала транзакции"})
// 	}

// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	article := models.Article{
// 		AuthorID:       userID,
// 		ArticleTitle:   title,
// 		ArticleContent: content,
// 	}
// 	if err := tx.Create(&article).Error; err != nil {
// 		tx.Rollback()
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "не удалось создать статью:" + err.Error()})
// 	}
// 	if inputTags != "" {
// 		tagNames := strings.Split(inputTags, ",")
// 		uniqueTags := make(map[string]bool)
// 		var tagsToProcess []string

// 		for _, tagName := range tagNames {
// 			tagName = strings.TrimSpace(tagName)
// 			if tagName != "" && !uniqueTags[tagName] {
// 				uniqueTags[tagName] = true
// 				tagsToProcess = append(tagsToProcess, tagName)
// 			}
// 		}
// 		if len(tagsToProcess) > 0 {
// 			var existTags []models.Tag
// 			if err := tx.Where("tag_content IN (?)", tagsToProcess).Find(&existTags).Error; err != nil {
// 				tx.Rollback()
// 				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при поиске тегов"})
// 			}
// 			existingTagMap := make(map[string]models.Tag)
// 			for _, tag := range existTags {
// 				existingTagMap[tag.TagContent] = tag
// 			}
// 			var newTags []models.Tag
// 			for _, tagname := range tagsToProcess {
// 				if _, exists := existingTagMap[tagname]; !exists {
// 					newTags = append(newTags, models.Tag{TagContent: tagname})
// 				}
// 			}
// 			if len(newTags) > 0 {
// 				if err := tx.Create(&newTags).Error; err != nil {
// 					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при создании тегов"})
// 				}
// 				for _, tag := range newTags {
// 					existingTagMap[tag.TagContent] = tag
// 				}
// 			}

// 			var articleTags []models.Tag
// 			for _, tagname := range tagsToProcess {
// 				articleTags = append(articleTags, existingTagMap[tagname])
// 			}
// 			if err := tx.Model(&article).Association("Tags").Append(articleTags); err != nil {
// 				tx.Rollback()
// 				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при связывании тега со статьей"})
// 			}
// 		}

// 	}
// 	if err := tx.Commit().Error; err != nil {
// 		tx.Rollback()
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при коммите транзакции"})
// 	}

// 	if err := database.DB.Preload("Tags").First(&article, article.ID).Error; err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при загрузке статьи"})
// 	}

// 	return c.JSON(http.StatusOK, map[string]interface{}{
// 		"message": "статья успешно создана",
// 		"article": article,
// 	})
// }

// func AllArticle(c echo.Context) error {
// 	articles, err := GetArticlesWithDetails(database.DB)
// 	if err != nil {
// 		log.Printf("error get articles from DB: %s", err)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error in get articles from DB"})
// 	}
// 	userID := c.Get("userID") // Ваша функция для получения ID пользователя
// 	var currentUser models.User
// 	if userID != 0 {
// 		database.DB.First(&currentUser, userID)
// 	}
// 	currentUsername := currentUser.Username
// 	return c.Render(http.StatusOK, "allArticle.html", map[string]interface{}{
// 		"articles":        articles,
// 		"currentUsername": currentUsername,
// 	})
// }

// func GetArticlesWithDetails(db *gorm.DB) ([]models.Article, error) {
// 	var articles []models.Article
// 	err := db.Preload("Author").Preload("Tags").Order("RANDOM()").Limit(10).Find(&articles).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return articles, nil
// }

// func HomePage(c echo.Context) error {
// 	userIDValue := c.Get("userID")
// 	if userIDValue == nil {
// 		return c.Redirect(http.StatusSeeOther, "/login-page")
// 	}
// 	userID := userIDValue.(uint)
// 	var user models.User
// 	if err := database.DB.First(&user, userID).Error; err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "user not found"})
// 	}
// 	return c.Render(http.StatusOK, "index.html", user)

// }
// func GetArticle(c echo.Context) error {
// 	articleID := c.Param("article_id")
// 	articleIDUint, err := strconv.ParseUint(articleID, 10, 32)
// 	if err != nil {
// 		log.Printf("error parse articleID -> uint: %s", err)
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат ID статьи"})
// 	}
// 	cacheKey := "article:" + articleID
// 	cachedData, err := redisClient.Get(c.Request().Context(), cacheKey).Result()
// 	if err == nil {
// 		// Кеш найден, возвращаем данные
// 		var article models.Article
// 		json.Unmarshal([]byte(cachedData), &article)
// 		return c.JSON(http.StatusOK, article)
// 	}

// 	article, err := GetArticleByIDFromDB(database.DB, articleIDUint)
// 	if err != nil {
// 		log.Printf("error in getting article by ID: %s", err)
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ошибка на стороне сервера"})
// 	}

// 	go func(art models.Article) {
// 		serialized, err := json.Marshal(art)
// 		if err != nil {
// 			log.Printf("failed to marshal article for cache: %v", err)
// 			return
// 		}
// 		// Устанавливаем TTL 5 минут с небольшим случайным отклонением (jitter) для защиты от одновременного протухания многих ключей :cite[1]
// 		ttl := 5*time.Minute + time.Duration(rand.Intn(30))*time.Second
// 		if err := redisClient.Set(context.Background(), cacheKey, serialized, ttl).Err(); err != nil {
// 			log.Printf("failed to set cache: %v", err)
// 		}
// 	}(article)

// 	return c.Render(http.StatusOK, "article.html", article)
// }

// func GetArticleByIDFromDB(db *gorm.DB, articleID uint64) (models.Article, error) {
// 	var article models.Article
// 	err := db.Preload("Author").Preload("Tags").First(&article, articleID).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return models.Article{}, fmt.Errorf("статья с ID %d не найдена ", articleID)
// 		}
// 		return models.Article{}, fmt.Errorf("ошибка при получении статьи: %s", err)
// 	}
// 	return article, nil
// }

// func DeleteArticle(c echo.Context) error {
// 	articleID := c.Param("article_id")
// 	referer := c.Request().Referer()
// 	articleUint, err := strconv.ParseUint(articleID, 10, 32)
// 	if err != nil {
// 		log.Printf("errror parse articleID -> uint: %s", err)
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат ID статьи"})
// 	}
// 	userID, ok := c.Get("userID").(uint)
// 	if !ok {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
// 	}
// 	var article models.Article
// 	result := database.DB.Preload("Author").First(&article, articleID)
// 	if result.Error != nil {
// 		return c.JSON(http.StatusNotFound, map[string]string{"error": "статья не найдена"})
// 	}
// 	if article.AuthorID != userID {
// 		return c.JSON(http.StatusForbidden, map[string]string{"error": "у вас нет прав для удаления данной записи"})
// 	}
// 	err = DeleteArticleByID(database.DB, articleUint)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка при удалении статьи"})
// 	}
// 	return c.Redirect(http.StatusFound, referer)
// }

// func DeleteArticleByID(db *gorm.DB, articleID uint64) error {
// 	err := db.Where("id = ?", articleID).Delete(&models.Article{}).Error
// 	if err != nil {
// 		log.Printf("error delete article with id %d:%s", articleID, err)
// 		return err
// 	}
// 	return nil
// }

// func SearchArticles(c echo.Context) error {
// 	searchQuery := c.FormValue("search-query")
// 	if searchQuery == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "Пустой поисковый запрос",
// 		})
// 	}

// 	page, _ := strconv.Atoi(c.QueryParam("page"))
// 	if page < 1 {
// 		page = 1
// 	}
// 	limit := 20
// 	offset := (page - 1) * limit

// 	var articles []models.Article

// 	// Безопасный поиск с использованием полнотекстовых возможностей PostgreSQL
// 	query := database.DB.Preload("Author").Preload("Tags").
// 		Where("articles.deleted_at IS NULL")

// 	if searchQuery != "" {
// 		// Используем phraseto_tsquery для поиска точной фразы
// 		query = query.Where(`
//             search_vector @@ phraseto_tsquery('russian', ?) OR
//             articles.id IN (
//                 SELECT at.article_id FROM article_tags at
//                 JOIN tags t ON t.id = at.tag_id
//                 WHERE to_tsvector('russian', t.tag_content) @@ phraseto_tsquery('russian', ?)
//             )
//         `, searchQuery, searchQuery).
// 			Select("*, ts_rank(search_vector, phraseto_tsquery('russian', ?)) as rank", searchQuery).
// 			Order("rank DESC, created_at DESC")
// 	}

// 	// Получаем результаты с пагинацией
// 	result := query.Offset(offset).Limit(limit).Find(&articles)

// 	if result.Error != nil {
// 		log.Printf("Ошибка при поиске статей: %v", result.Error)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": "Ошибка при поиске статей",
// 		})
// 	}
// 	userID := c.Get("userID") // Ваша функция для получения ID пользователя
// 	var currentUser models.User
// 	if userID != 0 {
// 		database.DB.First(&currentUser, userID)
// 	}
// 	currentUsername := currentUser.Username

// 	return c.Render(http.StatusOK, "allArticle.html", map[string]interface{}{
// 		"articles":        articles,
// 		"currentUsername": currentUsername,
// 	})
// }

// func Logout(c echo.Context) error {
// 	cookie := new(http.Cookie)
// 	cookie.Name = "jwt"
// 	cookie.Value = ""
// 	cookie.Expires = time.Now().Add(-time.Hour)
// 	cookie.Path = "/"
// 	c.SetCookie(cookie)
// 	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
// }
