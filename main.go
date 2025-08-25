package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// Define the circuit
type BalanceCircuit struct {
	Balance      frontend.Variable `gnark:",private"`
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
	ID           string                `json:"id"`
	NeededAmount int                   `json:"neededAmount"`
	Proof        groth16.Proof         `json:"proof"`
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

	// Compile the circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate the proving and verifying keys
	pk, _, err := groth16.Setup(ccs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create witness
	witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate the proof
	proof, err := groth16.Prove(ccs, pk, witness)
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

	// Compile the circuit (we need this to get the verifying key)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate the proving and verifying keys
	_, vk, err := groth16.Setup(ccs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create public witness (only the public inputs)
	publicWitness := BalanceCircuit{
		NeededAmount: req.NeededAmount,
	}
	
	witness, err := frontend.NewWitness(&publicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify the proof
	err = groth16.Verify(req.Proof, vk, witness)
	if err != nil {
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
