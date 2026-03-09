# Inventory API

A Go-based Inventory API with user authentication using JWT and Gin.

## Setup
1.  Install Go.
2.  Install dependencies: `go mod tidy`
3.  Run the application: `go run cmd/api/main.go`

## API Endpoints

### 1. Register User (Public)
`POST /auth/register`
- Request:
```json
{
  "username": "admin",
  "email": "admin@example.com",
  "password": "password123"
}
```

### 2. Login User (Public)
`POST /auth/login`
- Request:
```json
{
  "username": "admin",
  "password": "password123"
}
```
- Response: Returns a JWT token.

### 3. User Profile (Protected)
`GET /api/profile`
- Headers: `Authorization: Bearer <TOKEN>`
- Response: User details.

## Dependencies
- Gin
- GORM (SQLite)
- Go-JWT
- Bcrypt