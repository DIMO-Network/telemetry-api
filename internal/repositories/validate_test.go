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
		err := validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{ContainsAny: []string{vss.TagBehaviorHarshAcceleration, vss.TagSafetyCollision}}})
		require.NoError(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{ContainsAll: []string{vss.TagBehaviorHarshAcceleration}}})
		require.NoError(t, err)
	})
	t.Run("invalid tags", func(t *testing.T) {
		err := validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{ContainsAny: []string{"invalid"}}})
		require.Error(t, err)
		err = validateEventArgs(1, validFrom, validTo, &model.EventFilter{Tags: &model.StringArrayFilter{ContainsAll: []string{vss.TagBehaviorHarshAcceleration, "invalid"}}})
		require.Error(t, err)
	})

}

func TestValidateSegmentArgs(t *testing.T) {
	validFrom := time.Now().Add(-time.Hour)
	validTo := time.Now()

	t.Run("valid args", func(t *testing.T) {
		err := validateSegmentArgs(1, validFrom, validTo)
		require.NoError(t, err)
	})

	t.Run("tokenID <= 0", func(t *testing.T) {
		err := validateSegmentArgs(0, validFrom, validTo)
		require.Error(t, err)
	})

	t.Run("from after to", func(t *testing.T) {
		err := validateSegmentArgs(1, validTo.Add(time.Minute), validTo)
		require.Error(t, err)
	})

	t.Run("from equal to", func(t *testing.T) {
		err := validateSegmentArgs(1, validFrom, validFrom)
		require.Error(t, err)
	})

	t.Run("to in future", func(t *testing.T) {
		err := validateSegmentArgs(1, validFrom, time.Now().Add(time.Hour))
		require.ErrorContains(t, err, "to time cannot be in the future")
	})

	t.Run("date range exceeded", func(t *testing.T) {
		from := validTo.Add(-31 * 24 * time.Hour)
		err := validateSegmentArgs(1, from, validTo)
		require.Error(t, err)
	})
}

func TestValidateSegmentConfig(t *testing.T) {
	validConfig := &model.SegmentConfig{}
	otherMechanism := model.DetectionMechanismIgnitionDetection
	idlingMechanism := model.DetectionMechanismStaticRpm

	t.Run("nil config", func(t *testing.T) {
		require.NoError(t, validateSegmentConfig(nil, otherMechanism))
		require.NoError(t, validateSegmentConfig(nil, idlingMechanism))
	})

	t.Run("valid config other mechanism", func(t *testing.T) {
		require.NoError(t, validateSegmentConfig(validConfig, otherMechanism))
	})

	t.Run("valid config staticRpm with idling fields", func(t *testing.T) {
		cfg := &model.SegmentConfig{
			MaxIdleRpm:             ptr(1000),
			SignalCountThreshold:   ptr(5),
		}
		require.NoError(t, validateSegmentConfig(cfg, idlingMechanism))
	})

	t.Run("staticRpm maxIdleRpm out of range", func(t *testing.T) {
		cfg := &model.SegmentConfig{MaxIdleRpm: ptr(100)}
		require.Error(t, validateSegmentConfig(cfg, idlingMechanism))
		cfg.MaxIdleRpm = ptr(4000)
		require.Error(t, validateSegmentConfig(cfg, idlingMechanism))
	})
}

func ptr(i int) *int { return &i }
