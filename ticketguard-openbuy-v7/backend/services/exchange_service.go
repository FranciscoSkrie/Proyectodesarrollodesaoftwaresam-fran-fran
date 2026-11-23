package services

import (
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/utils"
)

type ExchangeService struct {
	db        *gorm.DB
	exchanges *dao.ExchangeDAO
	offers    *dao.OfferDAO
	tickets   *dao.TicketDAO
	users     *dao.UserDAO
}

type ExchangeInput struct {
	OfferedTicketID uint   `json:"offered_ticket_id" binding:"required"`
	Message         string `json:"message"`
	ContactInfo     string `json:"contact_info"`
}

func NewExchangeService(db *gorm.DB, exchanges *dao.ExchangeDAO, offers *dao.OfferDAO, tickets *dao.TicketDAO, users *dao.UserDAO) *ExchangeService {
	return &ExchangeService{db: db, exchanges: exchanges, offers: offers, tickets: tickets, users: users}
}

func (s *ExchangeService) Request(userID, offerID uint, input ExchangeInput) (*domain.ExchangeRequest, error) {
	var created *domain.ExchangeRequest
	err := s.db.Transaction(func(tx *gorm.DB) error {
		exchangeDAO := s.exchanges.WithTx(tx)
		offerDAO := s.offers.WithTx(tx)
		ticketDAO := s.tickets.WithTx(tx)
		userDAO := s.users.WithTx(tx)

		offer, err := offerDAO.FindByID(offerID)
		if err != nil {
			return err
		}
		if offer.Status != domain.OfferActive || offer.ScanStatus == domain.ScanMalicious {
			return utils.ErrUnsafeOffer
		}
		if offer.PublisherID == userID {
			return fmt.Errorf("%w: no podés solicitar intercambio sobre tu propia publicación", utils.ErrInvalidInput)
		}
		if offer.TicketID == nil || *offer.TicketID == 0 {
			return fmt.Errorf("%w: la publicación no tiene una entrada trazable asociada", utils.ErrInvalidInput)
		}

		offeredTicket, err := ticketDAO.FindActiveByOwner(input.OfferedTicketID, userID)
		if err != nil {
			return utils.ErrNotFound
		}
		targetTicket, err := ticketDAO.FindActiveByOwner(*offer.TicketID, offer.PublisherID)
		if err != nil {
			return utils.ErrNotFound
		}
		if offeredTicket.EventID != offer.EventID || targetTicket.EventID != offer.EventID {
			return fmt.Errorf("%w: las entradas del intercambio deben pertenecer al mismo foro", utils.ErrInvalidInput)
		}
		requester, err := userDAO.FindByID(userID)
		if err != nil {
			return err
		}

		contact := strings.TrimSpace(input.ContactInfo)
		if contact == "" {
			contact = requester.Email
		}
		priceDifference := roundMoney(targetTicket.Price - offeredTicket.Price)
		exchange := &domain.ExchangeRequest{
			OfferID:         offer.ID,
			EventID:         offer.EventID,
			RequesterID:     userID,
			ReceiverID:      offer.PublisherID,
			OfferedTicketID: offeredTicket.ID,
			TargetTicketID:  targetTicket.ID,
			Message:         strings.TrimSpace(input.Message),
			ContactInfo:     contact,
			PriceDifference: priceDifference,
			Status:          domain.ExchangePending,
		}
		if err := exchangeDAO.Create(exchange); err != nil {
			return err
		}
		created = exchange
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.exchanges.FindByID(created.ID)
}

func (s *ExchangeService) ListInbox(userID uint) ([]domain.ExchangeRequest, error) {
	return s.exchanges.ListInbox(userID)
}

func (s *ExchangeService) ListSent(userID uint) ([]domain.ExchangeRequest, error) {
	return s.exchanges.ListSent(userID)
}

func (s *ExchangeService) Accept(userID, exchangeID uint) (*domain.ExchangeRequest, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		exchangeDAO := s.exchanges.WithTx(tx)
		ticketDAO := s.tickets.WithTx(tx)
		offerDAO := s.offers.WithTx(tx)
		exchange, err := exchangeDAO.FindByID(exchangeID)
		if err != nil {
			return err
		}
		if exchange.ReceiverID != userID {
			return utils.ErrNotFound
		}
		if exchange.Status != domain.ExchangePending {
			return utils.ErrInvalidStatus
		}

		offeredTicket, err := ticketDAO.FindActiveByOwner(exchange.OfferedTicketID, exchange.RequesterID)
		if err != nil {
			return utils.ErrInvalidStatus
		}
		targetTicket, err := ticketDAO.FindActiveByOwner(exchange.TargetTicketID, exchange.ReceiverID)
		if err != nil {
			return utils.ErrInvalidStatus
		}

		offeredTicket.OwnerID = exchange.ReceiverID
		targetTicket.OwnerID = exchange.RequesterID
		if err := ticketDAO.Update(offeredTicket); err != nil {
			return err
		}
		if err := ticketDAO.Update(targetTicket); err != nil {
			return err
		}
		offer, err := offerDAO.FindByID(exchange.OfferID)
		if err != nil {
			return err
		}
		offer.Status = domain.OfferPaused
		if err := offerDAO.Update(offer); err != nil {
			return err
		}

		exchange.Status = domain.ExchangeAccepted
		return exchangeDAO.Update(exchange)
	})
	if err != nil {
		return nil, err
	}
	return s.exchanges.FindByID(exchangeID)
}

func (s *ExchangeService) Reject(userID, exchangeID uint) (*domain.ExchangeRequest, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		exchangeDAO := s.exchanges.WithTx(tx)
		exchange, err := exchangeDAO.FindByID(exchangeID)
		if err != nil {
			return err
		}
		if exchange.ReceiverID != userID {
			return utils.ErrNotFound
		}
		if exchange.Status != domain.ExchangePending {
			return utils.ErrInvalidStatus
		}
		exchange.Status = domain.ExchangeRejected
		return exchangeDAO.Update(exchange)
	})
	if err != nil {
		return nil, err
	}
	return s.exchanges.FindByID(exchangeID)
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
