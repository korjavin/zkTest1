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
	ID           string          `json:"id"`
	NeededAmount int             `json:"neededAmount"`
	Proof        json.RawMessage `json:"proof"`
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

	// Unmarshal the proof from JSON
	var proof groth16.Proof
	if err := json.Unmarshal(req.Proof, &proof); err != nil {
		http.Error(w, "invalid proof format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify the proof
	err = groth16.Verify(proof, vk, witness)
	if err != nil {
		http.Error(w, "invalid proof", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CORS middleware to allow frontend requests
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// API endpoints with CORS
	http.HandleFunc("/store/sum", enableCORS(storeBalance))
	http.HandleFunc("/get/proof/neededAmount", enableCORS(generateProof))
	http.HandleFunc("/validate", enableCORS(validateProof))

	// Serve static files for the demo frontend
	fs := http.FileServer(http.Dir("./web/"))
	http.Handle("/", fs)

	// Health check endpoint
	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "zkTest1 - Zero-Knowledge Proof Demo",
			"version": "1.0.0",
		})
	}))

	fmt.Println("🔐 zkTest1 Zero-Knowledge Proof Demo Server")
	fmt.Println("📊 API Server: http://localhost:8080")
	fmt.Println("🌐 Demo Frontend: http://localhost:8080")
	fmt.Println("📖 API Documentation: http://localhost:8080/#api")
	fmt.Println("🚀 Ready for zero-knowledge proof demonstrations!")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("❌ Failed to start server:", err)
	}
}
