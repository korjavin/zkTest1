# GitHub Actions CI/CD Pipeline

This directory contains the GitHub Actions workflows for the zkTest1 project. Our CI/CD pipeline ensures code quality, security, and proper testing of the zero-knowledge proof functionality.

## 📋 Workflow Files

### 🔄 `ci.yml` - Main CI Pipeline
**Triggers:** Push/PR to main branches, manual dispatch
**Duration:** ~15-45 minutes (depending on test selection)

**Jobs:**
1. **Fast Tests & Linting** (10 min) - Quick feedback loop
   - Code formatting checks (`gofmt`)
   - Static analysis (`go vet`)
   - Fast tests (`go test -short`) 
   - Build verification

2. **Unit Tests with Coverage** (15 min) - Core functionality
   - Circuit compilation tests
   - ZK proof generation/verification tests
   - Coverage reporting to Codecov

3. **Integration Tests** (15 min) - API functionality  
   - HTTP endpoint tests
   - Balance storage tests
   - Request validation tests

4. **Full ZK Tests** (45 min) - Complete validation
   - **Conditional:** Manual trigger, main branch, or `[full-tests]` in commit
   - Complete ZK proof workflow tests
   - End-to-end scenarios with multiple users
   - Concurrent operation tests

5. **Performance Benchmarks** (30 min) - Performance tracking
   - **Conditional:** Main branch or manual trigger
   - Proof generation benchmarks
   - Proof verification benchmarks
   - Results stored as artifacts

6. **Cross-Platform Build** (10 min) - Compatibility
   - Build on Ubuntu, Windows, macOS
   - Fast test execution on each platform

7. **Test Results Summary** (5 min) - Status reporting
   - Consolidated test status
   - GitHub summary with results table

### 🚀 `release.yml` - Release Pipeline
**Triggers:** Git tags (`v*`), manual dispatch
**Duration:** ~90 minutes

**Jobs:**
1. **Pre-Release Full Tests** (60 min)
   - Complete test suite execution
   - Performance benchmark validation

2. **Build Release Binaries** (20 min)
   - Cross-platform binary compilation
   - Linux, macOS, Windows support
   - AMD64 and ARM64 architectures

3. **Create GitHub Release** (10 min)
   - Automatic release creation
   - Binary artifact attachment
   - Release notes generation

### 🔒 `security.yml` - Security & Dependencies  
**Triggers:** Weekly schedule, main branch pushes, dependency changes
**Duration:** ~30 minutes

**Jobs:**
1. **Security Scan** (15 min)
   - Vulnerability detection (`govulncheck`)
   - Static security analysis (`gosec`)
   - Dependency vulnerability check (`nancy`)

2. **Dependency Analysis** (10 min)
   - Dependency graph analysis
   - License compatibility check
   - Update recommendations

3. **Code Quality Analysis** (15 min)
   - Static code analysis (`staticcheck`)
   - Dead code detection (`ineffassign`)
   - Typo detection (`misspell`)

4. **Performance Regression Check** (30 min)
   - **PR-only:** Benchmark comparison
   - Performance impact analysis
   - Regression detection

## 🎯 Test Strategy

### Fast Development Loop
```bash
# Triggered on every push/PR
Fast Tests (5 min) → Unit Tests (10 min) → Integration Tests (10 min)
```
**Goal:** Quick feedback for developers

### Comprehensive Validation  
```bash
# Triggered on main branch or manual
Full ZK Tests (45 min) → Benchmarks (30 min) → Cross-Platform (10 min)
```
**Goal:** Complete system validation

### Release Quality Assurance
```bash
# Triggered on version tags
Full Tests (60 min) → Multi-Platform Build (20 min) → Release (10 min)
```
**Goal:** Production-ready releases

## ⚡ Performance Optimizations

### Caching Strategy
- **Go modules:** Cached based on `go.sum` hash
- **Build artifacts:** Cached across workflow runs  
- **Dependencies:** Pre-downloaded and cached

### Parallel Execution
- **Unit & Integration tests:** Run in parallel after fast tests
- **Cross-platform builds:** Matrix strategy for parallel execution
- **Benchmark jobs:** Independent execution for faster feedback

### Conditional Execution
- **Full ZK tests:** Only when necessary (main branch, manual, or `[full-tests]`)
- **Benchmarks:** Only on main branch or manual trigger
- **Security scans:** Scheduled weekly + triggered by changes

## 🚦 Status Indicators

### Branch Protection
The following checks are **required** for PR merges:
- ✅ Fast Tests & Linting
- ✅ Unit Tests with Coverage  
- ✅ Integration Tests
- ✅ Cross-Platform Build (Ubuntu)

### Optional Checks
These provide additional confidence but don't block merges:
- 🔶 Full ZK Tests (main branch only)
- 🔶 Performance Benchmarks (main branch only)
- 🔶 Security Scans (scheduled)

## 📊 Monitoring & Reporting

### Test Coverage
- Unit test coverage reported to **Codecov**
- Coverage trends tracked over time
- Pull request coverage comparison

### Performance Tracking  
- Benchmark results stored as **artifacts**
- Performance regression detection in PRs
- Historical performance data

### Security Monitoring
- Weekly vulnerability scans
- Dependency update notifications
- Security advisory integration

## 🔧 Manual Triggers

### Full Test Suite
```
Actions → CI Pipeline → Run workflow → ✅ Run full test suite
```

### Security Scan
```  
Actions → Security & Dependencies → Run workflow
```

### Release
```
Actions → Release → Run workflow → Enter tag (e.g., v1.0.0)
```

## 🛠️ Local Development

Match CI environment locally:
```bash
# Fast tests (same as CI fast-tests job)
make test

# Full test suite (same as CI full-zk-tests job)  
make test-all

# Coverage report (same as CI unit-tests job)
make test-coverage

# Cross-platform build test
GOOS=linux GOARCH=amd64 go build .
```

## 📈 Metrics & KPIs

### CI Performance
- **Fast feedback:** < 10 minutes for basic validation
- **Full validation:** < 45 minutes for complete testing  
- **Release pipeline:** < 90 minutes end-to-end

### Test Coverage
- **Target:** > 80% code coverage
- **Critical paths:** 100% coverage for ZK proof functions
- **API endpoints:** 100% coverage for all handlers

### Security  
- **Zero known vulnerabilities** in production releases
- **Weekly security scans** for early detection
- **Automated dependency updates** for security patches

---

This CI/CD pipeline ensures the zkTest1 zero-knowledge proof system maintains high quality, security, and performance standards while providing fast feedback to developers. 🚀