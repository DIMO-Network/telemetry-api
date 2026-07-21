package queryRecorder

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog"
)

const testDev = "0x1f9090aaE28b8a3dCeaDf281B0F12828e676c326"

func TestNewReplayLogger(t *testing.T) {
	rl, err := NewReplayLogger("")
	if err != nil {
		t.Fatalf("unexpected error for empty list: %v", err)
	}
	if rl != nil {
		t.Fatal("expected nil ReplayLogger for empty list")
	}

	rl, err = NewReplayLogger(" , ")
	if err != nil {
		t.Fatalf("unexpected error for blank entries: %v", err)
	}
	if rl != nil {
		t.Fatal("expected nil ReplayLogger for blank entries")
	}

	if _, err = NewReplayLogger("not-an-address"); err == nil {
		t.Fatal("expected error for invalid address")
	}

	rl, err = NewReplayLogger(testDev + ", 0x90c4d6113ec88dd4bdf12f26db2b3998fd13a144")
	if err != nil {
		t.Fatalf("unexpected error for valid list: %v", err)
	}
	if len(rl.devs) != 2 {
		t.Fatalf("expected 2 developers, got %d", len(rl.devs))
	}
}

// replayCtx builds a context carrying validated JWT claims, a gqlgen operation
// context, and a logger writing to the returned buffer.
func replayCtx(subject string, tokenID *big.Int) (context.Context, *bytes.Buffer) {
	ctx := context.Background()
	if subject != "" {
		claims := &validator.ValidatedClaims{
			RegisteredClaims: validator.RegisteredClaims{Subject: subject},
			CustomClaims: &auth.TelemetryClaim{
				AssetDID: cloudevent.ERC721DID{TokenID: tokenID},
			},
		}
		ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, claims)
	}
	ctx = graphql.WithOperationContext(ctx, &graphql.OperationContext{
		RawQuery:      "query Test($tokenId: Int!) { signalsLatest(tokenId: $tokenId) { lastSeen } }",
		OperationName: "Test",
		Variables:     map[string]any{"tokenId": 123},
	})

	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	return logger.WithContext(ctx), buf
}

func runInterceptResponse(t *testing.T, rl *ReplayLogger, ctx context.Context) {
	t.Helper()
	resp := rl.InterceptResponse(ctx, func(context.Context) *graphql.Response {
		return &graphql.Response{}
	})
	if resp == nil {
		t.Fatal("expected response to pass through")
	}
}

func TestReplayLoggerRecordsAllowlistedDeveloper(t *testing.T) {
	// Lowercased in config, checksummed in the JWT: matching must be case-insensitive.
	rl, err := NewReplayLogger(testDev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, buf := replayCtx("0x1F9090AAE28B8A3DCEADF281B0F12828E676C326", big.NewInt(123))
	runInterceptResponse(t, rl, ctx)

	if buf.Len() == 0 {
		t.Fatal("expected a log line for allowlisted developer")
	}
	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	if line["message"] != ReplayLogMessage {
		t.Fatalf("expected message %q, got %q", ReplayLogMessage, line["message"])
	}
	if line["operationName"] != "Test" {
		t.Fatalf("expected operationName Test, got %q", line["operationName"])
	}
	if line["query"] == "" || line["query"] == nil {
		t.Fatal("expected raw query in log line")
	}
	if line["vehicleTokenId"] != "123" {
		t.Fatalf("expected vehicleTokenId 123, got %q", line["vehicleTokenId"])
	}
	vars, ok := line["variables"].(map[string]any)
	if !ok {
		t.Fatalf("expected variables object, got %T", line["variables"])
	}
	if vars["tokenId"] != float64(123) {
		t.Fatalf("expected tokenId variable 123, got %v", vars["tokenId"])
	}
}

func TestReplayLoggerSkipsOtherDevelopers(t *testing.T) {
	rl, err := NewReplayLogger(testDev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, buf := replayCtx("0x90c4d6113ec88dd4bdf12f26db2b3998fd13a144", big.NewInt(123))
	runInterceptResponse(t, rl, ctx)
	if buf.Len() != 0 {
		t.Fatalf("expected no log line for non-allowlisted developer, got %s", buf.String())
	}
}

func TestReplayLoggerSkipsUnauthenticatedRequests(t *testing.T) {
	rl, err := NewReplayLogger(testDev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, buf := replayCtx("", nil)
	runInterceptResponse(t, rl, ctx)
	if buf.Len() != 0 {
		t.Fatalf("expected no log line without claims, got %s", buf.String())
	}
}
