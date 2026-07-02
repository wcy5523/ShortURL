package dao

import (
	"shorturl/config"
	"shorturl/model"

	"gorm.io/gorm"
)

type VisitLogDAO struct{}

func NewVisitLogDAO() *VisitLogDAO {
	return &VisitLogDAO{}
}

func (d *VisitLogDAO) BatchInsert(logs []model.VisitLog) error {
	if len(logs) == 0 {
		return nil
	}
	return config.DB.Create(&logs).Error
}

func (d *VisitLogDAO) BatchUpdateVisitCount(codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	caseStmt := "CASE short_code "
	args := make([]interface{}, 0, len(codes)*2)
	for _, code := range codes {
		caseStmt += "WHEN ? THEN visit_count + 1 "
		args = append(args, code)
	}
	caseStmt += "END"
	return config.DB.Model(&model.ShortURL{}).
		Where("short_code IN ?", codes).
		Update("visit_count", gorm.Expr(caseStmt, args...)).
		Error
}
