package e2e_test

import (
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVINVCLatest(t *testing.T) {
	const testVC = `{"id":"2oYEQJDQFRYCI1UIf6QTy92MS3G","source":"0x0000000000000000000000000000000000000000","producer":"did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA_31653","specversion":"1.0","subject":"did:nft:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_39718","time":"2024-11-08T03:54:51Z","type":"dimo.verifiablecredential","datacontenttype":"application/json","dataversion":"VINVCv1.0","data":{"@context":["https://www.w3.org/ns/credentials/v2",{"vehicleIdentificationNumber":"https://schema.org/vehicleIdentificationNumber"},"https://attestation-api.dimo.zone/v1/vc/context"],"id":"urn:uuid:0b74cd3b-5998-4436-ada3-22dd6cfe2b3c","type":["VerifiableCredential","Vehicle"],"issuer":"https://attestation-api.dimo.zone/v1/vc/keys","validFrom":"2024-11-08T03:54:51Z","validTo":"2024-11-10T00:00:00Z","credentialSubject":{"id":"did:nft:137_erc721:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_39718","vehicleTokenId":39718,"vehicleContractAddress":"eth:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF","vehicleIdentificationNumber":"3eMCZ5AN5PM647548","recordedBy":"did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA_31653","recordedAt":"2024-11-08T00:40:29Z"},"credentialStatus":{"id":"https://attestation-api.dimo.zone/v1/vc/status/39718","type":"BitstringStatusListEntry","statusPurpose":"revocation","statusListIndex":0,"statusListCredential":"https://attestation-api.dimo.zone/v1/vc/status"},"proof":{"type":"DataIntegrityProof","cryptosuite":"ecdsa-rdfc-2019","verificationMethod":"https://attestation-api.dimo.zone/v1/vc/keys#key1","created":"2024-11-08T03:54:51Z","proofPurpose":"assertionMethod","proofValue":"381yXZFRShe2rrr9A3VFGeXS9izouz7Gor1GTb6Mwjkpge8eEn814QivzssEogoLKrzGN6WPKWBQrFLgfcsUhuaAWhS421Dn"}}}`
	testEvent := cloudevent.CloudEvent[json.RawMessage]{}
	err := json.Unmarshal([]byte(testVC), &testEvent)
	require.NoError(t, err)
	services := GetTestServices(t)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	services.FetchServer.SetCloudEventReturn(testEvent)

	// Create auth token
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetVINCredential})

	query := `
	query VIN {
		vinVCLatest(tokenId: 39718) {
			vehicleTokenId
			vin
			recordedBy
			recordedAt
			countryCode
			vehicleContractAddress
			validFrom
			validTo
			rawVC
		}
	}`

	var result struct {
		VINVC struct {
			VehicleTokenID         int    `json:"vehicleTokenId"`
			VIN                    string `json:"vin"`
			RecordedBy             string `json:"recordedBy"`
			RecordedAt             string `json:"recordedAt"`
			CountryCode            string `json:"countryCode"`
			VehicleContractAddress string `json:"vehicleContractAddress"`
			ValidFrom              string `json:"validFrom"`
			ValidTo                string `json:"validTo"`
			RawVC                  string `json:"rawVC"`
		} `json:"vinVCLatest"`
	}

	err = telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	actual := result.VINVC
	expectedJSON, err := json.Marshal(testVC)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(result.VINVC.RawVC)
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
	assert.Equal(t, 39718, actual.VehicleTokenID)
	assert.Equal(t, "3eMCZ5AN5PM647548", actual.VIN)
	assert.Equal(t, "did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA_31653", actual.RecordedBy)
	assert.Equal(t, "2024-11-08T00:40:29Z", actual.RecordedAt)
	assert.Equal(t, "eth:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF", actual.VehicleContractAddress)
	assert.Equal(t, "2024-11-08T03:54:51Z", actual.ValidFrom)
	assert.Equal(t, "2024-11-10T00:00:00Z", actual.ValidTo)
}
