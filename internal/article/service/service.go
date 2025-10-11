package service

import (
	"errors"
	"fmt"
	"log"
	"news/pkg/models"

	"gorm.io/gorm"
)

func GetArticlesWithDetails(db *gorm.DB) ([]models.Article, error) {
	var articles []models.Article

	err := db.
		Select("articles.id, articles.article_title, articles.article_content, articles.author_id, articles.created_at").
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
	return articles, nil
}

func DeleteArticleByID(db *gorm.DB, articleID uint64) error {
	err := db.Where("id = ?", articleID).Delete(&models.Article{}).Error
	if err != nil {
		log.Printf("error delete article with id %d:%s", articleID, err)
		return err
	}
	return nil
}

func GetArticleByIDFromDB(db *gorm.DB, articleID uint64) (models.Article, error) {
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
