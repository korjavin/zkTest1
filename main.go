package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/consensys/gnark"
	"github.com/consensys/gnark/frontend"
)

// Define the circuit
type BalanceCircuit struct {
	Balance      frontend.Variable `gnark:",public"`
	NeededAmount frontend.Variable `gnark:",public"`
}

func (circuit *BalanceCircuit) Define(api frontend.API) error {
	api.AssertIsLessOrEqual(circuit.NeededAmount, circuit.Balance)
	return nil
}

var (
	balances   = make(map[string]int)
	balancesMu sync.Mutex
)

type BalanceRequest struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

type ProofRequest struct {
	ID           string `json:"id"`
	NeededAmount int    `json:"neededAmount"`
}

type ValidateRequest struct {
	ID    string `json:"id"`
	Proof []byte `json:"proof"`
}

func storeBalance(w http.ResponseWriter, r *http.Request) {
	var req BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balancesMu.Lock()
	balances[req.ID] = req.Amount
	balancesMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func generateProof(w http.ResponseWriter, r *http.Request) {
	var req ProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balancesMu.Lock()
	balance, exists := balances[req.ID]
	balancesMu.Unlock()

	if !exists {
		http.Error(w, "balance not found", http.StatusNotFound)
		return
	}

	// Create a circuit
	var circuit BalanceCircuit
	circuit.Balance = balance
	circuit.NeededAmount = req.NeededAmount

	// Generate proof
	proof, err := gnark.NewProver(circuit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proof)
}

func validateProof(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balancesMu.Lock()
	balance, exists := balances[req.ID]
	balancesMu.Unlock()

	if !exists {
		http.Error(w, "balance not found", http.StatusNotFound)
		return
	}

	// Create a circuit
	var circuit BalanceCircuit
	circuit.Balance = balance

	// Verify proof
	valid, err := gnark.NewVerifier(circuit, req.Proof)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "invalid proof", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/store/sum", storeBalance)
	http.HandleFunc("/get/proof/neededAmount", generateProof)
	http.HandleFunc("/validate", validateProof)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
