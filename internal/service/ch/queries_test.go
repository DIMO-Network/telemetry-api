package ch

import (
	"testing"

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
