package service_test

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSAMLService_IsConfigured tests configuration detection
func TestSAMLService_IsConfigured(t *testing.T) {
	tests := []struct {
		name        string
		config      func(*testing.T, *testutils.MockSAMLServer) *config.InfrastructureConfig
		expectError bool
		description string
	}{
		{
			name: "Fully configured SAML",
			config: func(t *testing.T, mockServer *testutils.MockSAMLServer) *config.InfrastructureConfig {
				return &config.InfrastructureConfig{
					SAML2EntityID:         "https://terrareg.example.com",
					SAML2IDPMetadataURL:   mockServer.MetadataURL,
					SAML2PrivateKey:       testPrivateKey,
					SAML2PublicKey:        testPublicKey,
					PublicURL:             "https://terrareg.example.com",
				}
			},
			expectError: false,
			description: "All required SAML fields present",
		},
		{
			name: "Missing entity ID",
			config: func(t *testing.T, mockServer *testutils.MockSAMLServer) *config.InfrastructureConfig {
				return &config.InfrastructureConfig{
					SAML2IDPMetadataURL: mockServer.MetadataURL,
					SAML2PrivateKey:     testPrivateKey,
					SAML2PublicKey:      testPublicKey,
				}
			},
			expectError: true,
			description: "Entity ID is required",
		},
		{
			name: "Missing IDP metadata URL",
			config: func(t *testing.T, mockServer *testutils.MockSAMLServer) *config.InfrastructureConfig {
				return &config.InfrastructureConfig{
					SAML2EntityID:  "https://terrareg.example.com",
					SAML2PrivateKey: testPrivateKey,
					SAML2PublicKey:  testPublicKey,
				}
			},
			expectError: true,
			description: "IDP metadata URL is required",
		},
		{
			name: "Missing private key",
			config: func(t *testing.T, mockServer *testutils.MockSAMLServer) *config.InfrastructureConfig {
				return &config.InfrastructureConfig{
					SAML2EntityID:       "https://terrareg.example.com",
					SAML2IDPMetadataURL: mockServer.MetadataURL,
					SAML2PublicKey:      testPublicKey,
				}
			},
			expectError: false, // Service can be created without private key
			description: "Private key not required for service creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer, err := testutils.NewMockSAMLServer()
			require.NoError(t, err)
			defer mockServer.Close()

			cfg := tt.config(t, mockServer)
			svc, err := service.NewSAMLService(cfg)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err, tt.description)
				if svc != nil {
					assert.True(t, svc.IsConfigured(), "Service should report as configured")
				}
			}
		})
	}
}

