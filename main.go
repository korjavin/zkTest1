package main

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/io"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User represents a client with a bank balance
type User struct {
	ID      uint    `json:"id" gorm:"primaryKey"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

// Circuit defines the zk-SNARK circuit for proving balance
type Circuit struct {
	Balance   frontend.Variable `gnark:"balance"`
	Threshold frontend.Variable `gnark:"threshold"`
}

// Define circuit constraints
func (c *Circuit) Define(api frontend.API) error {
	api.AssertIsLessOrEqual(c.Threshold, c.Balance)
	return nil
}

// Database instance
var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("zkrollup.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	r := gin.Default()

	r.POST("/users", createUser)
	r.GET("/proof/:id", getProof)
	r.POST("/verify", verifyProof)

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func getProof(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	proof, err := generateZKProof(user.Balance, 100.0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"proof": proof,
	})
}

func verifyProof(c *gin.Context) {
	var payload struct {
		Proof     groth16.Proof `json:"proof"`
		Threshold float64       `json:"threshold"`
		Balance   float64       `json:"balance"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isValid, err := verifyZKProof(payload.Proof, payload.Balance, payload.Threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof verification failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isValid": isValid})
}

func generateZKProof(balance, threshold float64) (groth16.Proof, error) {
	var circuit Circuit
	witness := Circuit{
		Balance:   balance,
		Threshold: threshold,
	}

	r1cs, err := frontend.Compile(r1cs.NewBuilder, &circuit)
	if err != nil {
		return groth16.Proof{}, fmt.Errorf("failed to compile circuit: %w", err)
	}

	pk, _, err := groth16.Setup(r1cs)
	if err != nil {
		return groth16.Proof{}, fmt.Errorf("failed to setup zk-SNARK: %w", err)
	}

	witnessFull, err := io.NewWitness(&witness, io.PublicAndPrivate)
	if err != nil {
		return groth16.Proof{}, fmt.Errorf("failed to create witness: %w", err)
	}

	proof, err := groth16.Prove(r1cs, pk, witnessFull)
	if err != nil {
		return groth16.Proof{}, fmt.Errorf("failed to generate proof: %w", err)
	}

	return proof, nil
}

func verifyZKProof(proof groth16.Proof, balance, threshold float64) (bool, error) {
	var circuit Circuit
	publicWitness := Circuit{
		Threshold: threshold,
	}

	witnessFull, err := io.NewWitness(&publicWitness, io.PublicOnly)
	if err != nil {
		return false, fmt.Errorf("failed to create public witness: %w", err)
	}

	vk := groth16.NewVerifyingKey()
	if err := groth16.Verify(proof, vk, witnessFull); err != nil {
		return false, fmt.Errorf("proof verification failed: %w", err)
	}

	return true, nil
}

func TestGenerateZKProof(t *testing.T) {
	balance := 150.0
	threshold := 100.0
	proof, err := generateZKProof(balance, threshold)
	if err != nil {
		t.Fatalf("Failed to generate zk-proof: %v", err)
	}

	if proof == (groth16.Proof{}) {
		t.Fatalf("Generated proof is empty")
	}

	log.Printf("Generated proof: %v", proof)
	fmt.Printf("Generated proof: %v\n", proof)
}

func TestVerifyZKProof(t *testing.T) {
	balance := 150.0
	threshold := 100.0
	proof, err := generateZKProof(balance, threshold)
	if err != nil {
		t.Fatalf("Failed to generate zk-proof: %v", err)
	}

	isValid, err := verifyZKProof(proof, balance, threshold)
	if err != nil {
		t.Fatalf("Failed to verify zk-proof: %v", err)
	}

	if !isValid {
		t.Fatalf("Proof verification failed")
	}

	log.Println("Proof verified successfully")
	fmt.Println("Proof verified successfully")
}
