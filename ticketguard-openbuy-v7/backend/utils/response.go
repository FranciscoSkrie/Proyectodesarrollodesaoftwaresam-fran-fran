package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func Error(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal error"

	switch {
	case errors.Is(err, ErrInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, ErrInvalidLogin), errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
		message = "not found"
	case errors.Is(err, ErrConflict), errors.Is(err, ErrInsufficient), errors.Is(err, ErrUnsafeOffer), errors.Is(err, ErrInvalidStatus):
		status = http.StatusConflict
		message = err.Error()
	default:
		if err != nil && err.Error() != "" {
			message = err.Error()
		}
	}

	c.JSON(status, gin.H{"error": message})
}
