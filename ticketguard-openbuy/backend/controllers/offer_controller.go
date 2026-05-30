package controllers

import (
	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

type OfferController struct{ service *services.OfferService }

func NewOfferController(service *services.OfferService) *OfferController {
	return &OfferController{service: service}
}

func (h *OfferController) ListForEvent(c *gin.Context) {
	eventID, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	offers, err := h.service.ListForEvent(eventID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, offers)
}

func (h *OfferController) ListMine(c *gin.Context) {
	offers, err := h.service.ListBySeller(currentUserID(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, offers)
}

func (h *OfferController) Create(c *gin.Context) {
	var input services.OfferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	offer, err := h.service.Create(c.Request.Context(), currentUserID(c), input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, offer)
}
