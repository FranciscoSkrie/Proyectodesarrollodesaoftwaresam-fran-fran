package services

import (
	"testing"
	"time"

	"ticketguard/backend/domain"
)

func TestMapEventInput_DefaultDuration(t *testing.T) {
	input := EventInput{
		Title:    "Evento",
		Location: "Córdoba",
		StartsAt: time.Now().Add(time.Hour),
		Capacity: 10,
	}
	event := mapEventInput(input)
	if event.DurationMinutes != 120 {
		t.Fatalf("expected default duration 120, got %d", event.DurationMinutes)
	}
}

func TestEventValidate_InvalidCapacity(t *testing.T) {
	event := domain.Event{Title: "Evento", Location: "Córdoba", StartsAt: time.Now(), Capacity: 0, DurationMinutes: 120}
	if err := event.Validate(); err == nil {
		t.Fatal("expected validation error for capacity")
	}
}

func TestMapEventInput_TrimsTextAndDefaultsPrice(t *testing.T) {
	event := mapEventInput(EventInput{Title: "  Festival  ", Location: " Córdoba ", Capacity: 20})
	if event.Title != "Festival" || event.Location != "Córdoba" {
		t.Fatalf("text was not normalized: %#v", event)
	}
	if event.EntryPrice != 15000 {
		t.Fatalf("expected default price 15000, got %v", event.EntryPrice)
	}
}

func TestEventValidate_RequiredFields(t *testing.T) {
	event := domain.Event{Location: "Córdoba", Capacity: 10, DurationMinutes: 120, EntryPrice: 100}
	if err := event.Validate(); err == nil {
		t.Fatal("expected title validation error")
	}
}

func TestEventValidate_ValidEvent(t *testing.T) {
	event := domain.Event{Title: "Festival", Location: "Córdoba", Capacity: 10, DurationMinutes: 120, EntryPrice: 100}
	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid event, got %v", err)
	}
}
