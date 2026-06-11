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

	t.Run("aggregates at projection grain", func(t *testing.T) {
		stmt, args := getLastSeenQuery("subj", &model.SignalArgs{})
		want := "SELECT 'lastSeen' AS name, max(ts) AS ts, NULL AS value_number, NULL AS value_string, " +
			"CAST(tuple(0, 0, 0, 0), 'Tuple(latitude Float64, longitude Float64, hdop Float64, heading Float64)') AS value_location " +
			"FROM (SELECT max(timestamp) AS ts FROM `signal` WHERE (subject = ?) GROUP BY name);"
		assert.Equal(t, want, stmt)
		assert.Equal(t, []any{"subj"}, args)
	})

	t.Run("source filter stays in inner query", func(t *testing.T) {
		stmt, args := getLastSeenQuery("subj", &model.SignalArgs{
			Filter: &model.SignalFilter{Source: ref("0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E")},
		})
		want := "SELECT 'lastSeen' AS name, max(ts) AS ts, NULL AS value_number, NULL AS value_string, " +
			"CAST(tuple(0, 0, 0, 0), 'Tuple(latitude Float64, longitude Float64, hdop Float64, heading Float64)') AS value_location " +
			"FROM (SELECT max(timestamp) AS ts FROM `signal` WHERE (subject = ?) AND (source = ?) GROUP BY name);"
		assert.Equal(t, want, stmt)
		assert.Equal(t, []any{"subj", "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"}, args)
	})
}
