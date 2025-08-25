package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/consensys/gnark/backend/groth16"
)

// TestEndToEndWorkflow tests the complete workflow of storing balance, generating proof, and validating it
func TestEndToEndWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow end-to-end test")
	}

	// Clear balances for clean test
	balancesMu.Lock()
	balances = make(map[string]int)
	balancesMu.Unlock()

	scenarios := []struct {
		name         string
		userID       string
		balance      int
		neededAmount int
		shouldPass   bool
	}{
		{
			name:         "Sufficient balance scenario",
			userID:       "alice",
			balance:      200,
			neededAmount: 150,
			shouldPass:   true,
		},
		{
			name:         "Exact balance scenario",
			userID:       "bob",
			balance:      100,
			neededAmount: 100,
			shouldPass:   true,
		},
		{
			name:         "Insufficient balance scenario",
			userID:       "charlie",
			balance:      75,
			neededAmount: 100,
			shouldPass:   false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Step 1: Store balance
			t.Logf("Step 1: Storing balance %d for user %s", scenario.balance, scenario.userID)
			storeResp := storeBalanceE2E(t, scenario.userID, scenario.balance)
			if storeResp.Code != http.StatusOK {
				t.Fatalf("Failed to store balance: status %d, body: %s", storeResp.Code, storeResp.Body.String())
			}

			// Step 2: Generate proof
			t.Logf("Step 2: Generating proof for needed amount %d", scenario.neededAmount)
			startTime := time.Now()
			proofResp, proof := generateProofE2E(t, scenario.userID, scenario.neededAmount)
			proofTime := time.Since(startTime)
			t.Logf("Proof generation took: %v", proofTime)

			if scenario.shouldPass {
				if proofResp.Code != http.StatusOK {
					t.Fatalf("Expected proof generation to succeed, got status %d, body: %s", proofResp.Code, proofResp.Body.String())
				}

				// Step 3: Validate proof
				t.Logf("Step 3: Validating proof")
				startTime = time.Now()
				validateResp := validateProofE2E(t, scenario.userID, scenario.neededAmount, proof)
				validateTime := time.Since(startTime)
				t.Logf("Proof validation took: %v", validateTime)

				if validateResp.Code != http.StatusOK {
					t.Errorf("Expected proof validation to succeed, got status %d, body: %s", validateResp.Code, validateResp.Body.String())
				}
			} else {
				// For insufficient balance, proof generation should fail
				if proofResp.Code == http.StatusOK {
					t.Errorf("Expected proof generation to fail for insufficient balance, but it succeeded")
				}
			}
		})
	}
}

func TestConcurrentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow concurrent test")
	}

	// Clear balances for clean test
	balancesMu.Lock()
	balances = make(map[string]int)
	balancesMu.Unlock()

	numUsers := 3
	results := make(chan error, numUsers)

	// Run concurrent operations
	for i := 0; i < numUsers; i++ {
		go func(userID int) {
			userIDStr := fmt.Sprintf("user%d", userID)
			balance := 100 + userID*50 // Different balances
			neededAmount := 120

			// Store balance
			storeResp := storeBalanceE2E(t, userIDStr, balance)
			if storeResp.Code != http.StatusOK {
				results <- fmt.Errorf("user %d: failed to store balance", userID)
				return
			}

			// Generate proof (this is the expensive operation)
			proofResp, proof := generateProofE2E(t, userIDStr, neededAmount)
			if balance >= neededAmount {
				if proofResp.Code != http.StatusOK {
					results <- fmt.Errorf("user %d: expected proof generation to succeed", userID)
					return
				}

				// Validate proof
				validateResp := validateProofE2E(t, userIDStr, neededAmount, proof)
				if validateResp.Code != http.StatusOK {
					results <- fmt.Errorf("user %d: proof validation failed", userID)
					return
				}
			} else {
				// Should fail for insufficient balance
				if proofResp.Code == http.StatusOK {
					results <- fmt.Errorf("user %d: expected proof generation to fail", userID)
					return
				}
			}

			results <- nil
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numUsers; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out after 30 seconds")
		}
	}
}

func TestEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow edge case tests")
	}

	// Clear balances for clean test
	balancesMu.Lock()
	balances = make(map[string]int)
	balancesMu.Unlock()

	edgeCases := []struct {
		name         string
		userID       string
		balance      int
		neededAmount int
		expectError  bool
	}{
		{
			name:         "Zero balance and zero needed",
			userID:       "zero_user",
			balance:      0,
			neededAmount: 0,
			expectError:  false,
		},
		{
			name:         "Very large balance",
			userID:       "rich_user",
			balance:      1000000,
			neededAmount: 999999,
			expectError:  false,
		},
		{
			name:         "Negative needed amount",
			userID:       "negative_user",
			balance:      100,
			neededAmount: -10, // This should still work mathematically
			expectError:  false,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Store balance
			storeResp := storeBalanceE2E(t, tc.userID, tc.balance)
			if storeResp.Code != http.StatusOK {
				t.Fatalf("Failed to store balance: %v", storeResp.Body.String())
			}

			// Generate proof
			proofResp, proof := generateProofE2E(t, tc.userID, tc.neededAmount)
			
			if tc.expectError {
				if proofResp.Code == http.StatusOK {
					t.Errorf("Expected proof generation to fail, but it succeeded")
				}
				return
			}

			if proofResp.Code != http.StatusOK {
				t.Fatalf("Proof generation failed: %v", proofResp.Body.String())
			}

			// Validate proof
			validateResp := validateProofE2E(t, tc.userID, tc.neededAmount, proof)
			if validateResp.Code != http.StatusOK {
				t.Errorf("Proof validation failed: %v", validateResp.Body.String())
			}
		})
	}
}

// Helper functions for E2E testing

func storeBalanceE2E(t *testing.T, userID string, amount int) *httptest.ResponseRecorder {
	reqBody := BalanceRequest{
		ID:     userID,
		Amount: amount,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal store request: %v", err)
	}

	req, err := http.NewRequest("POST", "/store/sum", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create store request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(storeBalance)
	handler.ServeHTTP(rr, req)

	return rr
}

func generateProofE2E(t *testing.T, userID string, neededAmount int) (*httptest.ResponseRecorder, groth16.Proof) {
	reqBody := ProofRequest{
		ID:           userID,
		NeededAmount: neededAmount,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal proof request: %v", err)
	}

	req, err := http.NewRequest("POST", "/get/proof/neededAmount", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create proof request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(generateProof)
	handler.ServeHTTP(rr, req)

	var proof groth16.Proof
	if rr.Code == http.StatusOK {
		err := json.Unmarshal(rr.Body.Bytes(), &proof)
		if err != nil {
			t.Fatalf("Failed to unmarshal proof response: %v", err)
		}
	}

	return rr, proof
}

func validateProofE2E(t *testing.T, userID string, neededAmount int, proof groth16.Proof) *httptest.ResponseRecorder {
	reqBody := ValidateRequest{
		ID:           userID,
		NeededAmount: neededAmount,
		Proof:        proof,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal validate request: %v", err)
	}

	req, err := http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create validate request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(validateProof)
	handler.ServeHTTP(rr, req)

	return rr
}

func BenchmarkEndToEndWorkflow(b *testing.B) {
	// Setup
	balancesMu.Lock()
	balances = map[string]int{
		"benchmark_user": 200,
	}
	balancesMu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		userID := fmt.Sprintf("benchmark_user_%d", i)
		
		// Store balance (not timed)
		storeBalanceE2E_benchmark(userID, 200)
		
		b.StartTimer()
		// Time the proof generation and validation
		_, proof := generateProofE2E_benchmark(userID, 150)
		validateProofE2E_benchmark(userID, 150, proof)
	}
}

// Benchmark helpers that don't use testing.T
func storeBalanceE2E_benchmark(userID string, amount int) {
	reqBody := BalanceRequest{
		ID:     userID,
		Amount: amount,
	}
	
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/store/sum", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(storeBalance)
	handler.ServeHTTP(rr, req)
}

func generateProofE2E_benchmark(userID string, neededAmount int) (*httptest.ResponseRecorder, groth16.Proof) {
	reqBody := ProofRequest{
		ID:           userID,
		NeededAmount: neededAmount,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/get/proof/neededAmount", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(generateProof)
	handler.ServeHTTP(rr, req)

	var proof groth16.Proof
	if rr.Code == http.StatusOK {
		json.Unmarshal(rr.Body.Bytes(), &proof)
	}

	return rr, proof
}

func validateProofE2E_benchmark(userID string, neededAmount int, proof groth16.Proof) *httptest.ResponseRecorder {
	reqBody := ValidateRequest{
		ID:           userID,
		NeededAmount: neededAmount,
		Proof:        proof,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(validateProof)
	handler.ServeHTTP(rr, req)

	return rr
}