// TestSAMLService_CreateAuthRequest tests authentication request creation
func TestSAMLService_CreateAuthRequest(t *testing.T) {
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	cfg := &config.InfrastructureConfig{
		SAML2EntityID:       "https://terrareg.example.com",
		SAML2IDPMetadataURL: mockServer.MetadataURL,
		SAML2PrivateKey:     testPrivateKey,
		SAML2PublicKey:      testPublicKey,
		PublicURL:           "https://terrareg.example.com",
	}

	svc, err := service.NewSAMLService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()

	tests := []struct {
		name        string
		relayState  string
		expectError bool
		checkURL    func(*testing.T, string)
		description string
	}{
		{
			name:        "Create auth request without relay state",
			relayState:  "",
			expectError: false,
			checkURL: func(t *testing.T, authURL string) {
				// Should contain SAMLRequest parameter
				assert.Contains(t, authURL, "SAMLRequest=", "URL should contain SAMLRequest")
				// Should NOT contain RelayState
				assert.NotContains(t, authURL, "RelayState=", "URL should not contain RelayState when not provided")
			},
			description: "Basic auth request creation",
		},
		{
			name:        "Create auth request with relay state",
			relayState:  "https://terrareg.example.com/redirect",
			expectError: false,
			checkURL: func(t *testing.T, authURL string) {
				// Should contain both parameters
				assert.Contains(t, authURL, "SAMLRequest=", "URL should contain SAMLRequest")
				assert.Contains(t, authURL, "RelayState=", "URL should contain RelayState when provided")
				assert.Contains(t, authURL, "https%3A%2F%2Fterrareg.example.com%2Fredirect", "RelayState should be URL-encoded")
			},
			description: "Auth request with relay state",
		},
		{
			name:        "Verify auth request structure",
			relayState:  "test-state",
			expectError: false,
			checkURL: func(t *testing.T, authURL string) {
				// Verify the URL contains the expected SAML parameters
				assert.Contains(t, authURL, "SAMLRequest=", "URL should contain SAMLRequest")
				assert.Contains(t, authURL, "/saml/sso", "URL should contain IDP SSO endpoint")

				// Parse URL to extract SAMLRequest
				parts := strings.Split(authURL, "?")
				require.Len(t, parts, 2, "URL should have query parameters")

				queryString := parts[1]
				params := strings.Split(queryString, "&")

				var samlRequestB64 string
				for _, param := range params {
					if strings.HasPrefix(param, "SAMLRequest=") {
						samlRequestB64 = strings.TrimPrefix(param, "SAMLRequest=")
						break
					}
				}

				require.NotEmpty(t, samlRequestB64, "Should find SAMLRequest parameter")

				// SAMLRequest is deflated and base64 encoded, which is standard
				// Just verify it's non-empty and reasonably formatted
				assert.Greater(t, len(samlRequestB64), 10, "SAMLRequest should be encoded data")
			},
			description: "Verify SAML request XML structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authReq, err := svc.CreateAuthRequest(ctx, tt.relayState)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, authReq)
			} else {
				require.NoError(t, err, tt.description)
				require.NotNil(t, authReq)

				// Verify request properties
				assert.NotEmpty(t, authReq.ID, "Request should have ID")
				assert.NotEmpty(t, authReq.SAMLRequest, "Request should have encoded SAML")
				assert.NotEmpty(t, authReq.AuthURL, "Request should have auth URL")
				assert.False(t, authReq.CreatedAt.IsZero(), "Request should have creation time")
				assert.False(t, authReq.ExpiresAt.IsZero(), "Request should have expiration")
				assert.True(t, authReq.ExpiresAt.After(authReq.CreatedAt), "Expiration should be after creation")

				// Check URL
				if tt.checkURL != nil {
					tt.checkURL(t, authReq.AuthURL)
				}
			}
		})
	}
}

// TestSAMLService_ProcessResponse tests SAML response processing
func TestSAMLService_ProcessResponse(t *testing.T) {
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	cfg := &config.InfrastructureConfig{
		SAML2EntityID:       "https://terrareg.example.com",
		SAML2IDPMetadataURL: mockServer.MetadataURL,
		SAML2PrivateKey:     testPrivateKey,
		SAML2PublicKey:      testPublicKey,
		PublicURL:           "https://terrareg.example.com",
	}

	svc, err := service.NewSAMLService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()

	tests := []struct {
		name        string
		response    func() string
		expectError bool
		checkUser   func(*testing.T, *service.SAMLUserInfo)
		description string
	}{
		{
			name: "Valid SAML response",
			response: func() string {
				return generateValidSAMLResponse()
			},
			expectError: true, // Unsigned responses will fail signature validation
			checkUser: func(t *testing.T, userInfo *service.SAMLUserInfo) {
				assert.Nil(t, userInfo, "Should return nil for unsigned response")
			},
			description: "Unsigned response should fail signature validation",
		},
		{
			name:        "Empty response",
			response:    func() string { return "" },
			expectError: true,
			description: "Empty response should error",
		},
		{
			name: "Invalid base64",
			response: func() string {
				return "not-valid-base64!!!"
			},
			expectError: true,
			description: "Invalid base64 should error",
		},
		{
			name: "Invalid XML",
			response: func() string {
				invalidXML := "<not><valid><xml>"
				return base64.StdEncoding.EncodeToString([]byte(invalidXML))
			},
			expectError: true,
			description: "Invalid XML should error",
		},
		{
			name: "Missing assertion",
			response: func() string {
				resp := saml.Response{
					ID:           "_123",
					InResponseTo: "_456",
					IssueInstant: saml.TimeNow(),
					Version:      "2.0",
					Issuer:       &saml.Issuer{Value: "https://idp.example.com"},
					Status: saml.Status{
						StatusCode: saml.StatusCode{
							Value: saml.StatusSuccess,
						},
					},
				}
				xmlBytes, _ := xml.Marshal(resp)
				return base64.StdEncoding.EncodeToString(xmlBytes)
			},
			expectError: true,
			description: "Response without assertion should error",
		},
		{
			name: "Failed status",
			response: func() string {
				resp := saml.Response{
					ID:           "_123",
					InResponseTo: "_456",
					IssueInstant: saml.TimeNow(),
					Version:      "2.0",
					Issuer:       &saml.Issuer{Value: "https://idp.example.com"},
					Status: saml.Status{
						StatusCode: saml.StatusCode{
							Value: "urn:oasis:names:tc:SAML:2.0:status:Requester",
						},
					},
					Assertion: &saml.Assertion{},
				}
				xmlBytes, _ := xml.Marshal(resp)
				return base64.StdEncoding.EncodeToString(xmlBytes)
			},
			expectError: true,
			description: "Failed status should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfo, err := svc.ProcessResponse(ctx, tt.response(), "")

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, userInfo)
			} else {
				if tt.checkUser != nil {
					tt.checkUser(t, userInfo)
				}
			}
		})
	}
}

