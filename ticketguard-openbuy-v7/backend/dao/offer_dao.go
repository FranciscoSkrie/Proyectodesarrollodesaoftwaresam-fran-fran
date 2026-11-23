package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ticketguard/backend/domain"
)

type OfferDAO struct{ db *gorm.DB }

func NewOfferDAO(db *gorm.DB) *OfferDAO { return &OfferDAO{db: db} }

func (d *OfferDAO) WithTx(tx *gorm.DB) *OfferDAO { return &OfferDAO{db: tx} }

func (d *OfferDAO) Create(offer *domain.Offer) error { return d.db.Create(offer).Error }

func (d *OfferDAO) Update(offer *domain.Offer) error { return d.db.Save(offer).Error }

func (d *OfferDAO) FindByID(id uint) (*domain.Offer, error) {
	var offer domain.Offer
	if err := d.db.Preload("Event").Preload("Publisher").Preload("Ticket").First(&offer, id).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (d *OfferDAO) FindByIDForUpdate(id uint) (*domain.Offer, error) {
	var offer domain.Offer
	if err := d.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Event").Preload("Publisher").Preload("Ticket").First(&offer, id).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (d *OfferDAO) ListActiveByEvent(eventID uint) ([]domain.Offer, error) {
	var offers []domain.Offer
	err := d.db.Preload("Publisher").Preload("Ticket").Where("event_id = ? AND status = ?", eventID, domain.OfferActive).Order("price asc").Find(&offers).Error
	return offers, err
}

func (d *OfferDAO) ListByPublisher(publisherID uint) ([]domain.Offer, error) {
	var offers []domain.Offer
	err := d.db.Preload("Event").Preload("Ticket").Where("seller_id = ?", publisherID).Order("created_at desc").Find(&offers).Error
	return offers, err
}

func (d *OfferDAO) ListFlagged() ([]domain.Offer, error) {
	var offers []domain.Offer
	err := d.db.Preload("Event").Preload("Publisher").Preload("Ticket").
		Where("alert_cleared = ? AND (scan_status IN ? OR status = ?)", false, []domain.ScanStatus{domain.ScanSuspicious, domain.ScanMalicious}, domain.OfferBlocked).
		Order("created_at desc").Find(&offers).Error
	return offers, err
}

func (d *OfferDAO) ClearFlaggedAlerts() error {
	return d.db.Model(&domain.Offer{}).
		Where("alert_cleared = ? AND (scan_status IN ? OR status = ?)", false, []domain.ScanStatus{domain.ScanSuspicious, domain.ScanMalicious}, domain.OfferBlocked).
		Update("alert_cleared", true).Error
}
