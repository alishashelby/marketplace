package entity

import (
	"github.com/google/uuid"
	"time"
)

type Ad struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	Title     string    `json:"title" bson:"title"`
	Text      string    `json:"text" bson:"text"`
	ImageURL  string    `json:"image_url" bson:"image_url"`
	Price     float64   `json:"price" bson:"price"`
	Author    *Author   `json:"author" bson:"author"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Author struct {
	Username string    `json:"username" bson:"username"`
	ID       uuid.UUID `json:"id" bson:"_id"`
}

func NewAd(title, text, imageURL string, price float64, user *User) *Ad {
	return &Ad{
		ID:       uuid.New(),
		Title:    title,
		Text:     text,
		ImageURL: imageURL,
		Price:    price,
		Author: &Author{
			Username: user.Username,
			ID:       user.ID,
		},
		CreatedAt: time.Now(),
	}
}

const (
	OrderByAsc        = 1
	OrderByDesc       = -1
	SortByCreatedAt   = "created_at"
	SortByPrice       = "price"
	LimitMaxValue     = 40
	LimitDefaultValue = 10
)

const (
	ParamPage     = "page"
	ParamLimit    = "limit"
	ParamSortBy   = "sort_by"
	ParamOrderBy  = "order_by"
	ParamMinPrice = "min_price"
	ParamMaxPrice = "max_price"
)

type Options struct {
	Page     int
	Limit    int
	SortBy   string
	OrderBy  int
	MinPrice float64
	MaxPrice float64
}
