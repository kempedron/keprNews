package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string    `gorm:"type:varchar(50);not null;unique" json:"username"`
	PasswordHash string    `gorm:"type:varchar(100);not null" json:"-"`
	Articles     []Article `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
}

func (u *User) HashPassword(password string) error {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedpassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (User) TableName() string {
	return "users"
}

type Tag struct {
	gorm.Model
	TagContent string    `gorm:"type:varchar(50);not null;unique" json:"tagContent"`
	Articles   []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

type Article struct {
	gorm.Model
	AuthorID       uint   `gorm:"not null" json:"author_id"`
	ArticleTitle   string `gorm:"type:text;not null" json:"article_title"`
	ArticleContent string `gorm:"type:text;not null" json:"article_content"`
	NumViews       int    `gorm:"default:0" json:"num_views"`
	Author         User   `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"author"`
	Tags           []Tag  `gorm:"many2many:article_tags;" json:"tags,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}
