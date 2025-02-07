class FileHostUI {
    constructor() {
        this.dragDropArea = document.getElementById('dragDropArea');
        this.fileInput = document.getElementById('fileInput');
        this.uploadForm = document.getElementById('uploadForm');
        this.filesList = document.getElementById('filesList');
        this.progressOverlay = document.getElementById('uploadProgress');
        this.progressFill = document.querySelector('.progress-fill');
        this.progressText = document.querySelector('.progress-text');
        this.progressPercentage = document.querySelector('.progress-percentage');
        this.emptyState = document.querySelector('.empty-state');
        this.togglePasswordBtn = document.querySelector('.toggle-password');
        this.uploadButton = this.uploadForm.querySelector('.upload-button');
        this.progressCircle = document.querySelector('.progress-circle-fill');
        this.progressFilename = document.querySelector('.progress-filename');
        this.apiBaseUrl = window.location.hostname === 'localhost' 
            ? 'http://localhost:6002'
            : 'https://fileditch.ashank.tech';
        
        this.initializeEventListeners();
        this.loadExistingFiles();
        this.initializeInteractionMode();
    }

    initializeEventListeners() {
        // Drag and drop handlers with improved state management
        this.dragDropArea.addEventListener('click', () => this.fileInput.click());
        this.dragDropArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.dragDropArea.classList.add('dragover');
            this.dragDropArea.setAttribute('data-state', 'dragging');
        });
        this.dragDropArea.addEventListener('dragleave', () => {
            this.dragDropArea.classList.remove('dragover');
            this.dragDropArea.setAttribute('data-state', this.selectedFile ? 'selected' : 'idle');
        });
        this.dragDropArea.addEventListener('drop', (e) => {
            e.preventDefault();
            this.dragDropArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            if (files.length) this.handleFileSelection(files[0]);
        });

        // File input handler
        this.fileInput.addEventListener('change', (e) => {
            if (e.target.files.length) this.handleFileSelection(e.target.files[0]);
        });

        // Form submission
        this.uploadForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleUpload();
        });

        // Password toggle
        if (this.togglePasswordBtn) {
            this.togglePasswordBtn.addEventListener('click', () => this.togglePasswordVisibility());
        }

        // Add copy feedback
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-copy-feedback]')) {
                this.showCopyFeedback(e.target);
            }
        });
    }

    initializeInteractionMode() {
        // Detect keyboard vs mouse interaction
        window.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                document.body.classList.remove('using-mouse');
            }
        });

        window.addEventListener('mousedown', () => {
            document.body.classList.add('using-mouse');
        });

        window.addEventListener('touchstart', () => {
            document.body.classList.add('using-mouse');
        });

        // Add ripple effect for touch devices
        document.addEventListener('touchstart', (e) => {
            if (e.target.closest('.button-base')) {
                const button = e.target.closest('.button-base');
                const rect = button.getBoundingClientRect();
                const touch = e.touches[0];
                const x = touch.clientX - rect.left;
                const y = touch.clientY - rect.top;
                
                const ripple = document.createElement('div');
                ripple.className = 'ripple';
                ripple.style.left = `${x}px`;
                ripple.style.top = `${y}px`;
                
                button.appendChild(ripple);
                
                setTimeout(() => ripple.remove(), 600);
            }
        }, { passive: true });
    }

    togglePasswordVisibility() {
        const passwordInput = document.getElementById('password');
        const type = passwordInput.type === 'password' ? 'text' : 'password';
        passwordInput.type = type;
        this.togglePasswordBtn.querySelector('.icon').textContent = type === 'password' ? '👁️' : '👁️‍🗨️';
    }

    handleFileSelection(file) {
        this.selectedFile = file;
        this.dragDropArea.setAttribute('data-state', 'selected');
        
        const sizeMB = (file.size / (1024 * 1024)).toFixed(2);
        this.dragDropArea.innerHTML = `
            <div class="file-preview">
                <div class="file-icon">📄</div>
                <div class="file-details">
                    <strong>${file.name}</strong>
                    <span class="file-size">${sizeMB} MB</span>
                </div>
            </div>
            <p class="file-hint">Click or drag another file to change</p>
        `;
    }

    async handleUpload() {
        if (!this.selectedFile) {
            this.showToast('Please select a file first', 'error');
            return;
        }

        const formData = new FormData();
        formData.append('file', this.selectedFile);
        formData.append('duration', document.getElementById('duration').value);
        
        const password = document.getElementById('password').value;
        if (password) {
            formData.append('password', password);
        }

        try {
            this.setUploadState('loading');
            this.showProgress();
            
            const response = await fetch(`${this.apiBaseUrl}/upload`, {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) throw new Error('Upload failed');
            
            const result = await response.json();
            this.hideProgress();
            this.setUploadState('success');
            this.showUploadSuccess(result);
            this.loadExistingFiles();
            
            // Reset after success animation
            setTimeout(() => {
                this.resetForm();
                this.setUploadState('idle');
            }, 2000);
        } catch (error) {
            this.hideProgress();
            this.setUploadState('error');
            this.showToast(error.message, 'error');
            
            // Reset after error animation
            setTimeout(() => {
                this.setUploadState('idle');
            }, 2000);
        }
    }

    setUploadState(state) {
        this.uploadButton.dataset.state = state;
        this.uploadButton.disabled = state === 'loading';
        
        // Update button text based on state
        const buttonText = {
            idle: 'Upload File',
            loading: 'Uploading...',
            success: 'Uploaded!',
            error: 'Failed!'
        }[state];
        
        this.uploadButton.querySelector('.button-text').textContent = buttonText;
        
        // Add smooth icon transitions
        const icon = this.uploadButton.querySelector('.button-icon');
        if (state === 'success') {
            icon.innerHTML = `
                <svg viewBox="0 0 24 24" width="14" height="14">
                    <path fill="currentColor" d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                </svg>
            `;
        } else if (state === 'error') {
            icon.innerHTML = `
                <svg viewBox="0 0 24 24" width="14" height="14">
                    <path fill="currentColor" d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                </svg>
            `;
        } else if (state === 'idle') {
            icon.innerHTML = '→';
        }
    }

    showProgress() {
        this.progressOverlay.classList.remove('hidden');
        this.progressFilename.textContent = this.selectedFile.name;
        requestAnimationFrame(() => {
            this.progressOverlay.style.opacity = '1';
        });
    }

    hideProgress() {
        this.progressOverlay.style.opacity = '0';
        setTimeout(() => {
            this.progressOverlay.classList.add('hidden');
        }, 300);
    }

    updateProgress(percent) {
        const circumference = 2 * Math.PI * 15.9155; // radius of circle
        const offset = circumference - (percent / 100) * circumference;
        this.progressCircle.style.strokeDashoffset = offset;
        this.progressPercentage.textContent = `${percent}%`;
    }

    async loadExistingFiles() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/files`);
            if (!response.ok) throw new Error('Failed to load files');
            
            const files = await response.json();
            this.renderFilesList(files);
        } catch (error) {
            console.error('Failed to load files:', error);
            this.showToast('Failed to load files', 'error');
        }
    }

    renderFilesList(files) {
        if (files.length === 0) {
            this.emptyState.classList.remove('hidden');
            this.filesList.innerHTML = '';
            return;
        }

        this.emptyState.classList.add('hidden');
        this.filesList.innerHTML = '';
        
        files.forEach(file => {
            const fileElement = document.createElement('div');
            fileElement.className = 'file-item';
            fileElement.innerHTML = `
                <div class="file-info">
                    <strong>${file.name}</strong>
                    <p class="file-meta">Expires: ${this.formatExpiry(file.expiryDate)}</p>
                </div>
                <div class="file-actions">
                    <button class="action-button" data-copy-feedback="Copied!" onclick="app.copyLink('${file.id}')">
                        <span class="button-content">
                            <svg class="button-icon" viewBox="0 0 24 24" width="14" height="14">
                                <path fill="currentColor" d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                            </svg>
                            <span>Copy Link</span>
                        </span>
                    </button>
                    <button class="action-button delete" onclick="app.deleteFile('${file.id}')">
                        <span class="button-content">
                            <svg class="button-icon" viewBox="0 0 24 24" width="14" height="14">
                                <path fill="currentColor" d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                            </svg>
                            <span>Delete</span>
                        </span>
                    </button>
                </div>
            `;
            this.filesList.appendChild(fileElement);
        });
    }

    formatExpiry(date) {
        if (!date) return 'Never';
        const expiry = new Date(date);
        const now = new Date();
        const diff = expiry - now;
        
        // If less than a day
        if (diff < 86400000) {
            const hours = Math.ceil(diff / 3600000);
            return `${hours} hour${hours !== 1 ? 's' : ''} left`;
        }
        
        // If less than a week
        if (diff < 604800000) {
            const days = Math.ceil(diff / 86400000);
            return `${days} day${days !== 1 ? 's' : ''} left`;
        }
        
        return expiry.toLocaleDateString();
    }

    async copyLink(fileId) {
        const link = `${this.apiBaseUrl}/file/${fileId}`;
        try {
            await navigator.clipboard.writeText(link);
            this.showToast('Link copied to clipboard!', 'success');
        } catch (error) {
            this.showToast('Failed to copy link', 'error');
        }
    }

    showCopyFeedback(button) {
        const originalContent = button.innerHTML;
        const feedback = button.dataset.copyFeedback;
        
        button.setAttribute('data-state', 'success');
        button.innerHTML = `
            <span class="button-content">
                <svg class="button-icon" viewBox="0 0 24 24" width="14" height="14">
                    <path fill="currentColor" d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                </svg>
                <span>${feedback}</span>
            </span>
        `;
        
        setTimeout(() => {
            button.setAttribute('data-state', 'idle');
            button.innerHTML = originalContent;
        }, 2000);
    }

    async deleteFile(fileId) {
        if (!confirm('Are you sure you want to delete this file?')) return;
        
        try {
            const response = await fetch(`${this.apiBaseUrl}/files/${fileId}`, {
                method: 'DELETE'
            });
            if (!response.ok) throw new Error('Delete failed');
            
            this.loadExistingFiles();
            this.showToast('File deleted successfully', 'success');
        } catch (error) {
            this.showToast('Failed to delete file', 'error');
        }
    }

    resetForm() {
        this.uploadForm.reset();
        this.selectedFile = null;
        this.dragDropArea.setAttribute('data-state', 'idle');
        this.dragDropArea.innerHTML = `
            <div class="upload-icon">📁</div>
            <p>Drag & drop files here or <span class="highlight">click to select</span></p>
            <p class="file-hint">Supports any file type</p>
        `;
    }

    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        
        document.body.appendChild(toast);
        
        // Trigger animation
        requestAnimationFrame(() => {
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
                setTimeout(() => toast.remove(), 300);
            }, 3000);
        });
    }

    showUploadSuccess(result) {
        const link = `${this.apiBaseUrl}/file/${result.id}`;
        this.showToast('File uploaded successfully!', 'success');
        navigator.clipboard.writeText(link)
            .then(() => this.showToast('Link copied to clipboard!', 'success'))
            .catch(() => {});
    }
}

// Initialize the application
const app = new FileHostUI();