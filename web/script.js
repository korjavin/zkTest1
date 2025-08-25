// zkTest1 Frontend JavaScript
// Zero-Knowledge Proof Demo Interface

class ZKDemoApp {
    constructor() {
        this.apiBase = 'http://localhost:8080';
        this.currentUser = null;
        this.currentProof = null;
        this.demoState = {
            step1: false,
            step2: false,
            step3: false
        };
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupSmoothScrolling();
        this.setupInputValidation();
        this.resetDemo();
    }

    setupEventListeners() {
        // Demo step buttons
        document.getElementById('setBalanceBtn').addEventListener('click', () => this.handleSetBalance());
        document.getElementById('generateProofBtn').addEventListener('click', () => this.handleGenerateProof());
        document.getElementById('verifyProofBtn').addEventListener('click', () => this.handleVerifyProof());
        document.getElementById('resetDemoBtn').addEventListener('click', () => this.resetDemo());

        // Input change handlers
        document.getElementById('userBalance').addEventListener('input', () => this.updatePreview());
        document.getElementById('proofAmount').addEventListener('input', () => this.updatePreview());
        document.getElementById('userName').addEventListener('input', () => this.validateStep1());
    }

    setupSmoothScrolling() {
        // Smooth scrolling for navigation links
        document.querySelectorAll('.nav-link, a[href^="#"]').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href');
                if (targetId && targetId !== '#') {
                    const target = document.querySelector(targetId);
                    if (target) {
                        target.scrollIntoView({ 
                            behavior: 'smooth',
                            block: 'start'
                        });
                    }
                }
            });
        });
    }

    setupInputValidation() {
        // Real-time validation for balance and proof amount
        const balanceInput = document.getElementById('userBalance');
        const proofAmountInput = document.getElementById('proofAmount');

        balanceInput.addEventListener('input', (e) => {
            const value = parseFloat(e.target.value) || 0;
            if (value < 0) e.target.value = 0;
            this.updatePreview();
            this.validateStep1();
        });

        proofAmountInput.addEventListener('input', (e) => {
            const value = parseFloat(e.target.value) || 0;
            if (value < 0) e.target.value = 0;
            this.updatePreview();
        });
    }

    updatePreview() {
        const balance = document.getElementById('userBalance').value || 0;
        const proofAmount = document.getElementById('proofAmount').value || 0;
        
        document.getElementById('previewBalance').textContent = balance;
        document.getElementById('previewAmount').textContent = proofAmount;
        document.getElementById('verifyAmount').textContent = proofAmount;
    }

    validateStep1() {
        const name = document.getElementById('userName').value.trim();
        const balance = parseFloat(document.getElementById('userBalance').value) || 0;
        
        const btn = document.getElementById('setBalanceBtn');
        if (name && balance > 0) {
            btn.disabled = false;
            btn.innerHTML = '<i class="fas fa-piggy-bank"></i> Set My Balance';
        } else {
            btn.disabled = true;
        }
    }

    async handleSetBalance() {
        const name = document.getElementById('userName').value.trim();
        const balance = parseFloat(document.getElementById('userBalance').value);

        if (!name || !balance || balance <= 0) {
            this.showStatus('step1', 'error', 'Please enter a valid name and balance amount.');
            return;
        }

        this.showStatus('step1', 'loading', 'Setting up your balance...');

        try {
            const response = await fetch(`${this.apiBase}/store/sum`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: name.toLowerCase().replace(/\s+/g, ''),
                    amount: balance
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            this.currentUser = {
                id: name.toLowerCase().replace(/\s+/g, ''),
                name: name,
                balance: balance
            };

            this.demoState.step1 = true;
            this.showStatus('step1', 'success', `✅ Balance set! ${name} now has $${balance} (kept private)`);
            this.updateStepStates();
            this.enableStep2();

            // Auto-scroll to next step
            setTimeout(() => {
                document.getElementById('step2').scrollIntoView({ 
                    behavior: 'smooth',
                    block: 'center'
                });
            }, 1000);

        } catch (error) {
            console.error('Error setting balance:', error);
            this.showStatus('step1', 'error', 
                `Failed to set balance: ${error.message}. Make sure the zkTest1 server is running on port 8080.`
            );
        }
    }

    enableStep2() {
        const btn = document.getElementById('generateProofBtn');
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-key"></i> Generate ZK Proof';
    }

    async handleGenerateProof() {
        if (!this.currentUser) {
            this.showStatus('step2', 'error', 'Please set your balance first.');
            return;
        }

        const proofAmount = parseFloat(document.getElementById('proofAmount').value);
        if (!proofAmount || proofAmount <= 0) {
            this.showStatus('step2', 'error', 'Please enter a valid proof amount.');
            return;
        }

        // Check if proof is mathematically possible
        if (proofAmount > this.currentUser.balance) {
            this.showStatus('step2', 'error', 
                `⚠️ You're trying to prove you have ≥$${proofAmount}, but your balance is only $${this.currentUser.balance}. This will fail as expected in ZK - you cannot prove something that isn't true!`
            );
            return;
        }

        this.showLoadingModal(true);
        this.showStatus('step2', 'loading', '🔐 Generating zero-knowledge proof... This involves complex cryptographic calculations.');

        try {
            // Add realistic delay to show the complexity
            await this.delay(2000);

            const response = await fetch(`${this.apiBase}/get/proof/neededAmount`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: this.currentUser.id,
                    neededAmount: proofAmount
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const proofData = await response.json();
            this.currentProof = {
                data: proofData,
                amount: proofAmount,
                userId: this.currentUser.id
            };

            this.showLoadingModal(false);
            this.demoState.step2 = true;
            this.showStatus('step2', 'success', 
                `✅ ZK Proof generated! This mathematically proves you have ≥$${proofAmount} without revealing your actual balance of $${this.currentUser.balance}.`
            );
            
            this.displayProof(proofData);
            this.updateStepStates();
            this.enableStep3();

            // Auto-scroll to next step
            setTimeout(() => {
                document.getElementById('step3').scrollIntoView({ 
                    behavior: 'smooth',
                    block: 'center'
                });
            }, 1000);

        } catch (error) {
            console.error('Error generating proof:', error);
            this.showLoadingModal(false);
            this.showStatus('step2', 'error', 
                `Failed to generate proof: ${error.message}. The server may be processing other requests.`
            );
        }
    }

    displayProof(proofData) {
        const proofDisplay = document.getElementById('proofDisplay');
        proofDisplay.innerHTML = `
            <div class="proof-data">
                <div style="margin-bottom: 15px;">
                    <strong>🔐 Zero-Knowledge Proof Generated</strong>
                </div>
                <div style="background: #f0f0f0; padding: 15px; border-radius: 8px; font-family: monospace; font-size: 0.8rem; word-break: break-all; max-height: 100px; overflow-y: auto;">
                    ${JSON.stringify(proofData, null, 2).substring(0, 200)}...
                </div>
                <div style="margin-top: 10px; font-size: 0.9rem; color: #666;">
                    This cryptographic proof can verify your balance claim without exposing your actual balance.
                </div>
            </div>
        `;
    }

    enableStep3() {
        const btn = document.getElementById('verifyProofBtn');
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-check-circle"></i> Verify Proof';
    }

    async handleVerifyProof() {
        if (!this.currentProof) {
            this.showStatus('step3', 'error', 'Please generate a proof first.');
            return;
        }

        this.showStatus('step3', 'loading', '🔍 Verifying proof... Anyone can do this verification without accessing private data.');

        try {
            // Add delay to show verification process
            await this.delay(1500);

            const response = await fetch(`${this.apiBase}/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: this.currentProof.userId,
                    neededAmount: this.currentProof.amount,
                    proof: this.currentProof.data
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: Proof verification failed`);
            }

            this.demoState.step3 = true;
            this.showStatus('step3', 'success', 
                `🎉 Proof verified successfully! It's mathematically confirmed that you have ≥$${this.currentProof.amount}, but your exact balance remains private.`
            );
            
            this.updateStepStates();
            this.showDemoComplete();

        } catch (error) {
            console.error('Error verifying proof:', error);
            this.showStatus('step3', 'error', 
                `Proof verification failed: ${error.message}. This could happen if the proof data was corrupted or invalid.`
            );
        }
    }

    showDemoComplete() {
        const resultsDiv = document.getElementById('demoResults');
        resultsDiv.style.display = 'block';
        
        setTimeout(() => {
            resultsDiv.scrollIntoView({ 
                behavior: 'smooth',
                block: 'center'
            });
        }, 500);
    }

    resetDemo() {
        // Reset all state
        this.currentUser = null;
        this.currentProof = null;
        this.demoState = {
            step1: false,
            step2: false,
            step3: false
        };

        // Reset UI
        document.getElementById('userName').value = 'Alice';
        document.getElementById('userBalance').value = '1000';
        document.getElementById('proofAmount').value = '500';
        
        // Reset buttons
        document.getElementById('setBalanceBtn').disabled = false;
        document.getElementById('generateProofBtn').disabled = true;
        document.getElementById('verifyProofBtn').disabled = true;

        // Reset status messages
        this.hideAllStatus();

        // Reset proof display
        const proofDisplay = document.getElementById('proofDisplay');
        proofDisplay.innerHTML = `
            <div class="proof-placeholder">
                <i class="fas fa-lock"></i>
                <p>Generate a proof first</p>
            </div>
        `;

        // Hide results
        document.getElementById('demoResults').style.display = 'none';

        // Update preview
        this.updatePreview();

        // Update step states
        this.updateStepStates();

        // Scroll to first step
        setTimeout(() => {
            document.getElementById('step1').scrollIntoView({ 
                behavior: 'smooth',
                block: 'center'
            });
        }, 100);
    }

    updateStepStates() {
        // Update step visual states
        const steps = ['step1', 'step2', 'step3'];
        steps.forEach((stepId, index) => {
            const step = document.getElementById(stepId);
            step.classList.remove('active', 'completed');
            
            if (this.demoState[stepId]) {
                step.classList.add('completed');
            } else if (this.isStepActive(stepId)) {
                step.classList.add('active');
            }
        });
    }

    isStepActive(stepId) {
        switch (stepId) {
            case 'step1':
                return !this.demoState.step1;
            case 'step2':
                return this.demoState.step1 && !this.demoState.step2;
            case 'step3':
                return this.demoState.step2 && !this.demoState.step3;
            default:
                return false;
        }
    }

    showStatus(stepId, type, message) {
        const statusDiv = document.getElementById(`${stepId}Status`);
        statusDiv.className = `step-status ${type}`;
        statusDiv.innerHTML = `
            <i class="fas fa-${this.getStatusIcon(type)}"></i>
            ${message}
        `;
    }

    hideAllStatus() {
        ['step1Status', 'step2Status', 'step3Status'].forEach(id => {
            const statusDiv = document.getElementById(id);
            statusDiv.className = 'step-status';
            statusDiv.innerHTML = '';
        });
    }

    getStatusIcon(type) {
        switch (type) {
            case 'success': return 'check-circle';
            case 'error': return 'exclamation-triangle';
            case 'loading': return 'spinner fa-spin';
            default: return 'info-circle';
        }
    }

    showLoadingModal(show) {
        const modal = document.getElementById('loadingModal');
        if (show) {
            modal.classList.add('show');
            // Reset progress bar animation
            const progressFill = modal.querySelector('.progress-fill');
            progressFill.style.animation = 'none';
            setTimeout(() => {
                progressFill.style.animation = 'progress 5s ease-in-out';
            }, 100);
        } else {
            modal.classList.remove('show');
        }
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Utility function to check if API is available
    async checkAPIHealth() {
        try {
            const response = await fetch(`${this.apiBase}/health`, {
                method: 'GET',
                timeout: 5000
            });
            return response.ok;
        } catch (error) {
            return false;
        }
    }
}

// Educational content and animations
class EducationalAnimations {
    constructor() {
        this.init();
    }

    init() {
        this.setupScrollAnimations();
        this.setupHoverEffects();
        this.addEducationalTooltips();
    }

    setupScrollAnimations() {
        // Animate elements on scroll
        const animateOnScroll = () => {
            const elements = document.querySelectorAll('.learn-card, .app-card, .step');
            elements.forEach(el => {
                const rect = el.getBoundingClientRect();
                const isVisible = rect.top < window.innerHeight && rect.bottom > 0;
                
                if (isVisible) {
                    el.style.transform = 'translateY(0)';
                    el.style.opacity = '1';
                }
            });
        };

        // Set initial state
        document.querySelectorAll('.learn-card, .app-card, .step').forEach(el => {
            el.style.transform = 'translateY(50px)';
            el.style.opacity = '0';
            el.style.transition = 'transform 0.6s ease, opacity 0.6s ease';
        });

        window.addEventListener('scroll', animateOnScroll);
        animateOnScroll(); // Run once on load
    }

    setupHoverEffects() {
        // Add interactive hover effects to cards
        document.querySelectorAll('.learn-card, .app-card').forEach(card => {
            card.addEventListener('mouseenter', () => {
                card.style.transform = 'translateY(-10px) scale(1.02)';
            });
            
            card.addEventListener('mouseleave', () => {
                card.style.transform = 'translateY(0) scale(1)';
            });
        });
    }

    addEducationalTooltips() {
        // Add helpful tooltips to explain ZK concepts
        const tooltips = {
            'zk-snark': 'Zero-Knowledge Succinct Non-Interactive Argument of Knowledge - a cryptographic method to prove you know something without revealing what you know.',
            'groth16': 'A specific type of zk-SNARK that is very efficient for verification, making it practical for real-world use.',
            'circuit': 'A mathematical representation of the computation you want to prove. In our case, proving that balance ≥ needed_amount.',
            'witness': 'The private inputs (your actual balance) that satisfy the circuit constraints.',
            'proof': 'The cryptographic evidence that you have a valid witness, without revealing the witness itself.'
        };

        // Add tooltip functionality (simplified)
        Object.keys(tooltips).forEach(term => {
            document.querySelectorAll(`[data-tooltip="${term}"]`).forEach(el => {
                el.title = tooltips[term];
                el.style.cursor = 'help';
                el.style.borderBottom = '1px dotted #667eea';
            });
        });
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('🔐 zkTest1 Demo Loading...');
    
    // Initialize main demo app
    window.zkDemo = new ZKDemoApp();
    
    // Initialize educational animations
    new EducationalAnimations();
    
    // Add some helpful console messages for developers
    console.log(`
🚀 zkTest1 Zero-Knowledge Proof Demo

This demo shows how zero-knowledge proofs work in practice.
You can:
1. Set a private balance
2. Generate a proof that you have ≥ some amount
3. Verify the proof without revealing your actual balance

API Endpoints:
- POST ${window.zkDemo.apiBase}/store/sum
- POST ${window.zkDemo.apiBase}/get/proof/neededAmount  
- POST ${window.zkDemo.apiBase}/validate

Open the Network tab to see the actual API calls!
    `);
});

// Add some utility functions for debugging
window.debugZK = {
    checkAPI: () => window.zkDemo.checkAPIHealth(),
    resetDemo: () => window.zkDemo.resetDemo(),
    getCurrentState: () => ({
        user: window.zkDemo.currentUser,
        proof: window.zkDemo.currentProof,
        state: window.zkDemo.demoState
    })
};