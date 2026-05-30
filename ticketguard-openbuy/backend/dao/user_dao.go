package dao

import (
	"gorm.io/gorm"
	"ticketguard/backend/domain"
)

type UserDAO struct{ db *gorm.DB }

func NewUserDAO(db *gorm.DB) *UserDAO { return &UserDAO{db: db} }

func (d *UserDAO) WithTx(tx *gorm.DB) *UserDAO { return &UserDAO{db: tx} }

func (d *UserDAO) Create(user *domain.User) error { return d.db.Create(user).Error }

func (d *UserDAO) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := d.db.Where("email = ?", domain.NormalizeEmail(email)).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := d.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) Update(user *domain.User) error { return d.db.Save(user).Error }
