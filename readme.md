# 🚀 Go File Share API

## 🌟 Overview

The Go File Share API is a high-performance, secure file management system built with Go. It provides a robust, scalable solution for handling file operations over HTTPS with a clean, intuitive API. Built with clean architecture principles, it offers a flexible foundation that can be extended with different storage backends.

## 🎯 Key Features

### 🛡️ Security First
- **End-to-End HTTPS**: All communications are encrypted using TLS 1.3
- **Path Traversal Protection**: Built-in safeguards against directory traversal attacks
- **Basic Authentication**: Secure access control with username/password protection
- **CORS Protection**: Configurable CORS policies for web security

### 🚀 Performance Optimized
- **Efficient File Handling**: Stream-based processing for minimal memory usage
- **Concurrent Operations**: Handles multiple file operations efficiently
- **ZIP Streaming**: On-the-fly ZIP creation for folder downloads without temporary files

### 📁 Advanced File Operations
- **Directory Browsing**: Clean HTML interface with file details
- **Bulk Operations**: Upload/download multiple files or entire folders
- **On-Demand Zipping**: Download folders as ZIP archives with a single click
- **File Metadata**: View file sizes, modification dates, and types

### 🏗️ Clean Architecture
- **Modular Design**: Separated domain, application, and infrastructure layers
- **Dependency Injection**: Easy to test and maintain
- **Pluggable Storage**: Built with interfaces for easy storage backend swapping
- **Comprehensive Logging**: Built-in structured logging for monitoring and debugging

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Standard Library `net/http`
- **Authentication**: HTTP Basic Auth
- **TLS**: Built-in support with automatic certificate management
- **Testing**: Native Go testing with table-driven tests
- **Documentation**: OpenAPI 3.0 (Swagger) specification

## 📚 API Documentation

### Endpoints

#### 1. List Directory or Download File
```
GET /
```
- **Parameters**:
  - `path` (query, optional): Directory path to list or file to download
- **Responses**:
  - `200`: Directory listing (HTML) or file download
  - `401`: Authentication required
  - `403`: Forbidden (path traversal detected)
  - `404`: Path not found

#### 2. Upload Files/Folders
```
POST /upload
Content-Type: multipart/form-data
```
- **Parameters**:
  - `file` (form-data): File(s) to upload
  - `path` (form-data, optional): Target directory
- **Responses**:
  - `200`: Upload successful (HTML response)
  - `400`: Invalid request
  - `401`: Authentication required
  - `403`: Forbidden
  - `413`: Payload too large

#### 3. Health Check
```
GET /health
```
- **Responses**:
  - `200`: Service is healthy
  ```json
  {
    "status": "healthy"
  }
  ```

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- Valid TLS certificates (or use self-signed for development)

### Installation
1. Clone the repository
   ```bash
   git clone https://github.com/yourusername/go-file-share.git
   cd go-file-share/backend
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Configure environment variables
   ```bash
   export FILE_SHARE_ROOT=/path/to/storage
   export FILE_SHARE_USERNAME=admin
   export FILE_SHARE_PASSWORD=securepassword
   export TLS_CERT_FILE=path/to/cert.pem
   export TLS_KEY_FILE=path/to/key.pem
   ```

4. Run the server
   ```bash
   go run cmd/server/main.go
   ```

## 🛡️ Security Considerations

- Always use strong passwords
- Keep TLS certificates up to date
- Regularly audit file permissions
- Monitor access logs for suspicious activity
- Consider adding rate limiting in production

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ✨ What Makes It Unique

1. **Clean Architecture**: The codebase follows clean architecture principles, making it maintainable and testable.
2. **No External Dependencies**: Built using Go's standard library for maximum compatibility.
3. **Streaming Architecture**: Handles large files efficiently with minimal memory usage.
4. **Production Ready**: Includes health checks, proper error handling, and structured logging.
5. **Flexible Storage**: Easy to implement different storage backends (local filesystem, S3, etc.)
6. **Self-Contained**: No database required - perfect for simple deployments.

## 📞 Support

For support, please open an issue in the GitHub repository.