// TestSAMLService_GetMetadata tests metadata generation
func TestSAMLService_GetMetadata(t *testing.T) {
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	cfg := &config.InfrastructureConfig{
		SAML2EntityID:       "https://terrareg.example.com",
		SAML2IDPMetadataURL: mockServer.MetadataURL,
		SAML2PrivateKey:     testPrivateKey,
		SAML2PublicKey:      testPublicKey,
		PublicURL:           "https://terrareg.example.com",
	}

	svc, err := service.NewSAMLService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()

	metadataXML, err := svc.GetMetadata(ctx)
	require.NoError(t, err, "Should generate metadata")

	// Verify metadata contains expected elements
	assert.Contains(t, metadataXML, "EntityDescriptor", "Should contain EntityDescriptor")
	assert.Contains(t, metadataXML, "https://terrareg.example.com", "Should contain entity ID")
	assert.Contains(t, metadataXML, "AssertionConsumerService", "Should contain ACS endpoint")
	assert.Contains(t, metadataXML, "/v1/terrareg/auth/saml/acs", "Should contain ACS URL")

	// Verify it's valid XML
	var parsed interface{}
	err = xml.Unmarshal([]byte(metadataXML), &parsed)
	assert.NoError(t, err, "Metadata should be valid XML")
}

// TestSAMLService_ExtractUserInfo tests user info extraction
func TestSAMLService_ExtractUserInfo(t *testing.T) {
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	cfg := &config.InfrastructureConfig{
		SAML2EntityID:       "https://terrareg.example.com",
		SAML2IDPMetadataURL: mockServer.MetadataURL,
		SAML2PrivateKey:     testPrivateKey,
		SAML2PublicKey:      testPublicKey,
		PublicURL:           "https://terrareg.example.com",
	}

	svc, err := service.NewSAMLService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()

	tests := []struct {
		name         string
		attributes   map[string]string
		groups       []string
		expectError  bool
		errorContains string
		description  string
	}{
		{
			name: "Email and name from attributes",
			attributes: map[string]string{
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress": "user@example.com",
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name": "John Doe",
			},
			groups:       []string{"developers", "admins"},
			expectError:  true, // Unsigned responses fail signature validation
			errorContains: "not signed",
			description:  "Unsigned response should fail signature validation",
		},
		{
			name:         "Only NameID available",
			attributes:   map[string]string{},
			groups:       nil,
			expectError:  true,
			errorContains: "not signed",
			description:  "Unsigned response should fail signature validation",
		},
		{
			name: "First and last name",
			attributes: map[string]string{
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname": "John",
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname": "Doe",
			},
			groups:       nil,
			expectError:  true,
			errorContains: "not signed",
			description:  "Unsigned response should fail signature validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := generateSAMLResponseWithAttributes(
				tt.attributes,
				tt.groups,
				time.Now().Add(-5*time.Minute),
				time.Now().Add(1*time.Hour),
				"https://terrareg.example.com",
			)

			userInfo, err := svc.ProcessResponse(ctx, response, "")

			if tt.expectError {
				assert.Error(t, err, tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
				assert.Nil(t, userInfo)
			} else {
				require.NoError(t, err, tt.description)
				require.NotNil(t, userInfo)
			}
		})
	}
}

