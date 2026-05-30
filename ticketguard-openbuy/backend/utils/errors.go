package utils

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrInvalidLogin  = errors.New("invalid credentials")
	ErrUnsafeOffer   = errors.New("offer blocked by security scan")
	ErrInsufficient  = errors.New("not enough available tickets")
	ErrInvalidStatus = errors.New("invalid status for this operation")
)
