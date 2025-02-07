You are an expert software architect and developer. Your task is to generate the code for a complete, functional file uploading website with the following specifications. The entire project must be designed to run from a single Docker Compose setup on one server.

**Project Goal:**

Create a file uploading website where users can:

*   Upload files of various types and sizes (implement reasonable size limits).
*   Select a storage duration for their uploaded files (options should include: 1 hour, 1 day, 7 days, 30 days, and permanent). Files should be automatically deleted after the chosen duration.
*   Password protect individual files. Users accessing a password-protected file should be presented with a simple password entry page.
*   Interact with a simple, clean, and user-friendly web interface. The interface should be mobile-responsive.
*   (Basic File Management) Provide download links for uploaded files. Optionally, for the uploader, provide a simple list of their uploaded files and an option to manually delete files before expiry.

**Technology Stack:**

*   **Frontend:**
    *   Languages: HTML, CSS, JavaScript (plain JavaScript is acceptable for simplicity, or a lightweight framework if preferred for structure).
    *   Web Server for Frontend: Nginx (Dockerized).
*   **Backend:**
    *   Language: Go.
    *   Web Framework (Optional but Recommended for structure):  Choose a lightweight Go framework if it simplifies development (e.g., Gin or Echo), or use `net/http` directly for a simpler approach.
    *   Task Scheduling: Implement file deletion scheduling within the Go backend application itself using Go's `time` package and Goroutines (for simplicity in a single-server setup).
    *   Dockerized Go application.
*   **Database:**
    *   Database: MongoDB (Dockerized).
    *   Use a MongoDB Go driver to interact with the database from the backend.
*   **File Storage:**
    *   Cloud Storage:  Assume the use of external cloud storage (like AWS S3, Google Cloud Storage, or Azure Blob Storage) for persistent file storage. Generate placeholder code and instructions for integrating with a cloud storage service. Do not implement actual cloud storage integration, focus on the structure and placeholders.
*   **Reverse Proxy:**
    *   Nginx (Dockerized) to serve the frontend and proxy API requests to the backend.

**Dockerization and Deployment:**

*   The entire application must be containerized using Docker.
*   Provide a `docker-compose.yml` file that defines and orchestrates all services: frontend (Nginx), backend (Go application), database (MongoDB), and reverse proxy (Nginx).
*   The Docker Compose setup should be designed for deployment on a single server.
*   Include Dockerfiles for building the frontend, backend, and reverse proxy images. Use official Docker images where possible (e.g., for Nginx and MongoDB).

**Output Requirements:**

Generate the following:

1.  **Frontend Code:** HTML, CSS, and JavaScript files for the user interface.
2.  **Backend Code (Go):** Go source code for the API, including:
    *   File upload endpoint.
    *   File download endpoint (with password protection logic).
    *   Endpoint for setting password protection.
    *   Logic for handling storage durations and scheduling file deletion.
    *   Database interaction code (using MongoDB Go driver).
    *   Placeholder code for cloud storage integration (uploading, downloading, deleting files from cloud storage -  clearly indicate where actual cloud storage SDK integration would be needed).
3.  **Dockerfiles:** Dockerfiles for the frontend, backend, and reverse proxy services.
4.  **`docker-compose.yml` file:**  A complete Docker Compose file to orchestrate all services.
5.  **README.md:** A basic README file with instructions on:
    *   How to build and run the application using Docker Compose.
    *   How to configure environment variables (for database connection, cloud storage placeholders, etc.).
    *   Basic usage instructions for the file uploading site.

**Important Considerations for Code Generation:**

*   **Simplicity First:** Prioritize clear, understandable, and functional code over highly optimized or complex solutions, especially for this single-server, Docker Compose setup.
*   **Error Handling:** Include basic error handling in both frontend and backend.
*   **Security:** Implement basic password hashing for password protection (using bcrypt in Go).  Highlight other security considerations (HTTPS, input validation, etc.) in the README.
*   **Placeholders:** For cloud storage integration, use clear placeholders and comments indicating where actual cloud storage SDK code would be implemented. Focus on the overall architecture and flow.
*   **Comments:**  Include comments in the code to explain the functionality of different parts.

Please generate all the necessary code files and the README.md based on these specifications.