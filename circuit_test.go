package main

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func TestBalanceCircuit_Compilation(t *testing.T) {
	// Test that the circuit compiles successfully
	circuit := &BalanceCircuit{}

	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		t.Errorf("Circuit compilation failed: %v", err)
		return
	}

	// Basic sanity checks
	if ccs == nil {
		t.Error("Expected compiled circuit, got nil")
	}

	// Test setup works
	_, _, err = groth16.Setup(ccs)
	if err != nil {
		t.Errorf("Circuit setup failed: %v", err)
	}
}

func TestBalanceCircuit_CompileAndSetup(t *testing.T) {
	// Test that the circuit compiles successfully
	circuit := &BalanceCircuit{}

	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		t.Fatalf("Failed to compile circuit: %v", err)
	}

	// Test that setup works
	_, _, err = groth16.Setup(ccs)
	if err != nil {
		t.Fatalf("Failed to setup circuit: %v", err)
	}

	// Verify circuit properties
	if ccs.GetNbPublicVariables() != 2 { // 1 public input + 1 for the constant
		t.Errorf("Expected 2 public variables, got %d", ccs.GetNbPublicVariables())
	}

	if ccs.GetNbSecretVariables() != 1 { // 1 private input (balance)
		t.Errorf("Expected 1 secret variable, got %d", ccs.GetNbSecretVariables())
	}
}
