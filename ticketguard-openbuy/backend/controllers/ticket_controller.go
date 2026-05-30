package controllers

import (
	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

type TicketController struct{ service *services.TicketService }

func NewTicketController(service *services.TicketService) *TicketController {
	return &TicketController{service: service}
}

func (h *TicketController) Buy(c *gin.Context) {
	offerID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	ticket, err := h.service.Buy(currentUserID(c), offerID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, ticket)
}

func (h *TicketController) ListMine(c *gin.Context) {
	tickets, err := h.service.ListMine(currentUserID(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, tickets)
}

func (h *TicketController) Cancel(c *gin.Context) {
	ticketID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	if err := h.service.Cancel(currentUserID(c), ticketID); err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, gin.H{"message": "ticket cancelled"})
}

func (h *TicketController) Transfer(c *gin.Context) {
	ticketID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	var input services.TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	ticket, err := h.service.Transfer(currentUserID(c), ticketID, input.Email)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, ticket)
}
