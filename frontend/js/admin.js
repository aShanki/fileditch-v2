class AdminUI {
    constructor() {
        this.apiBaseUrl = window.location.hostname === 'localhost' 
            ? 'http://localhost:6002'
            : 'https://fileditch.ashank.tech';
        
        this.usersList = document.getElementById('usersList');
        this.addUserForm = document.getElementById('addUserForm');
        this.editUserForm = document.getElementById('editUserForm');
        
        this.initializeEventListeners();
        this.loadUsers();
    }

    initializeEventListeners() {
        if (this.addUserForm) {
            this.addUserForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.createUser();
            });
        }

        if (this.editUserForm) {
            this.editUserForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.updateUser();
            });
        }
    }

    async loadUsers() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/admin/users`, {
                headers: auth.getAuthHeaders()
            });

            if (!response.ok) throw new Error('Failed to load users');
            
            const users = await response.json();
            this.renderUsers(users);
        } catch (error) {
            this.showToast('Failed to load users', 'error');
        }
    }

    renderUsers(users) {
        this.usersList.innerHTML = '';
        
        users.forEach(user => {
            const userElement = document.createElement('div');
            userElement.className = 'user-item';
            userElement.innerHTML = `
                <div class="user-info">
                    <strong>${user.username}</strong>
                    <span class="user-role">${user.isAdmin ? 'Admin' : 'User'}</span>
                </div>
                <div class="user-actions">
                    <button class="action-button" onclick="app.showEditUserModal('${user.id}', '${user.username}', ${user.isAdmin})">
                        <span class="button-content">
                            <svg class="button-icon" viewBox="0 0 24 24" width="14" height="14">
                                <path fill="currentColor" d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                            </svg>
                            <span>Edit</span>
                        </span>
                    </button>
                    <button class="action-button delete" onclick="app.deleteUser('${user.id}')">
                        <span class="button-content">
                            <svg class="button-icon" viewBox="0 0 24 24" width="14" height="14">
                                <path fill="currentColor" d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                            </svg>
                            <span>Delete</span>
                        </span>
                    </button>
                </div>
            `;
            this.usersList.appendChild(userElement);
        });
    }

    showAddUserModal() {
        const modal = document.getElementById('addUserModal');
        modal.classList.remove('hidden');
    }

    hideAddUserModal() {
        const modal = document.getElementById('addUserModal');
        modal.classList.add('hidden');
        this.addUserForm.reset();
    }

    showEditUserModal(userId, username, isAdmin) {
        const modal = document.getElementById('editUserModal');
        document.getElementById('editUserId').value = userId;
        document.getElementById('editUsername').value = username;
        document.getElementById('editIsAdmin').checked = isAdmin;
        modal.classList.remove('hidden');
    }

    hideEditUserModal() {
        const modal = document.getElementById('editUserModal');
        modal.classList.add('hidden');
        this.editUserForm.reset();
    }

    async createUser() {
        const username = document.getElementById('newUsername').value;
        const password = document.getElementById('newPassword').value;
        const isAdmin = document.getElementById('isAdmin').checked;

        try {
            const response = await fetch(`${this.apiBaseUrl}/admin/users`, {
                method: 'POST',
                headers: auth.getAuthHeaders(),
                body: JSON.stringify({ username, password, isAdmin })
            });

            if (!response.ok) throw new Error('Failed to create user');

            this.hideAddUserModal();
            this.loadUsers();
            this.showToast('User created successfully', 'success');
        } catch (error) {
            this.showToast('Failed to create user', 'error');
        }
    }

    async updateUser() {
        const userId = document.getElementById('editUserId').value;
        const username = document.getElementById('editUsername').value;
        const password = document.getElementById('editPassword').value;
        const isAdmin = document.getElementById('editIsAdmin').checked;

        try {
            // Update user info
            await fetch(`${this.apiBaseUrl}/admin/users/${userId}`, {
                method: 'PUT',
                headers: auth.getAuthHeaders(),
                body: JSON.stringify({ username, isAdmin })
            });

            // Update password if provided
            if (password) {
                await fetch(`${this.apiBaseUrl}/admin/users/${userId}/password`, {
                    method: 'PUT',
                    headers: auth.getAuthHeaders(),
                    body: JSON.stringify({ password })
                });
            }

            this.hideEditUserModal();
            this.loadUsers();
            this.showToast('User updated successfully', 'success');
        } catch (error) {
            this.showToast('Failed to update user', 'error');
        }
    }

    async deleteUser(userId) {
        if (!confirm('Are you sure you want to delete this user?')) return;

        try {
            const response = await fetch(`${this.apiBaseUrl}/admin/users/${userId}`, {
                method: 'DELETE',
                headers: auth.getAuthHeaders()
            });

            if (!response.ok) throw new Error('Failed to delete user');

            this.loadUsers();
            this.showToast('User deleted successfully', 'success');
        } catch (error) {
            this.showToast('Failed to delete user', 'error');
        }
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
}

// Initialize the admin UI
const app = new AdminUI();