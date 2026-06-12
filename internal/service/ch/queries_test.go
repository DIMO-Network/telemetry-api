package ch

import (
	"testing"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stretchr/testify/assert"
)

func TestWithSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   qm.QueryMod
	}{
		{
			name:   "ethr DID extracts address",
			source: "did:ethr:137:0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E",
			want:   qm.Where(sourceWhere, "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"),
		},
		{
			name:   "raw address passed through",
			source: "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E",
			want:   qm.Where(sourceWhere, "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withSource(tt.source)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetLastSeenQuery(t *testing.T) {
	t.Run("nil args returns empty", func(t *testing.T) {
		stmt, args := getLastSeenQuery("subj", nil)
		assert.Empty(t, stmt)
		assert.Nil(t, args)
	})

	t.Run("reads signal_latest", func(t *testing.T) {
		stmt, args := getLastSeenQuery("subj", &model.SignalArgs{})
		want := "SELECT 'lastSeen' AS name, max(timestamp) AS ts, NULL AS value_number, NULL AS value_string, " +
			"CAST(tuple(0, 0, 0, 0), 'Tuple(latitude Float64, longitude Float64, hdop Float64, heading Float64)') AS value_location " +
			"FROM `signal_latest` WHERE (subject = ?) AND (kind = ?);"
		assert.Equal(t, want, stmt)
		assert.Equal(t, []any{"subj", uint8(0)}, args)
	})

	t.Run("source filter applies", func(t *testing.T) {
		stmt, args := getLastSeenQuery("subj", &model.SignalArgs{
			Filter: &model.SignalFilter{Source: ref("0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E")},
		})
		want := "SELECT 'lastSeen' AS name, max(timestamp) AS ts, NULL AS value_number, NULL AS value_string, " +
			"CAST(tuple(0, 0, 0, 0), 'Tuple(latitude Float64, longitude Float64, hdop Float64, heading Float64)') AS value_location " +
			"FROM `signal_latest` WHERE (subject = ?) AND (kind = ?) AND (source = ?);"
		assert.Equal(t, want, stmt)
		assert.Equal(t, []any{"subj", uint8(0), "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"}, args)
	})
}

func TestGetLatestQueriesReadLatestTable(t *testing.T) {
	nonLoc, _ := getLatestNonLocationQuery("subj", []string{"speed"}, nil)
	assert.Contains(t, nonLoc, "FROM `signal_latest`")
	assert.Contains(t, nonLoc, "(kind = ?)")

	loc, _ := getLatestLocationQuery("subj", []string{"currentLocationCoordinates"}, nil)
	assert.Contains(t, loc, "FROM `signal_latest`")
	assert.Contains(t, loc, "(kind = ?)")
	assert.NotContains(t, loc, "argMaxIf") // kind=1 rows are pre-filtered; plain argMax suffices

	all, _ := getAllLatestQuery("subj", nil)
	assert.Contains(t, all, "FROM `signal_latest`")
	assert.Contains(t, all, "(kind = ?)")

	distinct, _ := getDistinctQuery("subj", nil)
	assert.Contains(t, distinct, "FROM `signal_latest`")
	assert.Contains(t, distinct, "(kind = ?)")
}

func TestSummaryQueriesReadSummaryTables(t *testing.T) {
	sig, _ := getSignalSummariesQuery("subj", nil)
	assert.Contains(t, sig, "FROM `signal_summary`")
	assert.Contains(t, sig, "sum(count)")

	ev, _ := getEventSummariesQuery("subj")
	assert.Contains(t, ev, "FROM `event_summary`")
	assert.Contains(t, ev, "sum(count)")
}
