package main

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func TestProofGeneration(t *testing.T) {
	tests := []struct {
		name          string
		balance       int
		neededAmount  int
		shouldSucceed bool
	}{
		{
			name:          "Valid proof - sufficient balance",
			balance:       150,
			neededAmount:  100,
			shouldSucceed: true,
		},
		{
			name:          "Valid proof - exact balance",
			balance:       100,
			neededAmount:  100,
			shouldSucceed: true,
		},
		{
			name:          "Invalid proof - insufficient balance",
			balance:       50,
			neededAmount:  100,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and compile the circuit
			ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
			if err != nil {
				t.Fatalf("Failed to compile circuit: %v", err)
			}

			// Generate proving and verifying keys
			pk, vk, err := groth16.Setup(ccs)
			if err != nil {
				t.Fatalf("Failed to setup: %v", err)
			}

			// Create witness
			circuit := BalanceCircuit{
				Balance:      tt.balance,
				NeededAmount: tt.neededAmount,
			}

			witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
			if err != nil {
				t.Fatalf("Failed to create witness: %v", err)
			}

			// Generate proof
			proof, err := groth16.Prove(ccs, pk, witness)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected proof generation to succeed, but got error: %v", err)
					return
				}

				// Verify the proof
				publicWitness := BalanceCircuit{
					NeededAmount: tt.neededAmount,
				}

				pubWitness, err := frontend.NewWitness(&publicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
				if err != nil {
					t.Fatalf("Failed to create public witness: %v", err)
				}

				err = groth16.Verify(proof, vk, pubWitness)
				if err != nil {
					t.Errorf("Proof verification failed: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected proof generation to fail, but it succeeded")
				}
			}
		})
	}
}

func TestProofVerification(t *testing.T) {
	balance := 150
	neededAmount := 100

	// Setup circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
	if err != nil {
		t.Fatalf("Failed to compile circuit: %v", err)
	}

	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		t.Fatalf("Failed to setup: %v", err)
	}

	// Generate proof
	circuit := BalanceCircuit{
		Balance:      balance,
		NeededAmount: neededAmount,
	}

	witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("Failed to create witness: %v", err)
	}

	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	t.Run("Valid verification", func(t *testing.T) {
		publicWitness := BalanceCircuit{
			NeededAmount: neededAmount,
		}

		pubWitness, err := frontend.NewWitness(&publicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			t.Fatalf("Failed to create public witness: %v", err)
		}

		err = groth16.Verify(proof, vk, pubWitness)
		if err != nil {
			t.Errorf("Expected proof to verify, but got error: %v", err)
		}
	})

	t.Run("Invalid verification - wrong needed amount", func(t *testing.T) {
		wrongPublicWitness := BalanceCircuit{
			NeededAmount: neededAmount + 100, // Different needed amount
		}

		pubWitness, err := frontend.NewWitness(&wrongPublicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			t.Fatalf("Failed to create public witness: %v", err)
		}

		err = groth16.Verify(proof, vk, pubWitness)
		if err == nil {
			t.Error("Expected proof verification to fail with wrong needed amount, but it succeeded")
		}
	})
}

func BenchmarkProofGeneration(b *testing.B) {
	// Setup circuit once
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
	if err != nil {
		b.Fatalf("Failed to compile circuit: %v", err)
	}

	pk, _, err := groth16.Setup(ccs)
	if err != nil {
		b.Fatalf("Failed to setup: %v", err)
	}

	circuit := BalanceCircuit{
		Balance:      150,
		NeededAmount: 100,
	}

	witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	if err != nil {
		b.Fatalf("Failed to create witness: %v", err)
	}

	// Benchmark proof generation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := groth16.Prove(ccs, pk, witness)
		if err != nil {
			b.Fatalf("Failed to generate proof: %v", err)
		}
	}
}

func BenchmarkProofVerification(b *testing.B) {
	// Setup and generate proof once
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &BalanceCircuit{})
	if err != nil {
		b.Fatalf("Failed to compile circuit: %v", err)
	}

	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		b.Fatalf("Failed to setup: %v", err)
	}

	circuit := BalanceCircuit{
		Balance:      150,
		NeededAmount: 100,
	}

	witness, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	if err != nil {
		b.Fatalf("Failed to create witness: %v", err)
	}

	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		b.Fatalf("Failed to generate proof: %v", err)
	}

	publicWitness := BalanceCircuit{
		NeededAmount: 100,
	}

	pubWitness, err := frontend.NewWitness(&publicWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		b.Fatalf("Failed to create public witness: %v", err)
	}

	// Benchmark proof verification
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := groth16.Verify(proof, vk, pubWitness)
		if err != nil {
			b.Fatalf("Failed to verify proof: %v", err)
		}
	}
}
