class Auth {
    constructor() {
        this.apiBaseUrl = window.location.hostname === 'localhost' 
            ? 'http://localhost:6002/api'
            : 'https://fileditch.ashank.tech/api';
        
        this.token = localStorage.getItem('token');
        this.isAdmin = localStorage.getItem('isAdmin') === 'true';
        
        this.checkAuth();
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        const loginForm = document.getElementById('loginForm');
        if (loginForm) {
            loginForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.login();
            });
        }

        const togglePasswordBtn = document.querySelector('.toggle-password');
        if (togglePasswordBtn) {
            togglePasswordBtn.addEventListener('click', () => this.togglePasswordVisibility());
        }
    }

    async login() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const loginButton = document.querySelector('.login-button');
        
        try {
            loginButton.dataset.state = 'loading';
            
            const response = await fetch(`${this.apiBaseUrl}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });
            
            if (!response.ok) throw new Error('Invalid credentials');
            
            const data = await response.json();
            
            localStorage.setItem('token', data.token);
            localStorage.setItem('username', data.username);
            localStorage.setItem('isAdmin', data.isAdmin);
            
            loginButton.dataset.state = 'success';
            
            // Redirect based on user role
            setTimeout(() => {
                window.location.href = data.isAdmin ? '/admin.html' : '/index.html';
            }, 1000);
        } catch (error) {
            loginButton.dataset.state = 'error';
            this.showToast('Invalid username or password', 'error');
            
            setTimeout(() => {
                loginButton.dataset.state = 'idle';
            }, 2000);
        }
    }

    checkAuth() {
        // Skip auth check for login page
        if (window.location.pathname.includes('login.html')) {
            if (this.token) {
                window.location.href = '/index.html';
            }
            return;
        }

        // Redirect to login if no token
        if (!this.token) {
            window.location.href = '/login.html';
            return;
        }

        // Check admin access for admin pages
        if (window.location.pathname.includes('admin.html') && !this.isAdmin) {
            window.location.href = '/index.html';
            return;
        }
    }

    logout() {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        localStorage.removeItem('isAdmin');
        window.location.href = '/login.html';
    }

    togglePasswordVisibility() {
        const passwordInput = document.getElementById('password');
        const type = passwordInput.type === 'password' ? 'text' : 'password';
        passwordInput.type = type;
        this.togglePasswordBtn.querySelector('.icon').textContent = type === 'password' ? '👁️' : '👁️‍🗨️';
    }

    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        
        document.body.appendChild(toast);
        
        requestAnimationFrame(() => {
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
                setTimeout(() => toast.remove(), 300);
            }, 3000);
        });
    }

    getAuthHeaders() {
        return {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json'
        };
    }
}

// Initialize auth
const auth = new Auth();