package dao

import (
	"gorm.io/gorm"
	"ticketguard/backend/domain"
)

type OfferDAO struct{ db *gorm.DB }

func NewOfferDAO(db *gorm.DB) *OfferDAO { return &OfferDAO{db: db} }

func (d *OfferDAO) WithTx(tx *gorm.DB) *OfferDAO { return &OfferDAO{db: tx} }

func (d *OfferDAO) Create(offer *domain.Offer) error { return d.db.Create(offer).Error }

func (d *OfferDAO) Update(offer *domain.Offer) error { return d.db.Save(offer).Error }

func (d *OfferDAO) FindByID(id uint) (*domain.Offer, error) {
	var offer domain.Offer
	if err := d.db.Preload("Event").Preload("Seller").First(&offer, id).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (d *OfferDAO) ListActiveByEvent(eventID uint) ([]domain.Offer, error) {
	var offers []domain.Offer
	err := d.db.Preload("Seller").Where("event_id = ? AND status = ?", eventID, domain.OfferActive).Order("price asc").Find(&offers).Error
	return offers, err
}

func (d *OfferDAO) ListBySeller(sellerID uint) ([]domain.Offer, error) {
	var offers []domain.Offer
	err := d.db.Preload("Event").Where("seller_id = ?", sellerID).Order("created_at desc").Find(&offers).Error
	return offers, err
}
