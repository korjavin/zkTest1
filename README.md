# zkTest1 - Zero-Knowledge Proof Balance Verification API

A Go-based REST API that demonstrates zero-knowledge proofs (zk-SNARKs) for private balance verification. Users can prove they have a minimum balance threshold without revealing their actual balance amount.

## 🔒 What is This?

This project implements a privacy-preserving balance verification system using zero-knowledge proofs. It allows users to:

- **Prove** they have at least a certain amount of money without revealing the exact balance
- **Verify** such proofs without accessing private financial information
- **Maintain privacy** while satisfying regulatory or business requirements

## 🛠️ Technology Stack

- **Go 1.23.4** - Backend language
- **net/http** - Standard HTTP server
- **Gnark** - Zero-knowledge proof framework by ConsenSys
- **Groth16** - zk-SNARK proving system

## 📋 Features

### Core Functionality
- **User Management**: Create and store users with their balances
- **Proof Generation**: Generate zk-SNARK proofs for balance verification
- **Proof Verification**: Verify proofs without accessing private data
- **REST API**: Clean HTTP endpoints for all operations

### Privacy Benefits
- ✅ Prove you have ≥ $100 without revealing if you have $101 or $1,000,000
- ✅ No sensitive balance data transmitted during verification
- ✅ Cryptographically secure proofs that cannot be forged
- ✅ Perfect for compliance, loans, or membership verification

## 🚀 Quick Start

### Prerequisites
- Go 1.23.4 or later
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/korjavin/zkTest1.git
   cd zkTest1
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

## 🔌 API Endpoints

### 1. Store Balance
Stores a user's balance privately in the system.

```bash
POST /store/sum
Content-Type: application/json

{
  "id": "alice123",
  "amount": 150
}
```

**Response:**
```
HTTP 200 OK
```

### 2. Generate Proof
Generates a zk-SNARK proof that a user has at least the required amount.

```bash
POST /get/proof/neededAmount
Content-Type: application/json

{
  "id": "alice123",
  "neededAmount": 100
}
```

**Response:**
```json
{
  // zk-SNARK proof object
  "proof": {...}
}
```

### 3. Validate Proof
Validates a zk-SNARK proof without revealing the actual balance.

```bash
POST /validate
Content-Type: application/json

{
  "id": "alice123",
  "neededAmount": 100,
  "proof": {...}  // proof object from step 2
}
```

**Response:**
```
HTTP 200 OK (proof valid)
HTTP 401 Unauthorized (proof invalid)
```

## 🧪 Testing

### Automated Testing

The project includes comprehensive test coverage:

**Test Categories:**
- **Unit Tests** - Circuit logic, proof generation/verification
- **Integration Tests** - API endpoint functionality 
- **End-to-End Tests** - Complete workflows with multiple users
- **Benchmarks** - Performance testing for ZK operations

**Running Tests:**
```bash
# Quick tests (recommended for development)
make test
# or: go test -short -v

# All tests including slow ZK proof operations
make test-all
# or: go test -v -timeout 30m

# Specific test categories
make test-unit          # Circuit and proof tests
make test-integration   # API endpoint tests  
make test-e2e          # End-to-end workflows

# Performance testing
make benchmark         # All benchmarks
make benchmark-proof   # ZK proof benchmarks only

# Coverage report
make test-coverage     # Generates coverage.html
```

**Test Files:**
- `circuit_test.go` - Circuit compilation and logic tests
- `proof_test.go` - ZK proof generation and verification tests
- `api_test.go` - HTTP API integration tests
- `e2e_test.go` - End-to-end workflow tests
- `test_utils.go` - Testing utilities and helpers

### Manual Testing

Test the API manually with curl:

```bash
# 1. Store a balance
curl -X POST http://localhost:8080/store/sum \
  -H "Content-Type: application/json" \
  -d '{"id": "alice123", "amount": 150}'

# 2. Generate proof for minimum amount
curl -X POST http://localhost:8080/get/proof/neededAmount \
  -H "Content-Type: application/json" \
  -d '{"id": "alice123", "neededAmount": 100}'

# 3. Validate the proof (use the proof from step 2)
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"id": "alice123", "neededAmount": 100, "proof": {...}}'
```

## 📁 Project Structure

```
zkTest1/
├── main.go          # Main application with API endpoints and zk-proof logic
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── README.md        # This file
```

## 🔧 How It Works

### Zero-Knowledge Proof Circuit

The project defines a simple circuit in `main.go:15-23`:

```go
type BalanceCircuit struct {
    Balance      frontend.Variable `gnark:",private"`
    NeededAmount frontend.Variable `gnark:",public"`
}

func (circuit *BalanceCircuit) Define(api frontend.API) error {
    api.AssertIsLessOrEqual(circuit.NeededAmount, circuit.Balance)
    return nil
}
```

This circuit proves: **neededAmount ≤ balance** without revealing the actual balance value.

### Proof Generation Process

1. User's balance is stored privately in memory
2. Required amount (neededAmount) is a public input  
3. Circuit compiles the constraint: `neededAmount ≤ balance`
4. Groth16 generates a cryptographic proof
5. Proof can be verified by anyone without seeing the actual balance

## 🎯 Use Cases

- **Financial Services**: Prove creditworthiness without sharing account details
- **Membership Verification**: Prove eligibility without revealing wealth
- **Regulatory Compliance**: Satisfy KYC requirements while maintaining privacy
- **DeFi Applications**: Private balance verification for lending protocols
- **Corporate Finance**: Verify company assets without disclosure

## 🛡️ Security Considerations

- **Trusted Setup**: Groth16 requires a trusted setup ceremony (simplified in this demo)
- **Circuit Security**: The constraint logic must be carefully audited
- **Key Management**: Production systems need secure key storage
- **Proof Freshness**: Consider timestamps to prevent replay attacks

## 🔮 Future Enhancements

- [ ] Add proof caching and optimization
- [ ] Implement more complex financial circuits
- [ ] Support for multiple proof types
- [ ] Docker containerization
- [ ] Database migrations
- [ ] Comprehensive API documentation
- [ ] Performance benchmarking

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- **ConsenSys Gnark** - Excellent zk-SNARK framework
- **Gin Framework** - Fast HTTP framework for Go
- **GORM** - Fantastic ORM for Go
- **Zero-Knowledge Proof Community** - For advancing privacy-preserving technologies

---

**Note**: This is a demonstration project. For production use, implement proper security measures, key management, and consider a more robust trusted setup process.