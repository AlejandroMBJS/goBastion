# Quick Start Guide

Get up and running with Go Native FastAPI in 5 minutes.

## 1. Install & Run

```bash
# Install dependencies
go mod download

# Run the server
go run ./cmd/server
```

The server will start on http://localhost:8080

## 2. Test the API

### Option A: Use the Test Script
```bash
./test-api.sh
```

### Option B: Manual Testing with curl

#### Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "role": "user"
  }'
```

**Note**: If CSRF is enabled, you'll need to get a CSRF token first. See the test script for an example.

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

Save the `access_token` from the response.

#### Get Current User
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 3. Explore the API

- **Swagger UI**: http://localhost:8080/docs
- **OpenAPI JSON**: http://localhost:8080/docs/openapi.json
- **Admin Panel**: http://localhost:8080/admin (requires admin role)

## 4. Configuration

Edit `config/config.json` to change settings:

```json
{
  "server": {
    "port": ":8080"
  },
  "security": {
    "enable_csrf": false,  // Disable for easier testing
    "jwt_secret": "your-secret-key"
  }
}
```

## 5. Common Tasks

### Disable CSRF for Development
Set `enable_csrf: false` in `config/config.json`

### Create an Admin User
Register a user with `"role": "admin"`:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin",
    "email": "admin@example.com",
    "password": "adminpass123",
    "role": "admin"
  }'
```

### Access Admin Panel
1. Login as an admin user
2. Use the access token to access http://localhost:8080/admin

### Add a New Endpoint

1. Create a handler in `internal/app/router/`:
```go
func handleMyEndpoint(w http.ResponseWriter, r *http.Request, params map[string]string) {
    writeJSON(w, http.StatusOK, map[string]string{"message": "Hello!"})
}
```

2. Register the route:
```go
r.Handle("GET", "/api/v1/my-endpoint", handleMyEndpoint)
```

3. Restart the server

## 6. Database

The default SQLite database is created as `api.db` in the project root.

To view/edit the database:
```bash
sqlite3 api.db
```

Common queries:
```sql
-- View all users
SELECT * FROM users;

-- Count users
SELECT COUNT(*) FROM users;

-- Delete a user
DELETE FROM users WHERE id = 1;
```

## 7. Build for Production

```bash
# Build the binary
go build -o server ./cmd/server

# Run with environment variables
export APP_JWT_SECRET="production-secret-key"
export APP_DB_DSN="postgres://user:pass@localhost/prod_db"
./server
```

## 8. Docker

```bash
# Build
docker build -t go-native-fastapi .

# Run
docker run -p 8080:8080 go-native-fastapi
```

## 9. Troubleshooting

### "CSRF token missing" error
Either:
- Disable CSRF in config: `"enable_csrf": false`
- Use the test script which handles CSRF properly
- Get CSRF token from cookie and include in requests

### "Missing authorization header" error
Make sure to include the JWT token:
```bash
-H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Database locked error
Stop all running instances of the server and try again.

## 10. Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Review the [Security Hardening Checklist](README.md#security-hardening-checklist)
- Explore the code in `internal/app/router/` to see examples
- Check out the Swagger UI at http://localhost:8080/docs

## Architecture Overview

```
Framework (internal/framework/)     Application (internal/app/)
├── config/                         ├── router/
├── router/                         │   ├── auth.go
├── db/                             │   └── users.go
├── middleware/                     └── models/
├── security/                           └── user.go
├── admin/
├── docs/
└── view/
```

**Framework layer**: Rarely touched (router, DB, middleware, security)
**Application layer**: Where you work (routes, models, business logic)

---

**Happy coding!** 🚀
