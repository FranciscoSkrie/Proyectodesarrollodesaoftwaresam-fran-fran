package controllers

import (
	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

type ExchangeController struct{ service *services.ExchangeService }

func NewExchangeController(service *services.ExchangeService) *ExchangeController {
	return &ExchangeController{service: service}
}

func (h *ExchangeController) Request(c *gin.Context) {
	offerID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	var input services.ExchangeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	exchange, err := h.service.Request(currentUserID(c), offerID, input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, exchange)
}

func (h *ExchangeController) Inbox(c *gin.Context) {
	exchanges, err := h.service.ListInbox(currentUserID(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, exchanges)
}

func (h *ExchangeController) Sent(c *gin.Context) {
	exchanges, err := h.service.ListSent(currentUserID(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, exchanges)
}

func (h *ExchangeController) Accept(c *gin.Context) {
	exchangeID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	exchange, err := h.service.Accept(currentUserID(c), exchangeID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, exchange)
}

func (h *ExchangeController) Reject(c *gin.Context) {
	exchangeID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	exchange, err := h.service.Reject(currentUserID(c), exchangeID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, exchange)
}
