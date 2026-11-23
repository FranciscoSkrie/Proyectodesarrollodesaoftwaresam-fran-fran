package controllers

import (
	"strconv"

	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

type EventController struct{ service *services.EventService }

func NewEventController(service *services.EventService) *EventController {
	return &EventController{service: service}
}

func (h *EventController) List(c *gin.Context) {
	events, err := h.service.List(c.Query("q"), c.Query("category"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, events)
}

func (h *EventController) Get(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	event, err := h.service.Get(id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, event)
}

func (h *EventController) Create(c *gin.Context) {
	var input services.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	event, err := h.service.Create(currentUserID(c), input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, event)
}

func (h *EventController) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	var input services.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	event, err := h.service.Update(id, input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, event)
}

func (h *EventController) Cancel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	if err := h.service.Cancel(id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, gin.H{"message": "event cancelled"})
}

func (h *EventController) Report(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	report, err := h.service.Report(id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, report)
}

func parseID(raw string) (uint, error) {
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func currentUserID(c *gin.Context) uint {
	value, _ := c.Get("userID")
	id, _ := value.(uint)
	return id
}
