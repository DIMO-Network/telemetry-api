package repositories

import (
	"testing"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

func TestValidateEventArgs(t *testing.T) {
	validFrom := time.Now().Add(-time.Hour)
	validTo := time.Now()
	validFilter := &model.EventFilter{}

	t.Run("valid args", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, validTo, validFilter)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("tokenID < 1", func(t *testing.T) {
		err := validateEventArgs(0, validFrom, validTo, validFilter)
		if err == nil {
			t.Error("expected error for tokenID < 1, got nil")
		}
	})

	t.Run("from is zero", func(t *testing.T) {
		err := validateEventArgs(1, time.Time{}, validTo, validFilter)
		if err == nil {
			t.Error("expected error for zero from, got nil")
		}
	})

	t.Run("to is zero", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, time.Time{}, validFilter)
		if err == nil {
			t.Error("expected error for zero to, got nil")
		}
	})

	t.Run("from after to", func(t *testing.T) {
		from := validTo.Add(time.Second)
		err := validateEventArgs(1, from, validTo, validFilter)
		if err == nil {
			t.Error("expected error for from after to, got nil")
		}
	})
}
