# zkTest1 - Zero-Knowledge Proof Balance Verification API

A Go-based REST API that demonstrates zero-knowledge proofs (zk-SNARKs) for private balance verification. Users can prove they have a minimum balance threshold without revealing their actual balance amount.

## 🔒 What is This?

This project implements a privacy-preserving balance verification system using zero-knowledge proofs. It allows users to:

- **Prove** they have at least a certain amount of money without revealing the exact balance
- **Verify** such proofs without accessing private financial information
- **Maintain privacy** while satisfying regulatory or business requirements

## 🛠️ Technology Stack

- **Go 1.23.4** - Backend language
- **Gin** - HTTP web framework
- **GORM** - ORM for database operations
- **SQLite** - Database for user storage
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

### 1. Create User
Creates a new user with a balance.

```bash
POST /users
Content-Type: application/json

{
  "name": "Alice",
  "balance": 150.0
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Alice", 
  "balance": 150.0
}
```

### 2. Generate Proof
Generates a zk-SNARK proof that a user has at least the threshold balance.

```bash
GET /proof/1
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "name": "Alice",
    "balance": 150.0
  },
  "proof": "0x..." // zk-SNARK proof data
}
```

### 3. Verify Proof
Verifies a zk-SNARK proof without revealing the actual balance.

```bash
POST /verify
Content-Type: application/json

{
  "proof": "0x...",
  "threshold": 100.0,
  "balance": 150.0
}
```

**Response:**
```json
{
  "isValid": true
}
```

## 🧪 Testing

The project includes comprehensive tests for both proof generation and verification:

```bash
# Run all tests
go test -v

# Run specific test
go test -run TestGenerateZKProof -v
go test -run TestVerifyZKProof -v
```

## 📁 Project Structure

```
zkTest1/
├── main.go          # Main application with API endpoints and zk-proof logic
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── zkrollup.db      # SQLite database (created automatically)
└── README.md        # This file
```

## 🔧 How It Works

### Zero-Knowledge Proof Circuit

The project defines a simple circuit in `main.go:25-35`:

```go
type Circuit struct {
    Balance   frontend.Variable `gnark:"balance"`
    Threshold frontend.Variable `gnark:"threshold"`
}

func (c *Circuit) Define(api frontend.API) error {
    api.AssertIsLessOrEqual(c.Threshold, c.Balance)
    return nil
}
```

This circuit proves: **threshold ≤ balance** without revealing the actual balance value.

### Proof Generation Process

1. User's balance is used as a private input
2. Threshold is a public input
3. Circuit compiles the constraint: `threshold ≤ balance`
4. Groth16 generates a cryptographic proof
5. Proof can be verified by anyone without seeing the balance

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