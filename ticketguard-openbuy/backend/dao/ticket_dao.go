package dao

import (
	"gorm.io/gorm"
	"ticketguard/backend/domain"
)

type TicketDAO struct{ db *gorm.DB }

func NewTicketDAO(db *gorm.DB) *TicketDAO { return &TicketDAO{db: db} }

func (d *TicketDAO) WithTx(tx *gorm.DB) *TicketDAO { return &TicketDAO{db: tx} }

func (d *TicketDAO) Create(ticket *domain.Ticket) error { return d.db.Create(ticket).Error }

func (d *TicketDAO) Update(ticket *domain.Ticket) error { return d.db.Save(ticket).Error }

func (d *TicketDAO) FindByID(id uint) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := d.db.Preload("Event").Preload("Offer").First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (d *TicketDAO) ListByOwner(ownerID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := d.db.Preload("Event").Preload("Offer").Where("owner_id = ?", ownerID).Order("created_at desc").Find(&tickets).Error
	return tickets, err
}

func (d *TicketDAO) CountActiveByEvent(eventID uint) (int64, error) {
	var count int64
	err := d.db.Model(&domain.Ticket{}).Where("event_id = ? AND status = ?", eventID, domain.TicketActive).Count(&count).Error
	return count, err
}

func (d *TicketDAO) ListActiveByEvent(eventID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := d.db.Preload("Owner").Where("event_id = ? AND status = ?", eventID, domain.TicketActive).Find(&tickets).Error
	return tickets, err
}
