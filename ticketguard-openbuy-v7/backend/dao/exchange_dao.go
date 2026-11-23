package dao

import (
	"gorm.io/gorm"
	"ticketguard/backend/domain"
)

type ExchangeDAO struct{ db *gorm.DB }

func NewExchangeDAO(db *gorm.DB) *ExchangeDAO { return &ExchangeDAO{db: db} }

func (d *ExchangeDAO) WithTx(tx *gorm.DB) *ExchangeDAO { return &ExchangeDAO{db: tx} }

func (d *ExchangeDAO) Create(exchange *domain.ExchangeRequest) error {
	return d.db.Create(exchange).Error
}

func (d *ExchangeDAO) Update(exchange *domain.ExchangeRequest) error {
	return d.db.Save(exchange).Error
}

func (d *ExchangeDAO) FindByID(id uint) (*domain.ExchangeRequest, error) {
	var exchange domain.ExchangeRequest
	err := d.preload().First(&exchange, id).Error
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}

func (d *ExchangeDAO) ListInbox(userID uint) ([]domain.ExchangeRequest, error) {
	var exchanges []domain.ExchangeRequest
	err := d.preload().Where("receiver_id = ?", userID).Order("created_at desc").Find(&exchanges).Error
	return exchanges, err
}

func (d *ExchangeDAO) ListSent(userID uint) ([]domain.ExchangeRequest, error) {
	var exchanges []domain.ExchangeRequest
	err := d.preload().Where("requester_id = ?", userID).Order("created_at desc").Find(&exchanges).Error
	return exchanges, err
}

func (d *ExchangeDAO) preload() *gorm.DB {
	return d.db.Preload("Offer").Preload("Offer.Publisher").Preload("Event").Preload("Requester").Preload("Receiver").Preload("OfferedTicket").Preload("OfferedTicket.Event").Preload("TargetTicket").Preload("TargetTicket.Event")
}
