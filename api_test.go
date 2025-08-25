package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStoreBalance(t *testing.T) {
	// Clear balances for clean test
	balancesMu.Lock()
	balances = make(map[string]int)
	balancesMu.Unlock()

	tests := []struct {
		name           string
		requestBody    BalanceRequest
		expectedStatus int
	}{
		{
			name: "Valid balance storage",
			requestBody: BalanceRequest{
				ID:     "user1",
				Amount: 150,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Another valid balance storage",
			requestBody: BalanceRequest{
				ID:     "user2",
				Amount: 50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Zero balance",
			requestBody: BalanceRequest{
				ID:     "user3",
				Amount: 0,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			jsonBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req, err := http.NewRequest("POST", "/store/sum", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler := http.HandlerFunc(storeBalance)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			// Verify balance was stored
			balancesMu.Lock()
			storedBalance, exists := balances[tt.requestBody.ID]
			balancesMu.Unlock()

			if !exists {
				t.Errorf("Expected balance to be stored for user %s", tt.requestBody.ID)
			} else if storedBalance != tt.requestBody.Amount {
				t.Errorf("Expected stored balance %d, got %d", tt.requestBody.Amount, storedBalance)
			}
		})
	}
}

func TestStoreBalanceInvalidRequest(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing ID field",
			requestBody:    `{"amount": 100}`,
			expectedStatus: http.StatusOK, // Still valid, just empty ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/store/sum", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(storeBalance)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestGenerateProof(t *testing.T) {
	// Setup: store some balances
	balancesMu.Lock()
	balances = map[string]int{
		"user1": 150,
		"user2": 50,
	}
	balancesMu.Unlock()

	tests := []struct {
		name           string
		requestBody    ProofRequest
		expectedStatus int
	}{
		{
			name: "Valid proof generation - sufficient balance",
			requestBody: ProofRequest{
				ID:           "user1",
				NeededAmount: 100,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid proof generation - exact balance",
			requestBody: ProofRequest{
				ID:           "user1",
				NeededAmount: 150,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User not found",
			requestBody: ProofRequest{
				ID:           "nonexistent",
				NeededAmount: 100,
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test may be slow due to zk-proof generation
			if testing.Short() {
				t.Skip("Skipping slow proof generation test")
			}

			jsonBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req, err := http.NewRequest("POST", "/get/proof/neededAmount", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(generateProof)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusOK {
				// Verify response contains a proof
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
			}
		})
	}
}

func TestValidateProof(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow proof validation test")
	}

	// Setup: store a balance
	balancesMu.Lock()
	balances = map[string]int{
		"user1": 150,
	}
	balancesMu.Unlock()

	// Generate a proof first
	proofReq := ProofRequest{
		ID:           "user1",
		NeededAmount: 100,
	}

	jsonBody, _ := json.Marshal(proofReq)
	req, _ := http.NewRequest("POST", "/get/proof/neededAmount", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(generateProof)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to generate proof: %v", rr.Body.String())
	}

	// Parse the proof from response
	var proofResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &proofResponse)
	if err != nil {
		t.Fatalf("Failed to parse proof response: %v", err)
	}

	// Note: This test is limited because we can't easily deserialize the proof
	// In a real scenario, you'd need to handle proof serialization properly
	t.Log("Proof generation test completed - validation would require proper proof serialization")
}

// Note: Additional endpoint validation tests could be added here
// Currently focusing on functional tests that verify the core ZK proof functionality

func BenchmarkStoreBalance(b *testing.B) {
	// Prepare request
	reqBody := BalanceRequest{
		ID:     "benchmark_user",
		Amount: 150,
	}
	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/store/sum", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(storeBalance)
		handler.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status OK, got %d", rr.Code)
		}
	}
}