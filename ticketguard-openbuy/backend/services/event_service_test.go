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
