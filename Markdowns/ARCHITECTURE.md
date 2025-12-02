# Architecture Documentation

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Request                             │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Core Router                                 │
│                  (internal/framework/router)                     │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Middleware Stack                              │
│                                                                   │
│  1. RequestID      → Generates unique request IDs                │
│  2. Logging        → Logs method, path, status, duration         │
│  3. Recover        → Catches panics, returns 500                 │
│  4. WithTimeout    → Request timeout context                     │
│  5. SecurityHeaders → X-Frame-Options, HSTS, CSP, etc.          │
│  6. MaxBodySize    → Limits request body size                    │
│  7. CORS           → Cross-Origin Resource Sharing               │
│  8. RateLimit      → Per-IP rate limiting (optional)             │
│  9. CSRF           → Double-submit cookie (optional)             │
│  10. JWTAuth       → JWT token validation (optional)             │
│  11. RequireRole   → Role-based access (per-route)               │
│                                                                   │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Route Handlers                                │
│                 (internal/app/router)                            │
│                                                                   │
│  Auth Routes:         User Routes:        Admin Routes:         │
│  - register           - list              - dashboard            │
│  - login              - create            - user list            │
│  - refresh            - get               - user detail          │
│  - me                 - update            - user update          │
│                       - delete                                   │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                │                             │
                ▼                             ▼
┌────────────────────────────┐  ┌───────────────────────────────┐
│    Database Layer          │  │    Template Engine            │
│  (internal/framework/db)   │  │  (internal/framework/view)    │
│                            │  │                               │
│  - Connection pooling      │  │  - PHP-like syntax            │
│  - Query builder           │  │  - Auto HTML escaping         │
│  - Migrations              │  │  - File-based templates       │
│  - CRUD helpers            │  │                               │
└────────────┬───────────────┘  └───────────────────────────────┘
             │
             ▼
┌────────────────────────────┐
│       SQLite Database       │
│     (or Postgres/MySQL)     │
└─────────────────────────────┘
```

## Request Flow

### 1. Unauthenticated Request (e.g., Register)
```
Client Request
    │
    ├─> Router: Match route pattern
    │
    ├─> Middleware: RequestID (add UUID)
    ├─> Middleware: Logging (start timer)
    ├─> Middleware: Recover (defer panic handler)
    ├─> Middleware: WithTimeout (5s context)
    ├─> Middleware: SecurityHeaders (add headers)
    ├─> Middleware: MaxBodySize (limit 1MB)
    ├─> Middleware: CORS (check origin)
    ├─> Middleware: RateLimit (check IP limit)
    ├─> Middleware: CSRF (validate for POST)
    ├─> Middleware: JWTAuth (skip for /auth/register)
    │
    ├─> Handler: handleRegister
    │   ├─> Decode JSON body
    │   ├─> Validate input (name, email, password)
    │   ├─> Check if user exists
    │   ├─> Hash password (bcrypt)
    │   ├─> Insert into database
    │   ├─> Generate JWT tokens
    │   └─> Return response
    │
    └─> Response (201 Created)
```

### 2. Authenticated Request (e.g., Get User)
```
Client Request + Bearer Token
    │
    ├─> Router: Match route pattern
    │
    ├─> Middleware Stack (1-9)
    │
    ├─> Middleware: JWTAuth
    │   ├─> Extract Bearer token from header
    │   ├─> Parse JWT (HS256)
    │   ├─> Validate signature
    │   ├─> Check expiration
    │   ├─> Store claims in context
    │   └─> Continue (or 401 if invalid)
    │
    ├─> Handler: handleGetUser
    │   ├─> Get user ID from path params
    │   ├─> Query database
    │   └─> Return user JSON
    │
    └─> Response (200 OK)
```

### 3. Admin Request (e.g., Admin Panel)
```
Client Request + Admin JWT Token
    │
    ├─> Router: Match route pattern
    │
    ├─> Middleware Stack (1-10)
    │
    ├─> Middleware: RequireRole("admin")
    │   ├─> Get claims from context
    │   ├─> Check role == "admin"
    │   └─> Continue (or 403 if not admin)
    │
    ├─> Handler: handleAdminUsers
    │   ├─> Query all users from database
    │   ├─> Render HTML template
    │   └─> Return HTML
    │
    └─> Response (200 OK, text/html)
```

## Component Interactions

### Authentication Flow
```
┌──────────┐  Register   ┌──────────┐  Hash Pass  ┌──────────┐
│  Client  │ ────────────>│ Auth     │ ──────────> │ bcrypt   │
│          │              │ Handler  │             │          │
└──────────┘              └─────┬────┘             └──────────┘
     ▲                          │
     │                          ▼
     │                    ┌──────────┐
     │    JWT Tokens      │ Database │
     └────────────────────│  Insert  │
                          └──────────┘

