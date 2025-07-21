package dto

import (
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"time"
)

type UserDTO struct {
	Username string `json:"username" validate:"required,min=3,max=30,alpha"`
	Password string `json:"password" validate:"required,min=8,max=15,containsany=!@#?$&%,containsany=1234567890"`
}

type AdDTO struct {
	Title    string  `json:"title" validate:"required,min=5,max=20"`
	Text     string  `json:"text" validate:"required,min=20,max=1000"`
	ImageURL string  `json:"image_url" validate:"required,url"`
	Price    float64 `json:"price" validate:"required,gt=0"`
}

type AdResponse struct {
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	ImageURL  string    `json:"image_url"`
	Price     float64   `json:"price"`
	Username  string    `json:"username"`
	IsOwner   bool      `json:"is_owner,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAdResponse(ad *entity.Ad) *AdResponse {
	return &AdResponse{
		Title:     ad.Title,
		Text:      ad.Text,
		ImageURL:  ad.ImageURL,
		Price:     ad.Price,
		Username:  ad.Author.Username,
		CreatedAt: ad.CreatedAt,
	}
}

func (ar *AdResponse) ProcessOwner(ad *entity.Ad, curAuthorizedUserID uuid.UUID) {
	ar.IsOwner = ad.Author.ID == curAuthorizedUserID
}
