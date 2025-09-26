package database

import (
	"fmt"
	"log"
	"news/internal/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("не удалось подключиться к БД:", err)
	}
	err = DB.AutoMigrate(
		&models.User{},
		&models.Tag{},
		&models.Article{},
	)
	if err != nil {
		log.Printf("error migrate DB: %s", err)
		panic(err)
	}
	// Добавляем поле для полнотекстового поиска (если еще не существует)
	err = DB.Exec(`
        ALTER TABLE articles 
        ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
            setweight(to_tsvector('russian', coalesce(article_title, '')), 'A') ||
            setweight(to_tsvector('russian', coalesce(article_content, '')), 'B')
        ) STORED
    `).Error
	if err != nil {
		log.Printf("Ошибка при добавлении search_vector: %v", err)
		// Не прерываем выполнение, так как столбец может уже существовать
	}

	// Создаем индекс GIN для быстрого поиска (если еще не существует)
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_articles_search_vector ON articles USING gin(search_vector)").Error
	if err != nil {
		log.Printf("Ошибка при создании индекса для search_vector: %v", err)
	}

	// Создаем индекс для поиска по тегам (если еще не существует)
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_tags_content ON tags USING gin(to_tsvector('russian', tag_content))").Error
	if err != nil {
		log.Printf("Ошибка при создании индекса для тегов: %v", err)
	}

}
