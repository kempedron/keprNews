package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"news/pkg/database"
	"news/pkg/models"
	"time"

	"gorm.io/gorm"
)

func GetArticlesWithDetails(db *gorm.DB) ([]models.Article, error) {
	var articles []models.Article
	cacheKey := "articles:list"
	ctx := context.Background()

	cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedData), &articles); err == nil {
			log.Println("successfully get all articles from cache")
			return articles, nil

		}
	}
	log.Printf("successfully get data from cache: %v", articles)

	err = db.Select("articles.id, articles.article_title, articles.article_content, articles.author_id, articles.created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username")
		}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, tag_content")
		}).
		Order("id DESC").
		Limit(10).
		Find(&articles).Error

	if err != nil {
		return nil, err
	}
	if err := saveToCache(ctx, cacheKey, articles, 30*time.Minute); err != nil {
		log.Printf("failed to cache articles: %s", err)
	} else {
		log.Printf("articles succesfully cached")
	}

	return articles, nil
}

func DeleteArticleByID(db *gorm.DB, articleID uint64) error {
	ctx := context.Background()

	err := db.Where("id = ?", articleID).Delete(&models.Article{}).Error
	if err != nil {
		log.Printf("error delete article with id %d:%s", articleID, err)
		return err
	}
	if err := invalidateArticleCache(ctx, articleID); err != nil {
		log.Printf("failed cache invalidation for article %d: %s", articleID, err)
	} else {
		log.Printf("successfully cache invalidation for article %d", articleID)
	}
	return nil
}

func GetArticleByID(db *gorm.DB, articleID uint64) (models.Article, error) {
	var article models.Article
	ctx := context.Background()
	cachedKey := fmt.Sprint(articleID)
	cachedDate, err := database.RedisClient.Get(ctx, cachedKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedDate), &article); err == nil {
			log.Println("successfully get article by id from cache")
			return article, nil
		}
	}

	err = db.Preload("Author").Preload("Tags").First(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Article{}, fmt.Errorf("статья с ID %d не найдена ", articleID)
		}
		return models.Article{}, fmt.Errorf("ошибка при получении статьи: %s", err)
	}
	if err := saveToCache(ctx, cachedKey, article, 30*time.Minute); err != nil {
		log.Printf("failed to cache article %d: %s", articleID, err)
	} else {
		log.Printf("article %d succesfully cached", articleID)
	}
	return article, nil
}

func saveToCache(ctx context.Context, key string, data interface{}, expirations time.Duration) error {
	jsondata, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := database.RedisClient.Set(ctx, key, jsondata, expirations).Err(); err != nil {
		return fmt.Errorf("redis set data error: %w", err)
	}
	return nil
}

func invalidateArticleCache(ctx context.Context, articleID uint64) error {
	articleKey := fmt.Sprintf("article:%d", articleID)
	if err := database.RedisClient.Del(ctx, articleKey).Err(); err != nil {
		return fmt.Errorf("failed to delete article %d cache: %s", articleID, err)
	}
	return nil
}
