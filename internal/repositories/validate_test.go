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
		from := validTo.Add(-32 * 24 * time.Hour) // max is 31 days
		err := validateSegmentArgs(1, from, validTo)
		require.Error(t, err)
	})
}

func TestValidateSegmentConfig(t *testing.T) {
	validConfig := &model.SegmentConfig{}
	otherMechanism := model.DetectionMechanismIgnitionDetection
	idlingMechanism := model.DetectionMechanismIdling

	t.Run("nil config", func(t *testing.T) {
		require.NoError(t, validateSegmentConfig(nil, otherMechanism))
		require.NoError(t, validateSegmentConfig(nil, idlingMechanism))
	})

	t.Run("valid config other mechanism", func(t *testing.T) {
		require.NoError(t, validateSegmentConfig(validConfig, otherMechanism))
	})

	t.Run("valid config idling with idling fields", func(t *testing.T) {
		cfg := &model.SegmentConfig{
			MaxIdleRpm:           ptr(1000),
			SignalCountThreshold: ptr(5),
		}
		require.NoError(t, validateSegmentConfig(cfg, idlingMechanism))
	})

	t.Run("idling maxIdleRpm out of range", func(t *testing.T) {
		cfg := &model.SegmentConfig{MaxIdleRpm: ptr(100)}
		require.Error(t, validateSegmentConfig(cfg, idlingMechanism))
		cfg.MaxIdleRpm = ptr(4000)
		require.Error(t, validateSegmentConfig(cfg, idlingMechanism))
	})

	refuelMechanism := model.DetectionMechanismRefuel
	rechargeMechanism := model.DetectionMechanismRecharge
	t.Run("valid config refuel", func(t *testing.T) {
		cfg := &model.SegmentConfig{MinIncreasePercent: ptr(15)}
		require.NoError(t, validateSegmentConfig(cfg, refuelMechanism))
	})
	t.Run("valid config recharge", func(t *testing.T) {
		cfg := &model.SegmentConfig{MinIncreasePercent: ptr(20)}
		require.NoError(t, validateSegmentConfig(cfg, rechargeMechanism))
	})
	t.Run("refuel/recharge minIncreasePercent out of range", func(t *testing.T) {
		require.Error(t, validateSegmentConfig(&model.SegmentConfig{MinIncreasePercent: ptr(0)}, refuelMechanism))
		require.Error(t, validateSegmentConfig(&model.SegmentConfig{MinIncreasePercent: ptr(101)}, rechargeMechanism))
	})
}

func TestValidateSegmentLimit(t *testing.T) {
	t.Run("nil limit", func(t *testing.T) {
		require.NoError(t, validateSegmentLimit(nil))
	})
	t.Run("valid limit", func(t *testing.T) {
		require.NoError(t, validateSegmentLimit(ptr(1)))
		require.NoError(t, validateSegmentLimit(ptr(100)))
		require.NoError(t, validateSegmentLimit(ptr(200)))
	})
	t.Run("limit too low", func(t *testing.T) {
		require.Error(t, validateSegmentLimit(ptr(0)))
	})
	t.Run("limit too high", func(t *testing.T) {
		require.Error(t, validateSegmentLimit(ptr(201)))
	})
}

func ptr(i int) *int { return &i }
