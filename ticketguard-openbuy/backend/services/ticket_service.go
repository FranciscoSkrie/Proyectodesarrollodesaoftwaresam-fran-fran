package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/utils"
)

type TicketService struct {
	db      *gorm.DB
	users   *dao.UserDAO
	events  *dao.EventDAO
	offers  *dao.OfferDAO
	tickets *dao.TicketDAO
}

type TransferInput struct {
	Email string `json:"email" binding:"required,email"`
}

func NewTicketService(db *gorm.DB, users *dao.UserDAO, events *dao.EventDAO, offers *dao.OfferDAO, tickets *dao.TicketDAO) *TicketService {
	return &TicketService{db: db, users: users, events: events, offers: offers, tickets: tickets}
}

func (s *TicketService) Buy(userID, offerID uint) (*domain.Ticket, error) {
	var created *domain.Ticket
	err := s.db.Transaction(func(tx *gorm.DB) error {
		offerDAO := s.offers.WithTx(tx)
		ticketDAO := s.tickets.WithTx(tx)

		offer, err := offerDAO.FindByID(offerID)
		if err != nil {
			return err
		}
		if offer.Status != domain.OfferActive {
			return utils.ErrUnsafeOffer
		}
		if offer.ScanStatus == domain.ScanMalicious {
			return utils.ErrUnsafeOffer
		}
		if offer.Quantity <= 0 {
			return utils.ErrInsufficient
		}
		if offer.Event.Status != domain.EventActive {
			return utils.ErrInvalidStatus
		}
		activeCount, err := ticketDAO.CountActiveByEvent(offer.EventID)
		if err != nil {
			return err
		}
		if activeCount >= int64(offer.Event.Capacity) {
			return utils.ErrInsufficient
		}
		ticket := &domain.Ticket{
			Code:    generateTicketCode(userID, offerID),
			EventID: offer.EventID,
			OfferID: offer.ID,
			OwnerID: userID,
			Price:   offer.Price,
			Status:  domain.TicketActive,
		}
		offer.Quantity--
		if err := offerDAO.Update(offer); err != nil {
			return err
		}
		if err := ticketDAO.Create(ticket); err != nil {
			return err
		}
		created = ticket
		return nil
	})
	return created, err
}

func (s *TicketService) ListMine(userID uint) ([]domain.Ticket, error) {
	return s.tickets.ListByOwner(userID)
}

func (s *TicketService) Cancel(userID, ticketID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ticketDAO := s.tickets.WithTx(tx)
		offerDAO := s.offers.WithTx(tx)
		ticket, err := ticketDAO.FindByID(ticketID)
		if err != nil {
			return err
		}
		if ticket.OwnerID != userID {
			return utils.ErrNotFound
		}
		if ticket.Status != domain.TicketActive {
			return utils.ErrInvalidStatus
		}
		ticket.Status = domain.TicketCancelled
		if err := ticketDAO.Update(ticket); err != nil {
			return err
		}
		offer, err := offerDAO.FindByID(ticket.OfferID)
		if err == nil && offer.Status == domain.OfferActive {
			offer.Quantity++
			return offerDAO.Update(offer)
		}
		return nil
	})
}

func (s *TicketService) Transfer(userID, ticketID uint, targetEmail string) (*domain.Ticket, error) {
	var updated *domain.Ticket
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ticketDAO := s.tickets.WithTx(tx)
		userDAO := s.users.WithTx(tx)
		ticket, err := ticketDAO.FindByID(ticketID)
		if err != nil {
			return err
		}
		if ticket.OwnerID != userID {
			return utils.ErrNotFound
		}
		if ticket.Status != domain.TicketActive {
			return utils.ErrInvalidStatus
		}
		target, err := userDAO.FindByEmail(targetEmail)
		if err != nil {
			return err
		}
		ticket.OwnerID = target.ID
		ticket.Status = domain.TicketTransferred
		if err := ticketDAO.Update(ticket); err != nil {
			return err
		}
		updated = ticket
		return nil
	})
	return updated, err
}

func generateTicketCode(userID, offerID uint) string {
	return fmt.Sprintf("TG-%d-%d-%d", userID, offerID, time.Now().UnixNano())
}