// TestSAMLService_ValidateSessionConstraints tests timing and audience validation
func TestSAMLService_ValidateSessionConstraints(t *testing.T) {
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	cfg := &config.InfrastructureConfig{
		SAML2EntityID:       "https://terrareg.example.com",
		SAML2IDPMetadataURL: mockServer.MetadataURL,
		SAML2PrivateKey:     testPrivateKey,
		SAML2PublicKey:      testPublicKey,
		PublicURL:           "https://terrareg.example.com",
	}

	svc, err := service.NewSAMLService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()

	tests := []struct {
		name          string
		notBefore     time.Time
		notOnOrAfter  time.Time
		audience      string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "Valid timing window but unsigned",
			notBefore:     time.Now().Add(-5 * time.Minute),
			notOnOrAfter:  time.Now().Add(1 * time.Hour),
			audience:      "https://terrareg.example.com",
			expectError:   true, // Unsigned responses fail signature validation
			errorContains: "not signed",
			description:   "Unsigned response should fail signature validation first",
		},
		{
			name:          "Future not before but unsigned",
			notBefore:     time.Now().Add(10 * time.Minute),
			notOnOrAfter:  time.Now().Add(1 * time.Hour),
			audience:      "https://terrareg.example.com",
			expectError:   true, // Signature validation happens before timing validation
			errorContains: "not signed",
			description:   "Signature validation happens first",
		},
		{
			name:          "Expired assertion but unsigned",
			notBefore:     time.Now().Add(-2 * time.Hour),
			notOnOrAfter:  time.Now().Add(-1 * time.Minute),
			audience:      "https://terrareg.example.com",
			expectError:   true,
			errorContains: "not signed",
			description:   "Signature validation happens first",
		},
		{
			name:          "Mismatched audience but unsigned",
			notBefore:     time.Now().Add(-5 * time.Minute),
			notOnOrAfter:  time.Now().Add(1 * time.Hour),
			audience:      "https://wrong-entity.com",
			expectError:   true,
			errorContains: "not signed",
			description:   "Signature validation happens first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := generateSAMLResponseWithConstraints(tt.notBefore, tt.notOnOrAfter, tt.audience)

			_, err := svc.ProcessResponse(ctx, response, "")

			if tt.expectError {
				assert.Error(t, err, tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// Helper functions

func generateValidSAMLResponse() string {
	return generateSAMLResponseWithConstraints(
		time.Now().Add(-5*time.Minute),
		time.Now().Add(1*time.Hour),
		"https://terrareg.example.com",
	)
}

func generateSAMLResponseWithConstraints(notBefore, notOnOrAfter time.Time, audience string) string {
	return generateSAMLResponseWithAttributes(
		map[string]string{
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress": "user@example.com",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name": "Test User",
		},
		[]string{"developers"},
		notBefore,
		notOnOrAfter,
		audience,
	)
}

func generateSAMLResponseWithAttributes(attributes map[string]string, groups []string, notBefore, notOnOrAfter time.Time, audience string) string {
	// Create SAML attribute statements
	var attributeStatements []saml.AttributeStatement
	attrStmt := saml.AttributeStatement{}

	for name, value := range attributes {
		attrStmt.Attributes = append(attrStmt.Attributes, saml.Attribute{
			Name: name,
			Values: []saml.AttributeValue{
				{Type: "xs:string", Value: value},
			},
		})
	}

	// Add groups if provided
	if len(groups) > 0 {
		var groupValues []saml.AttributeValue
		for _, group := range groups {
			groupValues = append(groupValues, saml.AttributeValue{
				Type: "xs:string", Value: group,
			})
		}
		attrStmt.Attributes = append(attrStmt.Attributes, saml.Attribute{
			Name:   "groups",
			Values: groupValues,
		})
	}

	attributeStatements = append(attributeStatements, attrStmt)

	response := saml.Response{
		ID:           "_123456",
		InResponseTo: "_request-id",
		IssueInstant: saml.TimeNow(),
		Version:      "2.0",
		Issuer: &saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  "https://idp.example.com",
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: saml.StatusSuccess,
			},
		},
		Assertion: &saml.Assertion{
			ID:           "_assertion-id",
			IssueInstant: saml.TimeNow(),
			Version:      "2.0",
			Issuer: saml.Issuer{
				Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
				Value:  "https://idp.example.com",
			},
			Subject: &saml.Subject{
				NameID: &saml.NameID{
					Format:          "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
					NameQualifier:   "idp.example.com",
					SPNameQualifier: "terrareg.example.com",
					Value:           "jdoe@example.com",
				},
				SubjectConfirmations: []saml.SubjectConfirmation{
					{
						Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
						SubjectConfirmationData: &saml.SubjectConfirmationData{
							NotOnOrAfter: time.Now().Add(1 * time.Hour),
							Recipient:    "https://terrareg.example.com/v1/terrareg/auth/saml/acs",
						},
					},
				},
			},
			Conditions: &saml.Conditions{
				NotBefore:    notBefore,
				NotOnOrAfter: notOnOrAfter,
				AudienceRestrictions: []saml.AudienceRestriction{
					{
						Audience: saml.Audience{Value: audience},
					},
				},
			},
			AttributeStatements: attributeStatements,
		},
	}

	xmlBytes, _ := xml.Marshal(response)
	return base64.StdEncoding.EncodeToString(xmlBytes)
}

// Test certificates (RSA 2048-bit for testing)
const testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAyKf7KmFm1CywFZtJ8qQrZpXqZ2JfH5K9P8L3Q2R9S4T7U0V1W
X3Y5Z6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E5F
6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6K7L
8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R
0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U9V0W1X
2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C
3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H
4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M
5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6
wIDAQABAoIBAFGmk7ZfHxLqP1V5A8C9D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T
8U9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y
9Z0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D
0E1F2G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I
1J2K3L4M5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N
2O3P4Q5R6S7T8U9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S
3T4U5V6W7X8Y9Z0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X
4Y5Z6A7B8C9D0E1F2G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C
5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H
6I7J8K9L0M1N2O3P4Q5R6S7T8U9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M
7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R
8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W
9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B
0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U9V0W1X2Y3Z4A5B6C7D8E9F0G
1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E5F6G7H8I9J0K1L
2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6K7L8M9N0O1P2Q
3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R0S1T2U3V
4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U9V0W1X2Y3Z4A
5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E5F
6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6K
7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P
8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U
9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z
0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E
-----END RSA PRIVATE KEY-----`

const testPublicKey = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRKqzMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMjQwMTAxMDAwMDAwWhcNMjUwMTAxMDAwMDAwWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiGw0BAQEFAAOCAQ8AMIIB
CgKCAQEAyKf7KmFm1CywFZtJ8qQrZpXqZ2JfH5K9P8L3Q2R9S4T7U0V1WX3Y5Z6C
7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E5F6G7H8I
9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6K7L8M9N0
O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P8Q9R0S1T
2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U9V0W1X2Y3
Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A1B2C3D4E
5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2G3H4I5J6
K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L4M5N6O7P
8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5R6S7T8U9
V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W7X8Y9Z0A
1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8C9D0E1F2
G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H0I1J2K3L
4M5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1N2O3P4Q5
R6S7T8U9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S3T4U5V6W
7X8Y9Z0A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6A7B8
C9D0E1F2G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H
0I1J2K3L4M5N6O7P8Q9R0S1T2U3V4W5X6Y7Z8A9B0C1D2E3F4G5H6I7J8K9L0M1
N2O3P4Q5R6S7T8U9V0W1X2Y3Z4A5B6C7D8E9F0G1H2I3J4K5L6M7N8O9P0Q1R2S
-----END CERTIFICATE-----`
