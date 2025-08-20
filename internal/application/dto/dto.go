package dto

import (
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"time"
)

type UserDTO struct {
	Username string `json:"username" validate:"required,min=3,max=30,alpha" example:"alisha"`
	Password string `json:"password" validate:"required,min=8,max=15,containsany=!@#?$&%,containsany=1234567890" example:"1234567&"`
}

type AdDTO struct {
	Title    string  `json:"title" validate:"required,min=5,max=20" example:"Title of test ad"`
	Text     string  `json:"text" validate:"required,min=20,max=1000" example:"This is the test ad. Check new image."`
	ImageURL string  `json:"image_url" validate:"required,url" example:"https://upload.wikimedia.org/wikipedia/commons/c/c7/Tabby_cat_with_blue_eyes-3336579.jpg"`
	Price    float64 `json:"price" validate:"required,gt=0" example:"1500.5"`
}

type AdResponse struct {
	Title     string    `json:"title" example:"Title of test ad"`
	Text      string    `json:"text" example:"This is the test ad. Check new image."`
	ImageURL  string    `json:"image_url" example:"https://upload.wikimedia.org/wikipedia/commons/c/c7/Tabby_cat_with_blue_eyes-3336579.jpg"`
	Price     float64   `json:"price" example:"1500.5"`
	Username  string    `json:"username" example:"alisha"`
	IsOwner   bool      `json:"is_owner,omitempty" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2025-08-11T19:14:03.187Z"`
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
