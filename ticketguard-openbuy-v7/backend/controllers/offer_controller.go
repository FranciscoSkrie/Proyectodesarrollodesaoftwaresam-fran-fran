package controllers

import (
	"encoding/json"

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
	offers, err := h.service.ListByPublisher(currentUserID(c))
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
	evidence := services.OfferEvidenceInput{
		RequestIP:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		BrowserData:    c.GetHeader("Sec-CH-UA"),
		RequestHeaders: marshalEvidenceHeaders(c),
	}
	offer, err := h.service.Create(c.Request.Context(), currentUserID(c), input, evidence)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, offer)
}

func (h *OfferController) ListFlagged(c *gin.Context) {
	offers, err := h.service.ListFlagged()
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, offers)
}

func (h *OfferController) ClearFlaggedAlerts(c *gin.Context) {
	if err := h.service.ClearFlaggedAlerts(); err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, gin.H{"message": "alerts cleared"})
}

func marshalEvidenceHeaders(c *gin.Context) string {
	headers := map[string]string{
		"accept_language":       c.GetHeader("Accept-Language"),
		"origin":                c.GetHeader("Origin"),
		"referer":               c.GetHeader("Referer"),
		"x_forwarded_for":       c.GetHeader("X-Forwarded-For"),
		"x_real_ip":             c.GetHeader("X-Real-IP"),
		"sec_ch_ua":             c.GetHeader("Sec-CH-UA"),
		"sec_ch_ua_mobile":      c.GetHeader("Sec-CH-UA-Mobile"),
		"sec_ch_ua_platform":    c.GetHeader("Sec-CH-UA-Platform"),
		"sec_fetch_site":        c.GetHeader("Sec-Fetch-Site"),
		"sec_fetch_mode":        c.GetHeader("Sec-Fetch-Mode"),
		"sec_fetch_destination": c.GetHeader("Sec-Fetch-Dest"),
	}
	raw, err := json.MarshalIndent(headers, "", "  ")
	if err != nil {
		return ""
	}
	return string(raw)
}
