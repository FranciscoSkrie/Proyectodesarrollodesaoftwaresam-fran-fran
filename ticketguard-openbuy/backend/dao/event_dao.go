package dao

import (
	"strings"

	"gorm.io/gorm"
	"ticketguard/backend/domain"
)

type EventFilters struct {
	Q        string
	Category string
	Status   domain.EventStatus
}

type EventDAO struct{ db *gorm.DB }

func NewEventDAO(db *gorm.DB) *EventDAO { return &EventDAO{db: db} }

func (d *EventDAO) WithTx(tx *gorm.DB) *EventDAO { return &EventDAO{db: tx} }

func (d *EventDAO) List(filters EventFilters) ([]domain.Event, error) {
	var events []domain.Event
	query := d.db.Order("starts_at asc")
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if strings.TrimSpace(filters.Category) != "" {
		query = query.Where("category = ?", strings.TrimSpace(filters.Category))
	}
	if strings.TrimSpace(filters.Q) != "" {
		like := "%" + strings.TrimSpace(filters.Q) + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR location LIKE ?", like, like, like)
	}
	return events, query.Find(&events).Error
}

func (d *EventDAO) FindByID(id uint) (*domain.Event, error) {
	var event domain.Event
	if err := d.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *EventDAO) Create(event *domain.Event) error { return d.db.Create(event).Error }

func (d *EventDAO) Update(event *domain.Event) error { return d.db.Save(event).Error }
