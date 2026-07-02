package dao

import (
	"errors"
	"shorturl/config"
	"shorturl/model"

	"gorm.io/gorm"
)

type ShortURLDAO struct{}

func NewShortURLDAO() *ShortURLDAO {
	return &ShortURLDAO{}
}

func (d *ShortURLDAO) Create(shortURL *model.ShortURL) error {
	return config.DB.Create(shortURL).Error
}

func (d *ShortURLDAO) GetByCode(shortCode string) (*model.ShortURL, error) {
	var shortURL model.ShortURL
	err := config.DB.Where("short_code = ?", shortCode).First(&shortURL).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shortURL, nil
}

func (d *ShortURLDAO) IncrementVisit(shortCode string) error {
	return config.DB.Model(&model.ShortURL{}).
		Where("short_code = ?", shortCode).
		UpdateColumn("visit_count", gorm.Expr("visit_count + ?", 1)).
		Error
}

func (d *ShortURLDAO) IncrementVisitBatch(codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, code := range codes {
		counts[code]++
	}
	for code, cnt := range counts {
		err := config.DB.Model(&model.ShortURL{}).
			Where("short_code = ?", code).
			UpdateColumn("visit_count", gorm.Expr("visit_count + ?", cnt)).
			Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *ShortURLDAO) ListByUserID(userID uint64, page, pageSize int) ([]model.ShortURL, int64, error) {
	var list []model.ShortURL
	var total int64

	err := config.DB.Model(&model.ShortURL{}).
		Where("user_id = ?", userID).
		Count(&total).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error

	return list, total, err
}

func (d *ShortURLDAO) DeleteByCodeAndUserID(shortCode string, userID uint64) error {
	return config.DB.Where("short_code = ? AND user_id = ?", shortCode, userID).
		Delete(&model.ShortURL{}).Error
}