Login Flow:
┌──────────┐  Login      ┌──────────┐  Verify     ┌──────────┐
│  Client  │ ──────────> │ Auth     │ ──────────> │ bcrypt   │
│          │             │ Handler  │             │ Compare  │
└──────────┘             └─────┬────┘             └──────────┘
     ▲                         │
     │                         ▼
     │                   ┌──────────┐
     │   JWT Tokens      │ Database │
     └───────────────────│  Query   │
                         └──────────┘

Authenticated Request:
┌──────────┐  Bearer     ┌──────────┐  Validate   ┌──────────┐
│  Client  │ ──────────> │ JWT      │ ──────────> │ HMAC     │
│          │  Token      │ Auth MW  │             │ SHA256   │
└──────────┘             └─────┬────┘             └──────────┘
                               │
                               ▼
                         Store claims
                         in context
                               │
                               ▼
                         Continue to
                         handler
```

### Database Query Flow
```
┌──────────────────┐
│   Handler        │
│   (High-level)   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Database        │     Option A: Direct SQL
│  Abstraction     │ ──> db.GetUser(ctx, id)
│  (db/db.go)      │
└────────┬─────────┘     Option B: Query Builder
         │            ┌> qb.NewQB("users")
         │            │    .WhereEq("id", id)
         │            │    .BuildSelect()
         │            └> db.DB.QueryContext(...)
         ▼
┌──────────────────┐
│  database/sql    │
│  (stdlib)        │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  SQLite Driver   │
│  (go-sqlite3)    │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  SQLite DB File  │
│  (api.db)        │
└──────────────────┘
```

## Security Architecture

### Defense in Depth
```
Layer 1: Network
  - HTTPS (TLS) in production
  - Firewall rules
  - DDoS protection

Layer 2: Application (This Framework)
  - Rate limiting (per IP)
  - CORS (origin validation)
  - CSRF (double-submit cookie)
  - Max body size
  - Request timeouts

Layer 3: Authentication
  - JWT tokens (HS256)
  - Password hashing (bcrypt)
  - Token expiration
  - Refresh token rotation

Layer 4: Authorization
  - Role-based access control
  - Per-route permission checks
  - Context-based claims

Layer 5: Data
  - Parameterized queries (SQL injection prevention)
  - HTML escaping (XSS prevention)
  - Input validation
  - Database encryption (optional)
```

### JWT Token Structure
```
┌─────────────────────────────────────────────────────────┐
│                      JWT Token                          │
├─────────────────────────────────────────────────────────┤
│ Header (base64url)                                      │
│ {                                                       │
│   "alg": "HS256",                                       │
│   "typ": "JWT"                                          │
│ }                                                       │
├─────────────────────────────────────────────────────────┤
│ Payload (base64url)                                     │
│ {                                                       │
│   "sub": "123",         // User ID                      │
│   "role": "admin",      // User role                    │
│   "iat": 1234567890,    // Issued at                    │
│   "exp": 1234568790     // Expires at (15 min)          │
│ }                                                       │
├─────────────────────────────────────────────────────────┤
│ Signature (HMAC-SHA256)                                 │
│ HMAC-SHA256(                                            │
│   base64url(header) + "." + base64url(payload),         │
│   secret                                                │
│ )                                                       │
└─────────────────────────────────────────────────────────┘
```

### CSRF Protection
```
Double-Submit Cookie Pattern:

1. Initial GET Request:
   Server generates random token
   → Sets cookie: csrf_token=ABC123
   → Client stores cookie

2. POST/PUT/DELETE Request:
   Client sends:
   → Cookie: csrf_token=ABC123
   → Header: X-CSRF-Token: ABC123

3. Server Validation:
   → Reads both cookie and header
   → Validates: cookie == header
   → Validates: HMAC signature (optional)
   → Allow or reject (403)
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL,
    email           TEXT NOT NULL UNIQUE,
    role            TEXT NOT NULL,
    is_active       INTEGER NOT NULL DEFAULT 1,
    is_staff        INTEGER NOT NULL DEFAULT 0,
    is_superuser    INTEGER NOT NULL DEFAULT 0,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
```

### User Roles & Permissions
```
┌──────────────────────────────────────────────────────────┐
│ Role         │ is_staff │ is_superuser │ Permissions     │
├──────────────────────────────────────────────────────────┤
│ user         │    0     │      0       │ - View own data │
│              │          │              │ - Update own    │
│              │          │              │                 │
│ admin        │    1     │      1       │ - All of above  │
│              │          │              │ - Manage users  │
│              │          │              │ - Admin panel   │
│              │          │              │ - System config │
└──────────────────────────────────────────────────────────┘
```

## Configuration Hierarchy

```
┌─────────────────────────────────────────────────────────┐
│                    Configuration                         │
├─────────────────────────────────────────────────────────┤
│ 1. Hardcoded Defaults (in code)                         │
│    ↓ Overridden by                                      │
│ 2. config/config.json                                    │
│    ↓ Overridden by                                      │
│ 3. Environment Variables                                 │
│    (APP_PORT, APP_JWT_SECRET, etc.)                     │
└─────────────────────────────────────────────────────────┘

