package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// TestHelper provides utilities for testing the zkTest1 application
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// SetupCleanBalances clears the balances map for clean testing
func (h *TestHelper) SetupCleanBalances() {
	balancesMu.Lock()
	balances = make(map[string]int)
	balancesMu.Unlock()
}

// StoreBalance stores a balance for a user via HTTP API
func (h *TestHelper) StoreBalance(userID string, amount int) *httptest.ResponseRecorder {
	reqBody := BalanceRequest{
		ID:     userID,
		Amount: amount,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		h.t.Fatalf("Failed to marshal store request: %v", err)
	}

	req, err := http.NewRequest("POST", "/store/sum", bytes.NewBuffer(jsonBody))
	if err != nil {
		h.t.Fatalf("Failed to create store request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(storeBalance)
	handler.ServeHTTP(rr, req)

	return rr
}

// GenerateProof generates a proof for a user via HTTP API
func (h *TestHelper) GenerateProof(userID string, neededAmount int) (*httptest.ResponseRecorder, groth16.Proof) {
	reqBody := ProofRequest{
		ID:           userID,
		NeededAmount: neededAmount,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		h.t.Fatalf("Failed to marshal proof request: %v", err)
	}

	req, err := http.NewRequest("POST", "/get/proof/neededAmount", bytes.NewBuffer(jsonBody))
	if err != nil {
		h.t.Fatalf("Failed to create proof request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(generateProof)
	handler.ServeHTTP(rr, req)

	var proof groth16.Proof
	if rr.Code == http.StatusOK {
		err := json.Unmarshal(rr.Body.Bytes(), &proof)
		if err != nil {
			h.t.Fatalf("Failed to unmarshal proof response: %v", err)
		}
	}

	return rr, proof
}

// ValidateProof validates a proof via HTTP API
func (h *TestHelper) ValidateProof(userID string, neededAmount int, proof groth16.Proof) *httptest.ResponseRecorder {
	// Marshal proof to JSON first
	proofJSON, err := json.Marshal(proof)
	if err != nil {
		h.t.Fatalf("Failed to marshal proof: %v", err)
	}

	reqBody := ValidateRequest{
		ID:           userID,
		NeededAmount: neededAmount,
		Proof:        json.RawMessage(proofJSON),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		h.t.Fatalf("Failed to marshal validate request: %v", err)
	}

	req, err := http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonBody))
	if err != nil {
		h.t.Fatalf("Failed to create validate request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(validateProof)
	handler.ServeHTTP(rr, req)

	return rr
}

// CreateCircuitAndSetup creates and compiles a circuit with setup
func (h *TestHelper) CreateCircuitAndSetup() (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey) {
	circuit := &BalanceCircuit{}
	
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		h.t.Fatalf("Failed to compile circuit: %v", err)
	}

	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		h.t.Fatalf("Failed to setup circuit: %v", err)
	}

	return ccs, pk, vk
}

// GenerateTestProof generates a proof directly using the circuit (bypassing HTTP)
func (h *TestHelper) GenerateTestProof(balance, neededAmount int) (groth16.Proof, groth16.VerifyingKey) {
	ccs, pk, vk := h.CreateCircuitAndSetup()

	circuit := BalanceCircuit{
		Balance:      balance,
		NeededAmount: neededAmount,
	}

	witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	if err != nil {
		h.t.Fatalf("Failed to create witness: %v", err)
	}

	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		h.t.Fatalf("Failed to generate proof: %v", err)
	}

	return proof, vk
}

// VerifyTestProof verifies a proof directly using the circuit (bypassing HTTP)
func (h *TestHelper) VerifyTestProof(proof groth16.Proof, vk groth16.VerifyingKey, neededAmount int) bool {
	publicWitness := BalanceCircuit{
		NeededAmount: neededAmount,
	}
	
	witness, err := frontend.NewWitness(&publicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		h.t.Fatalf("Failed to create public witness: %v", err)
	}

	err = groth16.Verify(proof, vk, witness)
	return err == nil
}

// AssertStatusCode asserts that the response has the expected status code
func (h *TestHelper) AssertStatusCode(rr *httptest.ResponseRecorder, expected int, message string) {
	if rr.Code != expected {
		h.t.Errorf("%s: expected status %d, got %d. Body: %s", message, expected, rr.Code, rr.Body.String())
	}
}

// AssertBalanceStored checks that a balance was stored correctly
func (h *TestHelper) AssertBalanceStored(userID string, expectedAmount int) {
	balancesMu.Lock()
	actualAmount, exists := balances[userID]
	balancesMu.Unlock()

	if !exists {
		h.t.Errorf("Expected balance to be stored for user %s, but it was not found", userID)
		return
	}

	if actualAmount != expectedAmount {
		h.t.Errorf("Expected stored balance %d for user %s, got %d", expectedAmount, userID, actualAmount)
	}
}

// CreateTestScenario creates a complete test scenario with multiple users and balances
func (h *TestHelper) CreateTestScenario() map[string]int {
	h.SetupCleanBalances()
	
	scenario := map[string]int{
		"alice":   200,
		"bob":     100,
		"charlie": 50,
		"diana":   0,
	}

	for userID, balance := range scenario {
		rr := h.StoreBalance(userID, balance)
		h.AssertStatusCode(rr, http.StatusOK, fmt.Sprintf("storing balance for %s", userID))
		h.AssertBalanceStored(userID, balance)
	}

	return scenario
}

// LogTestStep logs a test step with formatting
func (h *TestHelper) LogTestStep(step int, description string, args ...interface{}) {
	message := fmt.Sprintf(description, args...)
	h.t.Logf("Step %d: %s", step, message)
}

// TestData represents common test data structures
type TestData struct {
	ValidScenarios []struct {
		Name         string
		UserID       string
		Balance      int
		NeededAmount int
		ShouldPass   bool
	}
	EdgeCases []struct {
		Name         string
		UserID       string
		Balance      int
		NeededAmount int
		ExpectError  bool
	}
}

// GetTestData returns predefined test data for common scenarios
func GetTestData() TestData {
	return TestData{
		ValidScenarios: []struct {
			Name         string
			UserID       string
			Balance      int
			NeededAmount int
			ShouldPass   bool
		}{
			{
				Name:         "Sufficient balance",
				UserID:       "user1",
				Balance:      150,
				NeededAmount: 100,
				ShouldPass:   true,
			},
			{
				Name:         "Exact balance",
				UserID:       "user2",
				Balance:      100,
				NeededAmount: 100,
				ShouldPass:   true,
			},
			{
				Name:         "Insufficient balance",
				UserID:       "user3",
				Balance:      75,
				NeededAmount: 100,
				ShouldPass:   false,
			},
		},
		EdgeCases: []struct {
			Name         string
			UserID       string
			Balance      int
			NeededAmount int
			ExpectError  bool
		}{
			{
				Name:         "Zero amounts",
				UserID:       "zero_user",
				Balance:      0,
				NeededAmount: 0,
				ExpectError:  false,
			},
			{
				Name:         "Large numbers",
				UserID:       "rich_user",
				Balance:      1000000,
				NeededAmount: 999999,
				ExpectError:  false,
			},
			{
				Name:         "Negative needed amount",
				UserID:       "negative_user",
				Balance:      100,
				NeededAmount: -10,
				ExpectError:  false,
			},
		},
	}
}

// SkipIfShort skips a test if running in short mode (useful for slow ZK tests)
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping %s (use -short=false to run)", reason)
	}
}