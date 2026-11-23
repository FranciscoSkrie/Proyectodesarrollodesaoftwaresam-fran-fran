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
	exchangeDAO := dao.NewExchangeDAO(db)

	scanner := clients.NewLinkScanner(cfg.VTAPIKey)
	authService := services.NewAuthService(userDAO, cfg)
	eventService := services.NewEventService(eventDAO, ticketDAO)
	offerService := services.NewOfferService(offerDAO, eventDAO, userDAO, ticketDAO, scanner)
	ticketService := services.NewTicketService(db, userDAO, eventDAO, offerDAO, ticketDAO)
	exchangeService := services.NewExchangeService(db, exchangeDAO, offerDAO, ticketDAO, userDAO)

	router := controllers.SetupRouter(cfg, controllers.Controllers{
		Auth:      controllers.NewAuthController(authService),
		Events:    controllers.NewEventController(eventService),
		Offers:    controllers.NewOfferController(offerService),
		Tickets:   controllers.NewTicketController(ticketService),
		Exchanges: controllers.NewExchangeController(exchangeService),
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
	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}, &domain.Offer{}, &domain.ExchangeRequest{}); err != nil {
		return err
	}
	if err := db.Model(&domain.User{}).
		Where("role IN ?", []string{"cliente", "vendedor"}).
		Update("role", domain.RoleUsuario).Error; err != nil {
		return err
	}
	return seed(db)
}

