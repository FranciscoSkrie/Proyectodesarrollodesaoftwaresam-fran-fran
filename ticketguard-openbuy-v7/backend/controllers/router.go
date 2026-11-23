package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ticketguard/backend/config"
	"ticketguard/backend/domain"
	"ticketguard/backend/middleware"
)

type Controllers struct {
	Auth      *AuthController
	Events    *EventController
	Offers    *OfferController
	Tickets   *TicketController
	Exchanges *ExchangeController
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
	private.POST("/me/password", c.Auth.ChangePassword)

	user := api.Group("")
	user.Use(middleware.Auth(cfg.JWTSecret), middleware.RequireRole(domain.RoleUsuario))
	user.GET("/me/tickets", c.Tickets.ListMine)
	user.POST("/events/:id/buy", c.Tickets.BuyEvent)
	user.POST("/offers/:id/buy", c.Tickets.Buy)
	user.POST("/tickets/:id/cancel", c.Tickets.Cancel)
	user.POST("/tickets/:id/transfer", c.Tickets.Transfer)
	user.GET("/me/offers", c.Offers.ListMine)
	user.POST("/me/offers", c.Offers.Create)
	user.POST("/offers/:id/exchange", c.Exchanges.Request)
	user.GET("/me/exchanges/inbox", c.Exchanges.Inbox)
	user.GET("/me/exchanges/sent", c.Exchanges.Sent)
	user.POST("/exchanges/:id/accept", c.Exchanges.Accept)
	user.POST("/exchanges/:id/reject", c.Exchanges.Reject)

	admin := api.Group("/admin")
	admin.Use(middleware.Auth(cfg.JWTSecret), middleware.RequireRole(domain.RoleAdmin))
	admin.POST("/events", c.Events.Create)
	admin.PUT("/events/:id", c.Events.Update)
	admin.DELETE("/events/:id", c.Events.Cancel)
	admin.GET("/events/:id/report", c.Events.Report)
	admin.GET("/security/offers", c.Offers.ListFlagged)
	admin.POST("/security/alerts/clear", c.Offers.ClearFlaggedAlerts)

	return r
}
