package domain

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCliente  UserRole = "cliente"
	RoleVendedor UserRole = "vendedor"
	RoleAdmin    UserRole = "admin"
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

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"size:120;not null"`
	Email        string         `json:"email" gorm:"size:160;not null;uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'cliente'"`
	Balance      float64        `json:"balance" gorm:"type:decimal(12,2);not null;default:0"`
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
	ImageURL        string         `json:"image_url" gorm:"size:500"`
	Status          EventStatus    `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedByID     uint           `json:"created_by_id"`
	CreatedBy       User           `json:"-" gorm:"foreignKey:CreatedByID"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type Offer struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	EventID     uint           `json:"event_id" gorm:"not null;index"`
	Event       Event          `json:"event" gorm:"foreignKey:EventID"`
	SellerID    uint           `json:"seller_id" gorm:"not null;index"`
	Seller      User           `json:"seller" gorm:"foreignKey:SellerID"`
	Title       string         `json:"title" gorm:"size:160;not null"`
	Price       float64        `json:"price" gorm:"type:decimal(12,2);not null"`
	Quantity    int            `json:"quantity" gorm:"not null"`
	ExternalURL string         `json:"external_url" gorm:"size:500"`
	ScanStatus  ScanStatus     `json:"scan_status" gorm:"type:varchar(20);not null;default:'pending'"`
	ScanVerdict string         `json:"scan_verdict" gorm:"size:500"`
	Status      OfferStatus    `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Ticket struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"size:64;not null;uniqueIndex"`
	EventID   uint           `json:"event_id" gorm:"not null;index"`
	Event     Event          `json:"event" gorm:"foreignKey:EventID"`
	OfferID   uint           `json:"offer_id" gorm:"not null;index"`
	Offer     Offer          `json:"offer" gorm:"foreignKey:OfferID"`
	OwnerID   uint           `json:"owner_id" gorm:"not null;index"`
	Owner     User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Price     float64        `json:"price" gorm:"type:decimal(12,2);not null"`
	Status    TicketStatus   `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidRole(role UserRole) bool {
	switch role {
	case RoleCliente, RoleVendedor, RoleAdmin:
		return true
	default:
		return false
	}
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
