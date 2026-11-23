package services

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"

	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/utils"
)

type EventService struct {
	events  *dao.EventDAO
	tickets *dao.TicketDAO
}

type EventInput struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Location        string    `json:"location" binding:"required"`
	StartsAt        time.Time `json:"starts_at" binding:"required"`
	DurationMinutes int       `json:"duration_minutes"`
	Capacity        int       `json:"capacity" binding:"required"`
	EntryPrice      float64   `json:"entry_price"`
	ImageURL        string    `json:"image_url"`
}

type EventReport struct {
	Event         domain.Event    `json:"event"`
	Capacity      int             `json:"capacity"`
	Sold          int64           `json:"sold"`
	Available     int64           `json:"available"`
	OccupationPct float64         `json:"occupation_pct"`
	Buyers        []domain.Ticket `json:"buyers"`
}

func NewEventService(events *dao.EventDAO, tickets *dao.TicketDAO) *EventService {
	return &EventService{events: events, tickets: tickets}
}

func (s *EventService) List(q, category string) ([]domain.Event, error) {
	return s.events.List(dao.EventFilters{Q: q, Category: category, Status: domain.EventActive})
}

func (s *EventService) Get(id uint) (*domain.Event, error) {
	return s.events.FindByID(id)
}

func (s *EventService) Create(adminID uint, input EventInput) (*domain.Event, error) {
	event := mapEventInput(input)
	event.CreatedByID = adminID
	event.Status = domain.EventActive
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", utils.ErrInvalidInput, err.Error())
	}
	if err := s.events.Create(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) Update(id uint, input EventInput) (*domain.Event, error) {
	event, err := s.events.FindByID(id)
	if err != nil {
		return nil, err
	}
	mapped := mapEventInput(input)
	if err := mapped.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", utils.ErrInvalidInput, err.Error())
	}
	event.Title, event.Description, event.Category = mapped.Title, mapped.Description, mapped.Category
	event.Location, event.StartsAt = mapped.Location, mapped.StartsAt
	event.DurationMinutes, event.Capacity = mapped.DurationMinutes, mapped.Capacity
	event.EntryPrice, event.ImageURL = mapped.EntryPrice, mapped.ImageURL
	if err := s.events.UpdateWithCapacityGuard(event); err != nil {
		if err == gorm.ErrInvalidData {
			return nil, fmt.Errorf("%w: capacity cannot be lower than active tickets", utils.ErrInvalidInput)
		}
		return nil, err
	}
	return event, nil
}

func (s *EventService) Cancel(id uint) error {
	event, err := s.events.FindByID(id)
	if err != nil {
		return err
	}
	event.Status = domain.EventCancelled
	return s.events.Update(event)
}

func (s *EventService) Report(id uint) (*EventReport, error) {
	event, err := s.events.FindByID(id)
	if err != nil {
		return nil, err
	}
	sold, err := s.tickets.CountActiveByEvent(id)
	if err != nil {
		return nil, err
	}
	buyers, err := s.tickets.ListActiveByEvent(id)
	if err != nil {
		return nil, err
	}
	available := int64(event.Capacity) - sold
	if available < 0 {
		available = 0
	}
	occupation := 0.0
	if event.Capacity > 0 {
		occupation = (float64(sold) / float64(event.Capacity)) * 100
	}
	return &EventReport{Event: *event, Capacity: event.Capacity, Sold: sold, Available: available, OccupationPct: occupation, Buyers: buyers}, nil
}

func mapEventInput(input EventInput) *domain.Event {
	duration := input.DurationMinutes
	if duration == 0 {
		duration = 120
	}
	entryPrice := input.EntryPrice
	if entryPrice == 0 {
		entryPrice = 15000
	}
	return &domain.Event{
		Title:           strings.TrimSpace(input.Title),
		Description:     strings.TrimSpace(input.Description),
		Category:        strings.TrimSpace(input.Category),
		Location:        strings.TrimSpace(input.Location),
		StartsAt:        input.StartsAt,
		DurationMinutes: duration,
		Capacity:        input.Capacity,
		EntryPrice:      entryPrice,
		ImageURL:        strings.TrimSpace(input.ImageURL),
	}
}
