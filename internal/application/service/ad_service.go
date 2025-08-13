package service

import "github.com/alishashelby/marketplace/internal/domain/entity"

//go:generate mockgen -source=ad_service.go -destination=ad_repo_mock.go -package=service AdRepository
type AdRepository interface {
	Save(ad *entity.Ad) error
	FindAll(ops *entity.Options) ([]*entity.Ad, error)
}

type AdService struct {
	repo AdRepository
}

func NewAdService(repo AdRepository) *AdService {
	return &AdService{repo: repo}
}

func (s *AdService) Create(ad *entity.Ad) error {
	return s.repo.Save(ad)
}

func (s *AdService) GetAds(ops *entity.Options) ([]*entity.Ad, error) {
	return s.repo.FindAll(ops)
}
