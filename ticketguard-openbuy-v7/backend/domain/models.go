package domain

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUsuario UserRole = "usuario"
	RoleAdmin   UserRole = "admin"
)

type EventStatus string

const (
	EventActive    EventStatus = "active"
	EventCancelled EventStatus = "cancelled"
)

type OfferStatus string

const (
	OfferActive  OfferStatus = "active"
	OfferPaused  OfferStatus = "paused"
	OfferBlocked OfferStatus = "blocked"
)

type ScanStatus string

const (
	ScanPending    ScanStatus = "pending"
	ScanSafe       ScanStatus = "safe"
	ScanSuspicious ScanStatus = "suspicious"
	ScanMalicious  ScanStatus = "malicious"
)

type TicketStatus string

const (
	TicketActive      TicketStatus = "active"
	TicketCancelled   TicketStatus = "cancelled"
	TicketTransferred TicketStatus = "transferred"
)

type ExchangeStatus string

const (
	ExchangePending   ExchangeStatus = "pending"
	ExchangeAccepted  ExchangeStatus = "accepted"
	ExchangeRejected  ExchangeStatus = "rejected"
	ExchangeCancelled ExchangeStatus = "cancelled"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"size:120;not null"`
	Email        string         `json:"email" gorm:"size:160;not null;uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'usuario'"`
	Balance      float64        `json:"-" gorm:"type:decimal(12,2);not null;default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Event struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Title           string         `json:"title" gorm:"size:180;not null"`
	Description     string         `json:"description" gorm:"type:text"`
	Category        string         `json:"category" gorm:"size:80;index"`
	Location        string         `json:"location" gorm:"size:180;not null"`
	StartsAt        time.Time      `json:"starts_at" gorm:"not null;index"`
	DurationMinutes int            `json:"duration_minutes" gorm:"not null;default:120"`
	Capacity        int            `json:"capacity" gorm:"not null"`
	EntryPrice      float64        `json:"entry_price" gorm:"type:decimal(12,2);not null;default:15000"`
	ImageURL        string         `json:"image_url" gorm:"size:500"`
	Status          EventStatus    `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedByID     uint           `json:"created_by_id"`
	CreatedBy       User           `json:"-" gorm:"foreignKey:CreatedByID"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type Offer struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	EventID          uint           `json:"event_id" gorm:"not null;index"`
	Event            Event          `json:"event" gorm:"foreignKey:EventID"`
	PublisherID      uint           `json:"publisher_id" gorm:"column:seller_id;not null;index"`
	Publisher        User           `json:"publisher" gorm:"foreignKey:PublisherID"`
	TicketID         *uint          `json:"ticket_id" gorm:"index"`
	Ticket           *Ticket        `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
	Title            string         `json:"title" gorm:"size:160;not null"`
	Price            float64        `json:"price" gorm:"type:decimal(12,2);not null"`
	Quantity         int            `json:"quantity" gorm:"not null"`
	ExternalURL      string         `json:"external_url" gorm:"type:text"`
	ContactInfo      string         `json:"contact_info" gorm:"type:text"`
	ExchangeDetails  string         `json:"exchange_details" gorm:"type:text"`
	ScanStatus       ScanStatus     `json:"scan_status" gorm:"type:varchar(20);not null;default:'pending'"`
	ScanVerdict      string         `json:"scan_verdict" gorm:"type:text"`
	Status           OfferStatus    `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	AlertCleared     bool           `json:"alert_cleared" gorm:"not null;default:false;index"`
	RequestIP        string         `json:"request_ip" gorm:"size:80"`
	RequestUserName  string         `json:"request_user_name" gorm:"size:120"`
	RequestUserEmail string         `json:"request_user_email" gorm:"size:160"`
	UserAgent        string         `json:"user_agent" gorm:"type:text"`
	BrowserData      string         `json:"browser_data" gorm:"type:text"`
	RequestHeaders   string         `json:"request_headers" gorm:"type:text"`
	TelemetryJSON    string         `json:"telemetry_json" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type Ticket struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"size:64;not null;uniqueIndex"`
	EventID     uint           `json:"event_id" gorm:"not null;index"`
	Event       Event          `json:"event" gorm:"foreignKey:EventID"`
	OfferID     *uint          `json:"offer_id" gorm:"index"`
	OwnerID     uint           `json:"owner_id" gorm:"not null;index"`
	Owner       User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Price       float64        `json:"price" gorm:"type:decimal(12,2);not null"`
	SeatInfo    string         `json:"seat_info" gorm:"size:180"`
	Description string         `json:"description" gorm:"type:text"`
	Status      TicketStatus   `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ExchangeRequest struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	OfferID         uint           `json:"offer_id" gorm:"not null;index"`
	Offer           Offer          `json:"offer" gorm:"foreignKey:OfferID"`
	EventID         uint           `json:"event_id" gorm:"not null;index"`
	Event           Event          `json:"event" gorm:"foreignKey:EventID"`
	RequesterID     uint           `json:"requester_id" gorm:"not null;index"`
	Requester       User           `json:"requester" gorm:"foreignKey:RequesterID"`
	ReceiverID      uint           `json:"receiver_id" gorm:"not null;index"`
	Receiver        User           `json:"receiver" gorm:"foreignKey:ReceiverID"`
	OfferedTicketID uint           `json:"offered_ticket_id" gorm:"not null;index"`
	OfferedTicket   Ticket         `json:"offered_ticket" gorm:"foreignKey:OfferedTicketID"`
	TargetTicketID  uint           `json:"target_ticket_id" gorm:"not null;index"`
	TargetTicket    Ticket         `json:"target_ticket" gorm:"foreignKey:TargetTicketID"`
	Message         string         `json:"message" gorm:"type:text"`
	ContactInfo     string         `json:"contact_info" gorm:"type:text"`
	PriceDifference float64        `json:"price_difference" gorm:"type:decimal(12,2);not null;default:0"`
	Status          ExchangeStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidRole(role UserRole) bool {
	switch role {
	case RoleUsuario, RoleAdmin:
		return true
	default:
		return false
	}
}

func NormalizeRole(role UserRole) UserRole {
	if role == RoleAdmin {
		return RoleAdmin
	}
	return RoleUsuario
}

func (e Event) Validate() error {
	if strings.TrimSpace(e.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(e.Location) == "" {
		return errors.New("location is required")
	}
	if e.Capacity <= 0 {
		return errors.New("capacity must be greater than zero")
	}
	if e.DurationMinutes <= 0 {
		return errors.New("duration must be greater than zero")
	}
	if e.EntryPrice <= 0 {
		return errors.New("entry price must be greater than zero")
	}
	return nil
}

func (o Offer) Validate() error {
	if o.EventID == 0 {
		return errors.New("event_id is required")
	}
	if strings.TrimSpace(o.Title) == "" {
		return errors.New("title is required")
	}
	if o.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	if o.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	return nil
}
