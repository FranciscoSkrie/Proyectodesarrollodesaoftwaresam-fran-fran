package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ticketguard/backend/config"
	"ticketguard/backend/domain"
	"ticketguard/backend/middleware"
)

type Controllers struct {
	Auth    *AuthController
	Events  *EventController
	Offers  *OfferController
	Tickets *TicketController
}

func SetupRouter(cfg config.Config, c Controllers) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	r.GET("/health", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	api.POST("/auth/register", c.Auth.Register)
	api.POST("/auth/login", c.Auth.Login)
	api.GET("/events", c.Events.List)
	api.GET("/events/:id", c.Events.Get)
	api.GET("/events/:id/offers", c.Offers.ListForEvent)

	private := api.Group("")
	private.Use(middleware.Auth(cfg.JWTSecret))
	private.POST("/offers/:id/buy", c.Tickets.Buy)
	private.GET("/me/tickets", c.Tickets.ListMine)
	private.POST("/tickets/:id/cancel", c.Tickets.Cancel)
	private.POST("/tickets/:id/transfer", c.Tickets.Transfer)

	seller := api.Group("/seller")
	seller.Use(middleware.Auth(cfg.JWTSecret), middleware.RequireRole(domain.RoleVendedor, domain.RoleAdmin))
	seller.GET("/offers", c.Offers.ListMine)
	seller.POST("/offers", c.Offers.Create)

	admin := api.Group("/admin")
	admin.Use(middleware.Auth(cfg.JWTSecret), middleware.RequireRole(domain.RoleAdmin))
	admin.POST("/events", c.Events.Create)
	admin.PUT("/events/:id", c.Events.Update)
	admin.DELETE("/events/:id", c.Events.Cancel)
	admin.GET("/events/:id/report", c.Events.Report)

	return r
}