Priority: ENV VARS > config.json > Defaults
```

## Error Handling Strategy

```
┌──────────────────────────────────────────────────────────┐
│                    Error Types                           │
├──────────────────────────────────────────────────────────┤
│ 1. Validation Errors (400 Bad Request)                   │
│    - Invalid JSON                                        │
│    - Missing required fields                             │
│    - Format errors (email, etc.)                         │
│                                                          │
│ 2. Authentication Errors (401 Unauthorized)              │
│    - Missing token                                       │
│    - Invalid token                                       │
│    - Expired token                                       │
│                                                          │
│ 3. Authorization Errors (403 Forbidden)                  │
│    - Insufficient permissions                            │
│    - CSRF validation failed                              │
│                                                          │
│ 4. Not Found Errors (404 Not Found)                     │
│    - Resource doesn't exist                              │
│    - Invalid ID                                          │
│                                                          │
│ 5. Conflict Errors (409 Conflict)                       │
│    - User already exists                                 │
│    - Duplicate entry                                     │
│                                                          │
│ 6. Rate Limit Errors (429 Too Many Requests)            │
│    - Exceeded rate limit                                 │
│                                                          │
│ 7. Server Errors (500 Internal Server Error)            │
│    - Database errors                                     │
│    - Unhandled panics (caught by Recover middleware)    │
│    - Unexpected errors                                   │
└──────────────────────────────────────────────────────────┘
```

## Scalability Considerations

### Current Architecture (Single Instance)
```
┌────────────┐
│   Client   │
└─────┬──────┘
      │
      ▼
┌────────────┐    ┌──────────┐
│   Server   │───>│ SQLite   │
│  (1 inst)  │    │ Database │
└────────────┘    └──────────┘

Limitations:
- Single point of failure
- Limited to vertical scaling
- In-memory rate limiting (per instance)
- Local session storage
```

### Production Architecture (Recommended)
```
┌────────────┐
│   Client   │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ Load       │
│ Balancer   │
└─────┬──────┘
      │
      ├─> ┌────────────┐
      │   │  Server 1  │─┐
      │   └────────────┘ │
      │                  │
      ├─> ┌────────────┐ │    ┌──────────┐
      │   │  Server 2  │─┼───>│ Postgres │
      │   └────────────┘ │    │  (RDS)   │
      │                  │    └──────────┘
      └─> ┌────────────┐ │
          │  Server 3  │─┘
          └────────────┘
                │
                ▼
          ┌──────────┐
          │  Redis   │ (Rate limiting, sessions)
          └──────────┘

Improvements:
- Horizontal scaling
- High availability
- Distributed rate limiting
- Database replication
```

## Performance Characteristics

```
┌─────────────────────────────────────────────────────────┐
│ Component           │ Performance Notes                 │
├─────────────────────────────────────────────────────────┤
│ Router              │ O(n) route matching               │
│                     │ (n = number of routes)            │
│                     │ Optimize: Use trie or radix tree  │
│                     │                                   │
│ Middleware          │ Chain execution (sequential)      │
│                     │ Each adds ~10-50μs overhead       │
│                     │                                   │
│ JWT Validation      │ ~100-200μs per request            │
│                     │ (HMAC-SHA256 + base64)            │
│                     │                                   │
│ Database Queries    │ SQLite: ~1-10ms (local disk)      │
│                     │ Postgres: ~5-50ms (network)       │
│                     │ Connection pooling helps          │
│                     │                                   │
│ JSON Encoding       │ ~50-500μs (depends on size)       │
│                     │ Standard library is fast enough   │
│                     │                                   │
│ Template Rendering  │ ~1-5ms (parse + execute)          │
│                     │ Cache parsed templates for speed  │
│                     │                                   │
│ Rate Limiting       │ ~10-20μs (in-memory lookup)       │
│                     │ Use Redis for distributed         │
└─────────────────────────────────────────────────────────┘

Expected Throughput:
- Simple GET: 10,000-50,000 req/s (single instance)
- Auth + DB:  1,000-5,000 req/s (single instance)
- HTML render: 500-2,000 req/s (single instance)
```

## Development Workflow

```
1. Edit Code
   ↓
2. go run ./cmd/server (automatic recompile)
   ↓
3. Test with curl or ./test-api.sh
   ↓
4. View logs in terminal
   ↓
5. Iterate
   ↓
6. go build (production binary)
   ↓
7. Deploy
```

## Monitoring & Observability

### Current Logging
```
[RequestID] Method Path - Status - Duration
[abc123] POST /api/v1/auth/login - 200 - 45ms
```

### Recommended Production Additions
- Structured logging (JSON format)
- Log aggregation (ELK, Splunk, CloudWatch)
- Metrics (Prometheus)
- Tracing (Jaeger, Zipkin)
- Health checks
- Profiling endpoints (pprof)

---

**This architecture is designed for clarity, maintainability, and production readiness.**
