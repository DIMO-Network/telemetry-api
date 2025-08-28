package repositories

import (
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

func TestValidateEventArgs(t *testing.T) {
	validFrom := time.Now().Add(-time.Hour)
	validTo := time.Now()
	validFilter := &model.EventFilter{}

	t.Run("valid args", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, validTo, validFilter)
		require.NoError(t, err)
	})

	t.Run("tokenID < 1", func(t *testing.T) {
		err := validateEventArgs(0, validFrom, validTo, validFilter)
		require.Error(t, err)
	})

	t.Run("from is zero", func(t *testing.T) {
		err := validateEventArgs(1, time.Time{}, validTo, validFilter)
		require.Error(t, err)
	})

	t.Run("to is zero", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, time.Time{}, validFilter)
		require.Error(t, err)
	})

	t.Run("from after to", func(t *testing.T) {
		from := validTo.Add(time.Second)
		err := validateEventArgs(1, from, validTo, validFilter)
		require.Error(t, err)
	})

	t.Run("valid tags", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasAny: []string{vss.TagBehaviorHarshAcceleration, vss.TagSafetyCollision}}})
		require.NoError(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasAll: []string{vss.TagBehaviorHarshAcceleration}}})
		require.NoError(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasNone: []string{vss.TagBehaviorHarshAcceleration}}})
		require.NoError(t, err)
	})
	t.Run("invalid tags", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasAny: []string{"invalid"}}})
		require.Error(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasAll: []string{vss.TagBehaviorHarshAcceleration, "invalid"}}})
		require.Error(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{HasNone: []string{"invalid"}}})
		require.Error(t, err)
	})

}
