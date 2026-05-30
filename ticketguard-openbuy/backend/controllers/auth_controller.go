package controllers

import (
	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct{ service *services.AuthService }

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

func (h *AuthController) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	resp, err := h.service.Register(input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Created(c, resp)
}

func (h *AuthController) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, utils.ErrInvalidInput)
		return
	}
	resp, err := h.service.Login(input)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.OK(c, resp)
}
