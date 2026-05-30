package main

import (
	"fmt"
	"log"
	"time"

	"ticketguard/backend/clients"
	"ticketguard/backend/config"
	"ticketguard/backend/controllers"
	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/services"
	"ticketguard/backend/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := connectWithRetry(cfg.DSN(), 20, 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrateAndSeed(db); err != nil {
		log.Fatal(err)
	}

	userDAO := dao.NewUserDAO(db)
	eventDAO := dao.NewEventDAO(db)
	offerDAO := dao.NewOfferDAO(db)
	ticketDAO := dao.NewTicketDAO(db)

	scanner := clients.NewLinkScanner(cfg.VTAPIKey)
	authService := services.NewAuthService(userDAO, cfg)
	eventService := services.NewEventService(eventDAO, ticketDAO)
	offerService := services.NewOfferService(offerDAO, eventDAO, scanner)
	ticketService := services.NewTicketService(db, userDAO, eventDAO, offerDAO, ticketDAO)

	router := controllers.SetupRouter(cfg, controllers.Controllers{
		Auth:    controllers.NewAuthController(authService),
		Events:  controllers.NewEventController(eventService),
		Offers:  controllers.NewOfferController(offerService),
		Tickets: controllers.NewTicketController(ticketService),
	})

	addr := ":" + cfg.AppPort
	log.Printf("TicketGuard backend running on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func connectWithRetry(dsn string, attempts int, delay time.Duration) (*gorm.DB, error) {
	var lastErr error
	for i := 1; i <= attempts; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		lastErr = err
		log.Printf("database not ready, retry %d/%d", i, attempts)
		time.Sleep(delay)
	}
	return nil, lastErr
}

func migrateAndSeed(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Offer{}, &domain.Ticket{}); err != nil {
		return err
	}
	return seed(db)
}

func seed(db *gorm.DB) error {
	seedUser := func(name, email, password string, role domain.UserRole) (*domain.User, error) {
		var existing domain.User
		if err := db.Where("email = ?", domain.NormalizeEmail(email)).First(&existing).Error; err == nil {
			return &existing, nil
		}
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user := &domain.User{Name: name, Email: domain.NormalizeEmail(email), PasswordHash: hash, Role: role, Balance: 100000}
		return user, db.Create(user).Error
	}

	admin, err := seedUser("Admin Demo", "admin@ticketguard.test", "Admin123!", domain.RoleAdmin)
	if err != nil {
		return err
	}
	seller, err := seedUser("Vendedor Demo", "seller@ticketguard.test", "Seller123!", domain.RoleVendedor)
	if err != nil {
		return err
	}
	_, err = seedUser("Cliente Demo", "cliente@ticketguard.test", "Cliente123!", domain.RoleCliente)
	if err != nil {
		return err
	}

	var count int64
	db.Model(&domain.Event{}).Count(&count)
	if count == 0 {
		events := []domain.Event{
			{Title: "Festival Córdoba Tech", Description: "Evento de tecnología, música y networking.", Category: "Tecnología", Location: "Córdoba", StartsAt: time.Now().AddDate(0, 1, 0), DurationMinutes: 180, Capacity: 300, ImageURL: "https://images.unsplash.com/photo-1492684223066-81342ee5ff30", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Concierto OpenBuy", Description: "Concierto con entradas publicadas por múltiples vendedores.", Category: "Música", Location: "Buenos Aires", StartsAt: time.Now().AddDate(0, 2, 0), DurationMinutes: 120, Capacity: 500, ImageURL: "https://images.unsplash.com/photo-1501386761578-eac5c94b800a", Status: domain.EventActive, CreatedByID: admin.ID},
		}
		for i := range events {
			if err := db.Create(&events[i]).Error; err != nil {
				return err
			}
		}
		var first domain.Event
		if err := db.First(&first).Error; err == nil {
			offer := domain.Offer{EventID: first.ID, SellerID: seller.ID, Title: fmt.Sprintf("Entrada general - %s", first.Title), Price: 15000, Quantity: 20, ExternalURL: "https://tickets.example.com/evento-seguro", ScanStatus: domain.ScanSafe, ScanVerdict: "seed: link de demostración seguro", Status: domain.OfferActive}
			return db.Create(&offer).Error
		}
	}
	return nil
}
