package services

import (
	"context"
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
	scanner clients.LinkScanner
}

type OfferInput struct {
	EventID     uint    `json:"event_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	ExternalURL string  `json:"external_url"`
}

func NewOfferService(offers *dao.OfferDAO, events *dao.EventDAO, scanner clients.LinkScanner) *OfferService {
	return &OfferService{offers: offers, events: events, scanner: scanner}
}

func (s *OfferService) ListForEvent(eventID uint) ([]domain.Offer, error) {
	return s.offers.ListActiveByEvent(eventID)
}

func (s *OfferService) ListBySeller(sellerID uint) ([]domain.Offer, error) {
	return s.offers.ListBySeller(sellerID)
}

func (s *OfferService) Create(ctx context.Context, sellerID uint, input OfferInput) (*domain.Offer, error) {
	event, err := s.events.FindByID(input.EventID)
	if err != nil {
		return nil, err
	}
	if event.Status != domain.EventActive {
		return nil, utils.ErrInvalidStatus
	}

	scanResult, err := s.scanner.ScanURL(ctx, input.ExternalURL)
	if err != nil {
		return nil, err
	}
	status := domain.OfferActive
	if scanResult.Status == domain.ScanMalicious {
		status = domain.OfferBlocked
	}
	offer := &domain.Offer{
		EventID:     input.EventID,
		SellerID:    sellerID,
		Title:       strings.TrimSpace(input.Title),
		Price:       input.Price,
		Quantity:    input.Quantity,
		ExternalURL: strings.TrimSpace(input.ExternalURL),
		ScanStatus:  scanResult.Status,
		ScanVerdict: scanResult.Verdict,
		Status:      status,
	}
	if err := offer.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", utils.ErrInvalidInput, err.Error())
	}
	if err := s.offers.Create(offer); err != nil {
		return nil, err
	}
	return offer, nil
}
