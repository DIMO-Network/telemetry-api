package e2e_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-jose/go-jose/v4"
)

type mockAuthServer struct {
	server                      *httptest.Server
	signer                      jose.Signer
	jwks                        jose.JSONWebKey
	defaultClaims               map[string]any
	VehicleContractAddress      string
	ManufacturerContractAddress string
}

func setupAuthServer(t *testing.T, vehicleContractAddress, manufacturerContractAddress common.Address) *mockAuthServer {
	t.Helper()

	// Generate RSA key
	sk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Generate key ID
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("Failed to generate key ID: %v", err)
	}
	keyID := hex.EncodeToString(b)

	// Create JWK
	jwk := jose.JSONWebKey{
		Key:       sk.Public(),
		KeyID:     keyID,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	// Create signer
	sig, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       sk,
	}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{
			"kid": keyID,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	defaultClaims := map[string]any{
		"aud": []string{
			"dimo.zone",
		},
		"exp": 9722006230,
		"iat": 1721833430,
		"iss": "http://127.0.0.1:3003",
		"sub": "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF/39718",
	}

	auth := &mockAuthServer{
		signer:                      sig,
		jwks:                        jwk,
		defaultClaims:               defaultClaims,
		VehicleContractAddress:      vehicleContractAddress.String(),
		ManufacturerContractAddress: manufacturerContractAddress.String(),
	}

	// Create test server with only JWKS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/keys" {
			http.NotFound(w, r)
			return
		}
		err := json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{jwk},
		})
		if err != nil {
			http.Error(w, "Failed to encode JWKS", http.StatusInternalServerError)
		}
	}))

	auth.server = server
	return auth
}

func (m *mockAuthServer) sign(claims map[string]interface{}) (string, error) {
	b, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	out, err := m.signer.Sign(b)
	if err != nil {
		return "", fmt.Errorf("failed to sign claims: %w", err)
	}

	token, err := out.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize token: %w", err)
	}

	return token, nil
}

func (m *mockAuthServer) CreateToken(t *testing.T, customClaims map[string]interface{}, privileges []int) string {
	t.Helper()

	claims := make(map[string]interface{})
	for k, v := range m.defaultClaims {
		claims[k] = v
	}
	for k, v := range customClaims {
		claims[k] = v
	}

	// Add privileges if provided
	if privileges != nil {
		claims["privilege_ids"] = privileges
	}

	token, err := m.sign(claims)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	return token
}

func (m *mockAuthServer) URL() string {
	return m.server.URL
}

func (m *mockAuthServer) Close() {
	m.server.Close()
}

// Helper function to create test tokens with specific claims and privileges
func (m *mockAuthServer) CreateVehicleToken(t *testing.T, tokenID string, privileges []int) string {
	return m.CreateToken(t, map[string]interface{}{
		"token_id":         tokenID,
		"contract_address": m.VehicleContractAddress,
	}, privileges)
}

func (m *mockAuthServer) CreateManufacturerToken(t *testing.T, tokenID string, privileges []int) string {
	return m.CreateToken(t, map[string]interface{}{
		"contract_address": m.ManufacturerContractAddress,
		"token_id":         tokenID,
	}, privileges)
}

func (m *mockAuthServer) CreateUserToken(t *testing.T, userID string, privileges []int) string {
	return m.CreateToken(t, map[string]interface{}{
		"sub": userID,
	}, privileges)
}
