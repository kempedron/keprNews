package database

import (
	"context"
	"fmt"
	"log"
	"news/pkg/models"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

var redisClient *redis.Client

func InitDB() error {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
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
		// не обрываем, т.к возможно столбец уже существует
		log.Printf("Ошибка при добавлении search_vector: %v", err)
	}

	// Создаем индекс GIN для быстрого поиска (если еще не существует)
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_articles_search_vector ON articles USING gin(search_vector)").Error
	if err != nil {
		log.Printf("Ошибка при создании индекса для search_vector: %v", err)
		return err
	}

	// Создаем индекс для поиска по тегам (если еще не существует)
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_tags_content ON tags USING gin(to_tsvector('russian', tag_content))").Error
	if err != nil {
		log.Printf("Ошибка при создании индекса для тегов: %v", err)
		return err
	}
	return nil
}

func InitRedis() error {
	if err := godotenv.Load("/root/.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	redisURL := os.Getenv("REDIS_URL")
	log.Println(redisURL)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return err
	}
	log.Println("Successfully connected to Redis")
	return nil
}
