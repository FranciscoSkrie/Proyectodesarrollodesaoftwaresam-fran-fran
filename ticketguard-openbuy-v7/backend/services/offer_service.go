package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ticketguard/backend/clients"
	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/utils"
)

type OfferService struct {
	offers  *dao.OfferDAO
	events  *dao.EventDAO
	users   *dao.UserDAO
	tickets *dao.TicketDAO
	scanner clients.LinkScanner
}

type OfferInput struct {
	EventID         uint           `json:"event_id" binding:"required"`
	TicketID        uint           `json:"ticket_id"`
	Title           string         `json:"title" binding:"required"`
	Price           float64        `json:"price" binding:"required"`
	Quantity        int            `json:"quantity"`
	ExternalURL     string         `json:"external_url"`
	ContactInfo     string         `json:"contact_info"`
	ExchangeDetails string         `json:"exchange_details"`
	Telemetry       map[string]any `json:"telemetry"`
}

type OfferEvidenceInput struct {
	RequestIP      string
	UserAgent      string
	BrowserData    string
	RequestHeaders string
}

func NewOfferService(offers *dao.OfferDAO, events *dao.EventDAO, users *dao.UserDAO, tickets *dao.TicketDAO, scanner clients.LinkScanner) *OfferService {
	return &OfferService{offers: offers, events: events, users: users, tickets: tickets, scanner: scanner}
}

func (s *OfferService) ListForEvent(eventID uint) ([]domain.Offer, error) {
	return s.offers.ListActiveByEvent(eventID)
}

func (s *OfferService) ListByPublisher(publisherID uint) ([]domain.Offer, error) {
	return s.offers.ListByPublisher(publisherID)
}

func (s *OfferService) ListFlagged() ([]domain.Offer, error) {
	return s.offers.ListFlagged()
}

func (s *OfferService) ClearFlaggedAlerts() error {
	return s.offers.ClearFlaggedAlerts()
}

func (s *OfferService) Create(ctx context.Context, userID uint, input OfferInput, evidence OfferEvidenceInput) (*domain.Offer, error) {
	event, err := s.events.FindByID(input.EventID)
	if err != nil {
		return nil, err
	}
	if event.Status != domain.EventActive {
		return nil, utils.ErrInvalidStatus
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}

	quantity := input.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	var ticketID *uint
	if input.TicketID > 0 {
		ticket, err := s.tickets.FindActiveByOwner(input.TicketID, userID)
		if err != nil {
			return nil, utils.ErrNotFound
		}
		if ticket.EventID != input.EventID {
			return nil, fmt.Errorf("%w: la entrada seleccionada no pertenece al foro elegido", utils.ErrInvalidInput)
		}
		ticketID = &ticket.ID
		quantity = 1
	}

	scanResult, err := s.scanner.ScanURL(ctx, input.ExternalURL)
	if err != nil {
		return nil, err
	}
	status := domain.OfferActive
	if scanResult.Status == domain.ScanMalicious {
		status = domain.OfferBlocked
	}

	telemetryJSON := ""
	if input.Telemetry != nil {
		if raw, err := json.MarshalIndent(input.Telemetry, "", "  "); err == nil {
			telemetryJSON = string(raw)
		}
	}

	offer := &domain.Offer{
		EventID:          input.EventID,
		PublisherID:      userID,
		TicketID:         ticketID,
		Title:            strings.TrimSpace(input.Title),
		Price:            input.Price,
		Quantity:         quantity,
		ExternalURL:      strings.TrimSpace(input.ExternalURL),
		ContactInfo:      strings.TrimSpace(input.ContactInfo),
		ExchangeDetails:  strings.TrimSpace(input.ExchangeDetails),
		ScanStatus:       scanResult.Status,
		ScanVerdict:      scanResult.Verdict,
		Status:           status,
		AlertCleared:     false,
		RequestIP:        strings.TrimSpace(evidence.RequestIP),
		RequestUserName:  strings.TrimSpace(user.Name),
		RequestUserEmail: strings.TrimSpace(user.Email),
		UserAgent:        strings.TrimSpace(evidence.UserAgent),
		BrowserData:      strings.TrimSpace(evidence.BrowserData),
		RequestHeaders:   strings.TrimSpace(evidence.RequestHeaders),
		TelemetryJSON:    telemetryJSON,
	}
	if offer.BrowserData == "" {
		offer.BrowserData = offer.UserAgent
	}
	if offer.ContactInfo == "" {
		offer.ContactInfo = user.Email
	}
	if err := offer.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", utils.ErrInvalidInput, err.Error())
	}
	if err := s.offers.Create(offer); err != nil {
		return nil, err
	}
	return offer, nil
}
