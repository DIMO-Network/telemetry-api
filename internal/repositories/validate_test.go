package repositories

import (
	"testing"
	"time"

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

}

func TestValidateSegmentArgs(t *testing.T) {
	validFrom := time.Now().Add(-time.Hour)
	validTo := time.Now()

	t.Run("valid args", func(t *testing.T) {
		err := validateSegmentArgs(1, validFrom, validTo)
		require.NoError(t, err)
	})

	t.Run("exactly 31 days passes", func(t *testing.T) {
		from := validTo.Add(-31 * 24 * time.Hour)
		err := validateSegmentArgs(1, from, validTo)
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

	t.Run("date range exceeded", func(t *testing.T) {
		from := validTo.Add(-33 * 24 * time.Hour) // max is 32 days
		err := validateSegmentArgs(1, from, validTo)
		require.Error(t, err)
	})
}

// TestValidateSegmentDateRange exercises the shared date-range rule used by both segments and dailyActivity.
func TestValidateSegmentDateRange(t *testing.T) {
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	t.Run("short range passes", func(t *testing.T) {
		from := to.Add(-time.Hour)
		require.NoError(t, validateSegmentDateRange(from, to))
	})

	t.Run("exactly 31 days passes", func(t *testing.T) {
		from := to.Add(-31 * 24 * time.Hour)
		require.NoError(t, validateSegmentDateRange(from, to))
	})

	t.Run("31 days plus 1 second passes", func(t *testing.T) {
		from := to.Add(-31*24*time.Hour - time.Second)
		require.NoError(t, validateSegmentDateRange(from, to))
	})

	t.Run("32 days plus 2 seconds fails", func(t *testing.T) {
		from := to.Add(-32*24*time.Hour - 2*time.Second)
		require.Error(t, validateSegmentDateRange(from, to))
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

func TestValidateFilter(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		require.NoError(t, validateFilter(nil))
	})

	t.Run("valid ethr DID", func(t *testing.T) {
		source := "did:ethr:137:0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"
		err := validateFilter(&model.SignalFilter{Source: &source})
		require.NoError(t, err)
	})

	t.Run("invalid source string", func(t *testing.T) {
		source := "smartcar"
		err := validateFilter(&model.SignalFilter{Source: &source})
		require.Error(t, err)
		require.ErrorAs(t, err, new(ValidationError))
	})

	t.Run("raw address rejected", func(t *testing.T) {
		source := "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"
		err := validateFilter(&model.SignalFilter{Source: &source})
		require.Error(t, err)
		require.ErrorAs(t, err, new(ValidationError))
	})

	t.Run("nil source", func(t *testing.T) {
		err := validateFilter(&model.SignalFilter{Source: nil})
		require.NoError(t, err)
	})
}

func strRef(s string) *string { return &s }

func TestValidateEventNameFilter(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(nil))
	})

	t.Run("valid 2-segment eq", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{Eq: strRef("behavior.harshBraking")}))
	})

	t.Run("valid 2-segment in", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{In: []string{"behavior.harshBraking", "safety.collision"}}))
	})

	t.Run("invalid eq - no dot", func(t *testing.T) {
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{Eq: strRef("harshBraking")}))
	})

	t.Run("invalid eq - empty segment", func(t *testing.T) {
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{Eq: strRef("behavior.")}))
	})

	t.Run("invalid eq - 3 segments", func(t *testing.T) {
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{Eq: strRef("a.b.c")}))
	})

	t.Run("invalid in - mixed", func(t *testing.T) {
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{In: []string{"behavior.harshBraking", "invalid"}}))
	})

	t.Run("valid startsWith - category with dot", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{StartsWith: strRef("behavior.")}))
	})

	t.Run("valid startsWith - partial name", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{StartsWith: strRef("behavior.harsh")}))
	})

	t.Run("invalid startsWith - no dot", func(t *testing.T) {
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{StartsWith: strRef("behavior")}))
	})

	t.Run("neq is NOT validated", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{Neq: strRef("anything-goes")}))
	})

	t.Run("or recursion", func(t *testing.T) {
		require.NoError(t, validateEventNameFilter(&model.StringValueFilter{
			Or: []*model.StringValueFilter{
				{Eq: strRef("behavior.harshBraking")},
				{StartsWith: strRef("safety.")},
			},
		}))
		require.Error(t, validateEventNameFilter(&model.StringValueFilter{
			Or: []*model.StringValueFilter{
				{Eq: strRef("invalid")},
			},
		}))
	})
}

func ptr(i int) *int { return &i }