func seed(db *gorm.DB) error {
	seedUser := func(name, email, password string, role domain.UserRole) (*domain.User, error) {
		var existing domain.User
		if err := db.Where("email = ?", domain.NormalizeEmail(email)).First(&existing).Error; err == nil {
			if existing.Role != role && existing.Role != domain.RoleAdmin {
				existing.Role = domain.NormalizeRole(existing.Role)
				_ = db.Save(&existing).Error
			}
			return &existing, nil
		}
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user := &domain.User{Name: name, Email: domain.NormalizeEmail(email), PasswordHash: hash, Role: role, Balance: 0}
		return user, db.Create(user).Error
	}

	admin, err := seedUser("Admin Demo", "admin@ticketguard.test", "Admin123!", domain.RoleAdmin)
	if err != nil {
		return err
	}
	usuario, err := seedUser("Usuario Demo", "usuario@ticketguard.test", "Usuario123!", domain.RoleUsuario)
	if err != nil {
		return err
	}
	_, err = seedUser("Usuario Intercambio", "intercambio@ticketguard.test", "Usuario123!", domain.RoleUsuario)
	if err != nil {
		return err
	}

	var count int64
	db.Model(&domain.Event{}).Count(&count)
	if count == 0 {
		events := []domain.Event{
			{Title: "Festival Córdoba Tech", Description: "Foro de publicaciones para entradas del festival. Los usuarios pueden comprar una entrada general desde el foro y solicitar intercambios en las publicaciones.", Category: "Tecnología", Location: "Córdoba", StartsAt: time.Now().AddDate(0, 1, 0), DurationMinutes: 180, Capacity: 300, EntryPrice: 15000, ImageURL: "https://images.unsplash.com/photo-1492684223066-81342ee5ff30", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Concierto OpenBuy", Description: "Foro de reventa e intercambio de entradas. Cada publicación queda registrada como evidencia del honeypot.", Category: "Música", Location: "Buenos Aires", StartsAt: time.Now().AddDate(0, 2, 0), DurationMinutes: 120, Capacity: 500, EntryPrice: 22000, ImageURL: "https://images.unsplash.com/photo-1501386761578-eac5c94b800a", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Final Deportiva Federal", Description: "Foro para publicaciones deportivas, enlaces externos y trueques entre usuarios.", Category: "Deportes", Location: "Córdoba", StartsAt: time.Now().AddDate(0, 1, 10), DurationMinutes: 150, Capacity: 750, EntryPrice: 18000, ImageURL: "https://images.unsplash.com/photo-1540747913346-19e32dc3e97e", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Foro Teatro Centro", Description: "Publicaciones de entradas para obras y funciones. El sistema analiza los enlaces pegados por usuarios.", Category: "Teatro", Location: "Rosario", StartsAt: time.Now().AddDate(0, 1, 20), DurationMinutes: 100, Capacity: 240, EntryPrice: 12000, ImageURL: "https://images.unsplash.com/photo-1503095396549-807759245b35", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Gaming & eSports Arena", Description: "Foro de entradas para torneos gaming. Útil para simular intentos de phishing sobre cuentas de juegos.", Category: "Gaming", Location: "Córdoba", StartsAt: time.Now().AddDate(0, 3, 0), DurationMinutes: 240, Capacity: 600, EntryPrice: 23000, ImageURL: "https://images.unsplash.com/photo-1542751371-adc38448a05e", Status: domain.EventActive, CreatedByID: admin.ID},
			{Title: "Congreso Seguridad Digital", Description: "Foro académico para pruebas del flujo completo: publicaciones, links sospechosos, alertas y reportes.", Category: "Conferencias", Location: "Mendoza", StartsAt: time.Now().AddDate(0, 2, 15), DurationMinutes: 300, Capacity: 400, EntryPrice: 20000, ImageURL: "https://images.unsplash.com/photo-1515187029135-18ee286d815b", Status: domain.EventActive, CreatedByID: admin.ID},
		}
		for i := range events {
			if err := db.Create(&events[i]).Error; err != nil {
				return err
			}
		}
		var seededEvents []domain.Event
		if err := db.Find(&seededEvents).Error; err != nil {
			return err
		}
		for _, event := range seededEvents {
			publisherTicket := domain.Ticket{Code: fmt.Sprintf("TG-SEED-%d-%d", usuario.ID, event.ID), EventID: event.ID, OwnerID: usuario.ID, Price: event.EntryPrice, SeatInfo: "Sector A - Fila 4", Description: "Entrada semilla asociada a publicación", Status: domain.TicketActive}
			if err := db.Create(&publisherTicket).Error; err != nil {
				return err
			}
			ticketID := publisherTicket.ID
			offer := domain.Offer{EventID: event.ID, PublisherID: usuario.ID, TicketID: &ticketID, Title: fmt.Sprintf("Intercambio sector A - %s", event.Title), Price: event.EntryPrice, Quantity: 1, ExternalURL: "https://tickets.example.com/evento-seguro", ContactInfo: usuario.Email, ExchangeDetails: "Sector A, fila 4. Acepto cambios por sector similar o mejor ubicación.", ScanStatus: domain.ScanSafe, ScanVerdict: "seed: link de demostración seguro", Status: domain.OfferActive, RequestIP: "127.0.0.1", RequestUserName: usuario.Name, RequestUserEmail: usuario.Email, UserAgent: "seed/local", BrowserData: "seed/local", RequestHeaders: "seed/local", TelemetryJSON: "seed/local"}
			if err := db.Create(&offer).Error; err != nil {
				return err
			}

			if event.Category == "Gaming" {
				suspicious := domain.Offer{EventID: event.ID, PublisherID: usuario.ID, Title: "Entrada premium con link acortado", Price: event.EntryPrice + 5000, Quantity: 1, ExternalURL: "https://bit.ly/ticketguard-demo-oferta", ContactInfo: usuario.Email, ExchangeDetails: "Publicación semilla sospechosa", ScanStatus: domain.ScanSuspicious, ScanVerdict: "seed: link acortado marcado como sospechoso para probar alertas", Status: domain.OfferActive, RequestIP: "127.0.0.1", RequestUserName: usuario.Name, RequestUserEmail: usuario.Email, UserAgent: "seed/local", BrowserData: "seed/local", RequestHeaders: "seed/local", TelemetryJSON: "seed/local"}
				if err := db.Create(&suspicious).Error; err != nil {
					return err
				}
			}

			if event.Category == "Conferencias" {
				blocked := domain.Offer{EventID: event.ID, PublisherID: usuario.ID, Title: "Oferta bloqueada por phishing demo", Price: event.EntryPrice, Quantity: 1, ExternalURL: "https://loginform-phish.example.com/credenciales", ContactInfo: usuario.Email, ExchangeDetails: "Publicación bloqueada de prueba", ScanStatus: domain.ScanMalicious, ScanVerdict: "seed: indicador compatible con phishing; publicación bloqueada para compradores", Status: domain.OfferBlocked, RequestIP: "127.0.0.1", RequestUserName: usuario.Name, RequestUserEmail: usuario.Email, UserAgent: "seed/local", BrowserData: "seed/local", RequestHeaders: "seed/local", TelemetryJSON: "seed/local"}
				if err := db.Create(&blocked).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
