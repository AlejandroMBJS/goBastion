# goBastion

goBastion is a modern, production-ready Go framework designed for rapid development of web applications and APIs. This repository serves as a template for new projects, providing a solid foundation with a pre-configured structure, authentication, database integration, and a powerful custom template engine.

## ✨ What's New

- **🎨 Modern Template Syntax**: Clean, Go-like template syntax with `go::` / `@` constructs
- **💅 Tailwind-Styled Templates**: Beautiful, responsive UI out of the box
- **🏠 Professional Home Page**: Next.js/Django-style boilerplate landing page at `/`
- **🔒 Security First**: Auto HTML escaping, CSRF protection, JWT auth, rate limiting
- **👥 Complete Admin CRUD**: Full user management with create, read, update, delete
- **✅ Comprehensive Tests**: Full test coverage for template engine and routes

## 📖 Philosophy & Architecture

### Design Principles

goBastion is built on these core principles:

1. **Security by Default**: CSRF protection, XSS prevention, rate limiting, and JWT authentication are enabled out of the box
2. **Convention over Configuration**: Sensible defaults with the ability to customize through a centralized JSON configuration
3. **Developer Experience First**: Clean, intuitive template syntax that feels natural to Go developers
4. **Production Ready**: Built with real-world production needs in mind - not just a toy framework
5. **Minimal Magic**: Explicit over implicit - you should understand what's happening under the hood

### Framework Architecture

goBastion follows a clean separation between **framework core** and **application layer**:

```
goBastion/
├── internal/framework/     ← Framework core (stable, rarely modified)
│   ├── config/             - Configuration management
│   ├── db/                 - Database abstractions & query builder
│   ├── middleware/         - HTTP middlewares (CSRF, auth, rate limiting)
│   ├── router/             - HTTP router
│   ├── security/           - Security primitives (JWT, CSRF tokens)
│   └── view/               - Template engine (go:: / @ syntax)
│
├── internal/app/           ← Application layer (customize freely)
│   ├── models/             - Your data models
│   ├── router/             - Your route handlers
│   └── chat/               - Example: real-time chat feature
│
├── templates/              ← Your HTML templates (go:: / @ syntax)
│   ├── home.html           - Landing page
│   ├── auth/               - Login, register
│   └── admin/              - Admin dashboard & CRUD
│
└── config/                 ← Configuration
    └── config.json         - Centralized JSON configuration
```

**Key Concept**: The `internal/framework/` directory contains stable, battle-tested code that you should **rarely modify**. The `internal/app/` and `templates/` directories are where you build your application.

### Template Engine Architecture

The custom template engine (`internal/framework/view/view.go`) preprocesses templates:

```
Your Template (go:: / @)  →  Preprocessor  →  Go html/template  →  Safe HTML Output
   ↓                             ↓                    ↓
go:: if .User              {{ if .User          <p>Hello John</p>
  <p>Hello @.User.Name</p>   <p>Hello {{.User.Name}}</p>
::end                        {{ end }}
```

This gives you a clean syntax while maintaining Go's template security (auto HTML escaping, type safety).

## 🚀 Quickstart

Get the demo running in 3 steps:

### 1. Build and Run

```bash
# Build the server
go build -o gobastion-server ./cmd/server

# Run it
./gobastion-server
```

The server starts on `http://localhost:8080`

### 2. Create an Admin User

In a new terminal, use the registration endpoint or create directly via SQL:

```bash
# Option A: Via API (register then promote to admin)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin User","email":"admin@example.com","password":"admin123","role":"admin"}'

# Option B: Direct SQL (SQLite)
sqlite3 app.db "INSERT INTO users (name, email, role, password_hash, is_active, is_staff, is_superuser) VALUES ('Admin', 'admin@example.com', 'admin', '\$2a\$10\$...', 1, 1, 1);"
```

**Default Test Credentials (for demo):**
- Email: `admin@example.com`
- Password: `admin123`

*Note: Remember to change these in production!*

### 3. Explore the Demo

| URL | Description |
|-----|-------------|
| `http://localhost:8080/` | Beautiful landing page |
| `http://localhost:8080/login` | Login page (use credentials above) |
| `http://localhost:8080/register` | User registration |
| `http://localhost:8080/admin` | Admin dashboard (requires login) |
| `http://localhost:8080/admin/users` | User CRUD management |
| `http://localhost:8080/docs` | Interactive API documentation (Swagger) |
| `http://localhost:8080/api/v1/users` | REST API endpoints |

### What You Get Out of the Box

✅ **Complete Authentication Flow**
- Login / Logout / Registration
- JWT-based auth with HTTP-only cookies
- Role-based access control (admin, user)

✅ **Full Admin Panel**
- User list with search and filters
- Create new users
- Edit user details and permissions
- Delete users with confirmation

✅ **Security Features**
- CSRF protection on all forms
- XSS prevention via auto-escaping
- Rate limiting
- Security headers (CSP, X-Frame-Options, etc.)

✅ **Developer Experience**
- Custom template engine with clean syntax
- Query builder for database operations
- Hot reload with Air (if configured)
- Comprehensive error handling

## Project Generator CLI

A command-line interface (CLI) tool, `go-bastion`, has been developed to streamline the creation of new projects based on this template. This CLI automates the process of cloning the repository, configuring module names, and setting up the initial project structure.

### How to Build and Install the CLI

To build and install the `go-bastion` CLI, making it available globally in your system's PATH, execute the following command from the root of this repository:

```bash
go install ./cmd/go-bastion
```text

This will compile the CLI and place the `go-bastion` executable in your Go binary path (e.g., `$GOPATH/bin` or `$HOME/go/bin`), allowing you to run it from any directory.

### How to Use the CLI

Once installed, you can use the `go-bastion` CLI in two modes:

#### 1. Non-Interactive Mode

To create a new project with a specified name directly, use the following command:

```bash
go-bastion my-new-app
```text

Replace `my-new-app` with your desired project name. The CLI will clone the template into a new directory named `my-new-app`, configure the Go module, and prepare the project for development.

#### 2. Interactive Mode

If you prefer to be prompted for the project name, simply run the CLI without any arguments:

```bash
go-bastion
```text

The CLI will then ask you: `¿Cómo quieres llamar a tu nuevo proyecto? >` (What do you want to name your new project?). Enter your desired project name, and the CLI will proceed with the project generation.

### What the CLI Does

The `go-bastion` CLI performs the following automated steps:

1.  **Clones the Template Repository**: Fetches the latest version of this repository into your specified project directory.
2.  **Removes Git History**: Deletes the `.git` directory from the new project, ensuring a clean start for your version control.
3.  **Removes Generator Source**: Eliminates the `cmd/go-bastion` directory from the new project, as it's part of the generator itself and not the generated application.
4.  **Replaces Module Names**: Updates all occurrences of the original Go module name (`go-native-fastapi`) with your new project's module name (e.g., `github.com/AlejandroMBJS/my-new-app`) in `go.mod` and all `.go` files.
5.  **Adds Replace Directive**: Modifies the `go.mod` file to include a `replace` directive, allowing the Go toolchain to correctly resolve local module paths during development.
6.  **Runs `go mod tidy`**: Executes `go mod tidy` in the new project directory to synchronize dependencies and clean up `go.mod` and `go.sum`.

After these steps, your new Go project will be ready for development, with its own independent module and a clean slate for version control.

## 🎨 Template Engine

goBastion includes a powerful custom template engine with a clean, modern syntax. The engine uses **only two constructs**:

### 1. Echo Expressions (`@expr`)

Output values with automatic HTML escaping:

```html
<h1>@.Title</h1>
<p>Hello, @user.Name!</p>
<p>Email: @user.Email</p>
```text

### 2. Logic Blocks (`go:: ... ::end`)

Control flow using real Go code:

```html
go:: if user != nil
  <p>Welcome, @user.Name!</p>
::end

go:: range .Items
  <li>@.Name - @.Price</li>
::end
```

### Complete Example

```html
<h1>@.Title</h1>

go:: if .Error
<div class="error">@.Error</div>
::end

go:: if .Users
<table>
  go:: range .Users
  <tr>
    <td>@.ID</td>
    <td>@.Name</td>
    <td>
      go:: if eq .Role "admin"
      <span>Admin</span>
      go:: else
      <span>User</span>
      ::end
    </td>
  </tr>
  ::end
</table>
go:: else
<p>No users found</p>
::end
```text

### Key Features

- ✅ **Auto HTML Escaping**: All `@expr` outputs are automatically escaped
- ✅ **Clean Syntax**: Only two constructs to learn (`go::` and `@`)
- ✅ **Type Safe**: Compiles to Go's `html/template`
- ✅ **Backward Compatible**: Old PHP-style tags still work (deprecated)
- ✅ **Security First**: Built-in CSRF protection and XSS prevention

### Full Documentation

For complete details on the template syntax, including advanced features, best practices, and security guidelines, see:

📚 **[TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md)** - Complete template syntax reference

## ⚙️ Configuration

goBastion uses a centralized JSON configuration file at `config/config.json` that controls all major framework and application settings.

### Configuration File Structure

```json
{
  "app": {
    "name": "goBastion",
    "environment": "development",
    "base_url": "http://localhost:8080",
    "locale": "en-US",
    "description": "Modern Go web framework"
  },
  "server": {
    "host": "0.0.0.0",
    "port": ":8080",
    "read_timeout_seconds": 10,
    "write_timeout_seconds": 10,
    "idle_timeout_seconds": 60,
    "allowed_origins": ["http://localhost:3000"]
  },
  "database": {
    "driver": "sqlite3",
    "dsn": "file:api.db?_foreign_keys=on",
    "max_open_conns": 10,
    "max_idle_conns": 5,
    "conn_max_lifetime_minutes": 30
  },
  "security": {
    "enable_csrf": true,
    "csrf_cookie_name": "csrf_token",
    "enable_jwt": true,
    "jwt_secret": "change-me-in-prod",
    "access_token_minutes": 15
  },
  "frontend": {
    "theme": {
      "primary_color": "indigo",
      "secondary_color": "purple",
      "dark_mode": false
    }
  },
  "admin": {
    "enable_dashboard_metrics": true,
    "registration_open": true
  },
  "features": {
    "enable_chat": true,
    "enable_notifications": false
  }
}
```

### Configuration Sections

| Section | Purpose | Safe to Modify |
|---------|---------|----------------|
| `app` | Application name, environment, branding | ✅ Yes |
| `server` | HTTP server settings (port, timeouts) | ✅ Yes |
| `database` | Database driver and connection settings | ✅ Yes |
| `security` | CSRF, JWT, authentication settings | ⚠️ **Carefully** |
| `frontend` | Theme colors, UI preferences | ✅ Yes |
| `admin` | Admin panel feature flags | ✅ Yes |
| `features` | Application feature flags | ✅ Yes |

### Environment Variable Overrides

The config system supports environment variable overrides:

```bash
export APP_PORT=":9000"
export APP_DB_DRIVER="postgres"
export APP_DB_DSN="postgres://user:pass@localhost/dbname"
export APP_JWT_SECRET="your-secret-key"

./gobastion-server  # Will use environment variables
```

### Using Config in Your Code

**In handlers:**

```go
import "github.com/AlejandroMBJS/goBastion/internal/framework/config"

// Access via loaded config
cfg, _ := config.Load("config/config.json")
appName := cfg.App.Name
environment := cfg.App.Environment
```

**In templates:**

Pass config values through handler data:

```go
data := map[string]any{
    "Title": "My Page",
    "AppName": fullConfig.App.Name,
    "Environment": fullConfig.App.Environment,
}
views.Render(w, "template", data)
```

Then in your template:

```html
<title>@.AppName - @.Title</title>

go:: if eq .Environment "development"
  <div class="debug-banner">Development Mode</div>
::end
```

### ⚠️ Security Configuration Warning

**CRITICAL**: Always change these values in production:

- `security.jwt_secret` - Use a strong, random secret (minimum 32 characters)
- `app.environment` - Set to `"production"`
- `database.dsn` - Use proper production credentials
- `server.allowed_origins` - Whitelist only your actual domains

**Never commit production secrets to Git!** Use environment variables for sensitive values in production.

## 🎨 Editor Support

Get syntax highlighting for goBastion templates in your favorite editor!

### Vim/Neovim (using `gBCode`)

Install syntax highlighting with our pre-built binary:

```bash
# Build the installer
go build -o gBCode ./cmd/gobastion-vim-installer/

# Run the installer
./gBCode

# Or build and run in one step
go run ./cmd/gobastion-vim-installer/
```text

The installer will:
- ✅ Auto-detect your Vim/Neovim installation
- ✅ Install syntax files to the correct location
- ✅ Support `.gb.html`, `.gobastion.html`, `.bastion.html`, and `.gb.tmpl` files
- ✅ Provide syntax highlighting for `go::`, `::end`, and `@` expressions
- ✅ Include embedded Go syntax in logic blocks

**Manual installation** (alternative):
```bash
# For Neovim/LazyVim users
cd goBastionTemplates
./gBTemplatesNvim.sh
```text

📖 **[Full Vim/Neovim Guide](cmd/gobastion-vim-installer/README.md)**

### VS Code (using `.vsix` extension)

Install the pre-packaged VS Code extension:

#### Option 1: Install from Pre-built VSIX

```bash
# The extension is already packaged
code --install-extension goBastionTemplates/gobastion-templates-0.1.0.vsix
```text

**Or via VS Code UI:**
1. Open VS Code
2. Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on macOS)
3. Type "Extensions: Install from VSIX..."
4. Select `goBastionTemplates/gobastion-templates-0.1.0.vsix`
5. Reload VS Code

#### Option 2: Build from Source

```bash
# Install vsce (if not already installed)
npm install -g @vscode/vsce

# Package the extension
cd goBastionTemplates
vsce package

# Install the generated .vsix file
code --install-extension gobastion-templates-0.1.0.vsix
```text

**Features:**
- ✅ Syntax highlighting for all goBastion constructs
- ✅ Auto-closing pairs for brackets and tags
- ✅ Embedded Go syntax in logic blocks
- ✅ Support for `.gb.html`, `.gobastion.html`, `.bastion.html`, and `.gb.tmpl`

📖 **[Full VS Code Extension Guide](goBastionTemplates/README.md)**

### Supported File Extensions

Both editors support these file extensions:
- `*.gb.html` - Primary goBastion template extension (recommended)
- `*.gobastion.html` - Alternative extension
- `*.bastion.html` - Alternative extension
- `*.gb.tmpl` - Template extension

## 🏠 Home Page & Routes

Once you start the server with `go run ./cmd/server/ serve`, you can access:

- **Home**: http://localhost:8080/ - Beautiful Next.js/Django-style landing page
- **Login**: http://localhost:8080/login - User authentication
- **Register**: http://localhost:8080/register - User registration
- **Admin Dashboard**: http://localhost:8080/admin - Admin panel (requires auth)
- **Chat**: http://localhost:8080/chat - Real-time chat with SSE + HTMX 🆕
- **API Docs**: http://localhost:8080/docs - Interactive Swagger UI

All templates use the modern `go::` / `@` syntax and are styled with Tailwind CSS for a professional, responsive design.

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Test template engine
go test ./internal/framework/view/... -v

# Test HTTP routes
go test ./internal/app/router/... -v

# Run all tests
go test ./internal/... -v
```text

The framework includes:
- ✅ 10 template engine tests (echo, logic, escaping, etc.)
- ✅ 2 HTTP route tests (home page, 404 handling)
- ✅ Full coverage of new template syntax

## 🚀 Advanced Examples

goBastion includes production-ready examples demonstrating senior-level Go patterns and modern web technologies.

### Real-Time Chat Application

A complete real-time chat demonstrating advanced Go concurrency and HTMX integration:

**🎯 Features:**
- **Real-time messaging** using Server-Sent Events (SSE)
- **Concurrent message broker** with Go channels and goroutines
- **HTMX integration** for reactive UI without JavaScript frameworks
- **Thread-safe operations** using mutexes
- **Multiple chat rooms** with message history
- **Live statistics** with automatic updates

**🔗 Try it:**
```bash
go run ./cmd/server/ serve
# Visit: http://localhost:8080/chat
```text

**📊 Go Patterns Demonstrated:**

1. **Concurrent Message Broker**
```go
type Broker struct {
    clients    map[string]*Client
    mu         sync.RWMutex
    register   chan *Client
    broadcast  chan Message
    ctx        context.Context
}
```text

2. **Event Loop with Channels**
```go
for {
    select {
    case client := <-b.register:
        b.registerClient(client)
    case message := <-b.broadcast:
        b.broadcastMessage(message)
    case <-b.ctx.Done():
        return
    }
}
```text

3. **SSE Streaming**
```go
for {
    select {
    case msg := <-client.Messages:
        fmt.Fprintf(w, "data: %s\n\n", data)
        flusher.Flush()
    case <-ctx.Done():
        return
    }
}
```text

**🎨 HTMX Features:**
- Form submission without page reload
- Server-Sent Events integration
- Partial template rendering
- Auto-scrolling chat messages

**📁 Files:**
- `internal/app/chat/broker.go` - Concurrent message broker
- `internal/app/router/chat.go` - HTTP handlers with SSE
- `templates/chat/room.html` - HTMX-powered chat UI

### Worker Pool Pattern

Background job processing with worker pools:

```go
func ProcessJobsWithWorkerPool(jobs []Job, numWorkers int) []Result {
    jobChan := make(chan Job, len(jobs))
    resultChan := make(chan Result, len(jobs))

    // Start workers (fan-out)
    for i := 0; i < numWorkers; i++ {
        go worker(ctx, i, jobChan, resultChan)
    }

    // Send jobs
    for _, job := range jobs {
        jobChan <- job
    }
    close(jobChan)

    // Collect results (fan-in)
    for i := 0; i < len(jobs); i++ {
        results = append(results, <-resultChan)
    }

    return results
}
```text

### Rate Limiter (Token Bucket)

Custom rate limiter implementation:

```go
type RateLimiter struct {
    tokens    chan struct{}
    refillInterval time.Duration
}

func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```text

**🎯 Patterns Used:**
- ✅ Goroutines & Channels - Concurrent message passing
- ✅ sync.RWMutex - Thread-safe map operations
- ✅ Context - Cancellation & timeouts
- ✅ Select Statement - Multiplexing channels
- ✅ Worker Pools - Parallel job processing
- ✅ Token Bucket - Rate limiting
- ✅ SSE - Real-time server push
- ✅ HTMX - Modern reactive UI

**📚 Learn More:**
See the chat implementation in `internal/app/chat/` and `internal/app/router/chat.go` for complete examples.

## 📁 Project Structure

```text
goBastion/
├── cmd/
│   ├── server/
│   │   └── main.go              # Application entry point
│   └── gobastion-vim-installer/ # Vim/Neovim syntax installer
├── config/
│   └── config.json              # Configuration file
├── internal/
│   ├── app/                     # Your application code
│   │   ├── chat/                # 🆕 Real-time chat (SSE + Go channels)
│   │   │   └── broker.go        # Concurrent message broker
│   │   ├── models/              # Data models
│   │   └── router/              # Route handlers
│   │       ├── chat.go          # 🆕 Chat routes (SSE, HTMX, worker pools)
│   │       ├── auth.go          # Authentication routes
│   │       └── users.go         # User CRUD routes
│   └── framework/               # Framework core
│       ├── admin/               # Admin panel
│       ├── config/              # Config management
│       ├── db/                  # Database layer
│       ├── middleware/          # HTTP middleware
│       ├── router/              # HTTP router
│       ├── security/            # Auth & security
│       └── view/                # Template engine
├── templates/                   # HTML templates (new syntax!)
│   ├── home.html                # Landing page
│   ├── auth/                    # Login/register pages
│   ├── admin/                   # Admin templates
│   └── chat/                    # 🆕 Real-time chat UI
│       ├── room.html            # Chat interface (HTMX + SSE)
│       └── messages.html        # Message partial template
├── goBastionTemplates/          # Editor support
│   ├── gobastion-templates-0.1.0.vsix  # VS Code extension
│   └── gBTemplatesNvim.sh       # Vim/Neovim installer script
├── static/
│   └── css/
│       └── output.css           # Tailwind CSS
├── TEMPLATE_SYNTAX.md           # Template syntax guide
├── REFACTOR_SUMMARY.md          # Refactor details
└── index.html                   # Framework documentation
```

### 🎯 What to Modify (and What Not To)

Understanding which files are safe to modify is crucial for extending goBastion without breaking core functionality.

#### ✅ **SAFE TO MODIFY** - Application Layer

These are the files you're **expected** to customize for your application:

| Directory/File | Purpose | Modify Freely |
|---------------|---------|---------------|
| `templates/**/*.html` | Your HTML templates | ✅ Yes - Add your pages here |
| `internal/app/models/` | Your data models | ✅ Yes - Define your domain models |
| `internal/app/router/` | Your route handlers | ✅ Yes - Add your business logic |
| `internal/app/chat/` | Example app feature | ✅ Yes - Customize or remove |
| `static/` | Static assets (CSS, JS, images) | ✅ Yes - Add your assets |
| `config/config.json` | Configuration | ✅ Yes - Configure your app |
| `cmd/server/main.go` | Application entry point | ⚠️ **Carefully** - See below |

**Example modifications you should make:**
- Add new routes in `internal/app/router/`
- Create new templates in `templates/`
- Add your models in `internal/app/models/`
- Customize styling in `static/css/`
- Configure ports/database in `config/config.json`

#### ⛔ **DO NOT MODIFY** - Framework Core

These files contain battle-tested framework code. Modifying them can break security, performance, or core functionality:

| Directory | Purpose | Modify? |
|-----------|---------|---------|
| `internal/framework/view/` | Template engine | ❌ **No** - Core template preprocessing |
| `internal/framework/router/` | HTTP router | ❌ **No** - Request routing logic |
| `internal/framework/middleware/` | HTTP middlewares | ❌ **No** - CSRF, auth, rate limiting |
| `internal/framework/security/` | Security primitives | ❌ **No** - JWT, CSRF token generation |
| `internal/framework/db/` | Database layer | ❌ **No** - Query builder, connections |
| `internal/framework/config/` | Config management | ❌ **No** - Config loading & validation |
| `internal/framework/admin/` | Admin panel core | ⚠️ **Rarely** - See extending section |

**Why not modify these?**
- They contain security-critical code (CSRF, JWT, XSS prevention)
- They're shared infrastructure used throughout the framework
- Breaking changes here affect the entire application
- Updates and bug fixes are applied to these files

#### 🔧 **MODIFY CAREFULLY** - Integration Points

Some files bridge the framework and your application. Modify these with understanding:

**`cmd/server/main.go`** - Application bootstrap
- ✅ **DO**: Register your custom routes
- ✅ **DO**: Add your middleware to the stack
- ✅ **DO**: Initialize your services
- ❌ **DON'T**: Remove framework initialization
- ❌ **DON'T**: Change the server startup order
- ❌ **DON'T**: Skip security middleware registration

**Example safe modification in `main.go`:**
```go
// ✅ GOOD: Adding your custom routes
router.RegisterAppRoutes(r, db)  // Existing framework routes
registerMyCustomRoutes(r, db)     // Your new routes

// ❌ BAD: Removing framework routes
// router.RegisterAuthRoutes(r, cfg.Security, tmplEngine)  // Don't comment out!
```

### 🚀 Extending the Framework

If you need to extend framework core functionality, prefer these approaches:

1. **Configuration First**: Check if `config/config.json` supports your need
2. **Middleware Pattern**: Add new middleware instead of modifying existing ones
3. **Template Helpers**: Extend template functions via the view engine's function map
4. **Custom Handlers**: Write new handlers in `internal/app/router/` that use framework utilities

**Example: Adding a Custom Middleware**
```go
// ✅ GOOD: Create new middleware in your app code
// File: internal/app/middleware/custom.go

package middleware

import "net/http"

func MyCustomMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your logic here
        next.ServeHTTP(w, r)
    })
}

// Then in main.go:
r.Use(middleware.MyCustomMiddleware)
```

**Example: Adding Template Functions**
```go
// ✅ GOOD: Extend template functions without modifying view.go
// In your handler or init code:

funcMap := template.FuncMap{
    "myCustomFunc": func(s string) string {
        return strings.ToUpper(s)
    },
}

// Register with view engine
views.AddFunctions(funcMap)
```

### ⚠️ When You MUST Modify Framework Core

In rare cases, you may need to modify framework code (e.g., customizing the template engine syntax). If you must:

1. **Understand the implications** - Read all code comments in the file
2. **Maintain security** - Never bypass CSRF, XSS prevention, or authentication
3. **Document your changes** - Add comments explaining why you changed it
4. **Consider contributing back** - If it's a general improvement, submit a PR
5. **Test thoroughly** - Run all tests and verify security features still work

**Files that might need careful extension:**
- `internal/framework/view/view.go` - If you want custom template syntax
- `internal/framework/admin/admin.go` - If you need admin panel customization
- `internal/framework/middleware/` - If existing middlewares don't fit your needs

## 🚀 Quick Start

1. **Install the CLI**:
   ```bash
   go install github.com/AlejandroMBJS/goBastion/cmd/go-bastion@latest
   ```text

2. **Create a new project**:
   ```bash
   go-bastion my-new-app
   cd my-new-app
   ```text

3. **Start the server**:
   ```bash
   go run ./cmd/server/ serve
   ```text

4. **Visit your app**:
   - Open http://localhost:8080/ to see the beautiful home page
   - Check http://localhost:8080/docs for API documentation

## 📖 Documentation

- **[index.html](index.html)** - Complete framework documentation
- **[TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md)** - Template syntax reference
- **[REFACTOR_SUMMARY.md](REFACTOR_SUMMARY.md)** - Recent changes and improvements

## 💡 Features

- 🔒 **Security First**: JWT auth, CSRF protection, rate limiting
- 🎨 **Modern Templates**: Clean `go::` / `@` syntax
- 💅 **Tailwind CSS**: Beautiful, responsive UI out of the box
- 🗄️ **Database Ready**: SQLite/PostgreSQL/MySQL support
- 📚 **Auto API Docs**: OpenAPI 3.0 with Swagger UI
- 🛠️ **CLI Tools**: Migrations, seeding, module generation
- ⚡ **Live Reload**: Development server with automatic restart
- ✅ **Well Tested**: Comprehensive test suite included

## 🔒 Security Best Practices

goBastion is built with security as a top priority. This section explains the security features and how to use them correctly.

### Built-In Security Features

#### 1. **CSRF Protection** (Cross-Site Request Forgery)

CSRF protection is **enabled by default** and protects all form submissions.

**How it works:**
- Server generates a unique token per session
- Token is stored in an HTTP-only cookie
- Forms must include the token in a hidden field
- Server validates token matches before processing POST/PUT/DELETE requests

**Usage in templates:**
```html
<form method="POST" action="/admin/users">
  go:: if .CSRFToken
  <input type="hidden" name="csrf_token" value="@.CSRFToken">
  ::end

  <!-- Your form fields -->
  <button type="submit">Submit</button>
</form>
```

**Configuration:**
```json
{
  "security": {
    "enable_csrf": true,
    "csrf_cookie_name": "csrf_token"
  }
}
```

**⚠️ CRITICAL:** Never disable CSRF protection in production!

#### 2. **XSS Protection** (Cross-Site Scripting)

The template engine **automatically HTML-escapes** all `@expr` outputs.

**Safe by default:**
```html
<!-- User input is automatically escaped -->
<p>Hello @.UserInput</p>

<!-- If UserInput = "<script>alert('xss')</script>" -->
<!-- Output: Hello &lt;script&gt;alert('xss')&lt;/script&gt; -->
```

**When you need raw HTML (dangerous!):**
```go
// In handler - sanitize first!
import "html/template"

data := map[string]any{
    "SafeHTML": template.HTML(sanitizedContent),
}
```

**Best practices:**
- ✅ Always use `@expr` for user input
- ✅ Sanitize HTML before marking as `template.HTML`
- ❌ NEVER pass unsanitized user input as raw HTML
- ❌ NEVER bypass template escaping for user input

#### 3. **JWT Authentication**

JWT (JSON Web Token) authentication provides secure, stateless authentication.

**How it works:**
- User logs in with email/password
- Server generates JWT with claims (user ID, role, expiration)
- JWT stored in HTTP-only cookie (not accessible to JavaScript)
- Server validates JWT on protected routes

**Protecting routes:**
```go
import "github.com/AlejandroMBJS/goBastion/internal/framework/middleware"

// Require authentication
authRequired := middleware.RequireAuth()
r.Handle("GET", "/profile", authRequired(handleProfile))

// Require specific role
adminOnly := middleware.RequireRole("admin")
r.Handle("GET", "/admin", adminOnly(handleAdmin))
```

**Configuration:**
```json
{
  "security": {
    "enable_jwt": true,
    "jwt_secret": "CHANGE-THIS-IN-PRODUCTION",
    "access_token_minutes": 15
  }
}
```

**⚠️ CRITICAL Security Settings:**
- `jwt_secret`: **MUST** be changed in production
- Use a strong, random secret (minimum 32 characters)
- Never commit secrets to Git
- Use environment variables: `export APP_JWT_SECRET="your-secret"`

#### 4. **Rate Limiting**

Rate limiting prevents brute force attacks and API abuse.

**Configuration:**
```json
{
  "rate_limit": {
    "enabled": true,
    "requests_per_minute": 60
  }
}
```

**How it works:**
- Tracks requests per IP address
- Returns `429 Too Many Requests` when limit exceeded
- Resets every minute

**Customization:**
- Development: Set higher limits or disable
- Production: Set conservative limits (60 req/min is reasonable)
- APIs: Consider lower limits (30 req/min)

#### 5. **SQL Injection Prevention**

The database layer uses **prepared statements** by default.

**Safe (parameterized queries):**
```go
// ✅ GOOD: Using query builder (safe)
users, err := db.Query("SELECT * FROM users WHERE email = ?", email)

// ✅ GOOD: Using named parameters
result := db.QueryBuilder().
    Select("*").
    From("users").
    Where("email = ?", email).
    Execute()
```

**Dangerous (string concatenation):**
```go
// ❌ BAD: Never concatenate user input into SQL!
query := "SELECT * FROM users WHERE email = '" + email + "'"
// Vulnerable to: email = "' OR '1'='1"
```

**Best practices:**
- ✅ Always use parameterized queries
- ✅ Use the query builder for complex queries
- ❌ NEVER concatenate user input into SQL strings

#### 6. **Password Security**

User passwords are hashed with **bcrypt** (industry standard).

**Best practices:**
- ✅ Passwords hashed with bcrypt (cost factor 10)
- ✅ Never store plain-text passwords
- ✅ Never log passwords
- ✅ Enforce minimum password length (8+ characters)

**In your handlers:**
```go
import "golang.org/x/crypto/bcrypt"

// Hashing passwords
passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verifying passwords
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
if err != nil {
    // Password incorrect
}
```

### Production Security Checklist

Before deploying to production, verify these critical settings:

#### ⚠️ Must Change:
- [ ] `security.jwt_secret` - Set to strong random value (32+ chars)
- [ ] `app.environment` - Set to `"production"`
- [ ] `database.dsn` - Use production credentials, not defaults
- [ ] Default admin password - Change immediately after first login

#### ⚠️ Must Enable:
- [ ] HTTPS/TLS - Use reverse proxy (nginx, Caddy) with SSL certificate
- [ ] `security.enable_csrf` - Keep enabled (`true`)
- [ ] `security.enable_jwt` - Keep enabled (`true`)
- [ ] `rate_limit.enabled` - Keep enabled (`true`)

#### ⚠️ Must Secure:
- [ ] Environment variables - Use secrets management (not in code)
- [ ] Database credentials - Use environment variables or secrets
- [ ] File permissions - Restrict `config/config.json` access
- [ ] CORS origins - Whitelist only your actual domains in `server.allowed_origins`

#### ⚠️ Recommended:
- [ ] Set up database backups
- [ ] Enable request logging for security auditing
- [ ] Use a Web Application Firewall (WAF)
- [ ] Implement monitoring and alerting
- [ ] Regular security updates for dependencies

### Environment Variables for Secrets

**Never commit secrets to Git!** Use environment variables:

```bash
# Production environment variables
export APP_ENVIRONMENT="production"
export APP_JWT_SECRET="your-super-secret-key-min-32-chars"
export APP_DB_DSN="postgres://user:pass@prod-db:5432/myapp"
export APP_PORT=":8080"

# Start server (reads environment variables)
./gobastion-server
```

**In CI/CD pipelines:**
- Use secrets management (GitHub Secrets, AWS Secrets Manager, etc.)
- Never log secrets in build output
- Rotate secrets regularly

### Common Security Mistakes to Avoid

#### ❌ DON'T: Disable CSRF protection
```json
{
  "security": {
    "enable_csrf": false  // ❌ DON'T DO THIS IN PRODUCTION!
  }
}
```

#### ❌ DON'T: Use weak JWT secrets
```json
{
  "security": {
    "jwt_secret": "secret"  // ❌ TOO SHORT! Use 32+ characters
  }
}
```

#### ❌ DON'T: Bypass template escaping for user input
```go
// ❌ BAD: XSS vulnerability!
data := map[string]any{
    "UserComment": template.HTML(userInput),  // Dangerous!
}
```

#### ❌ DON'T: Commit secrets to Git
```bash
# ❌ BAD: Secret visible in Git history
git add config/config.json  # Contains JWT secret
git commit -m "Add config"
```

#### ❌ DON'T: Skip authentication middleware
```go
// ❌ BAD: Admin route without authentication!
r.Handle("GET", "/admin", handleAdmin)  // Anyone can access!

// ✅ GOOD: Require admin role
adminAuth := middleware.RequireRole("admin")
r.Handle("GET", "/admin", adminAuth(handleAdmin))
```

### Security Resources

- **OWASP Top 10**: https://owasp.org/Top10/
- **Go Security**: https://go.dev/doc/security/
- **bcrypt Guide**: https://github.com/golang/crypto/tree/master/bcrypt
- **JWT Best Practices**: https://tools.ietf.org/html/rfc8725

### Reporting Security Issues

If you discover a security vulnerability, please email the maintainers directly instead of opening a public issue.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the MIT License.

## goBastion Template Syntax

goBastion uses a custom template engine built on top of Go's `html/template` package. The engine provides a clean, intuitive syntax while maintaining all the security benefits of Go templates, including automatic HTML escaping.

## Overview

Templates are typically stored in files with the `.gb.html` extension (for example `home.gb.html`).

The goBastion template syntax uses **only two constructs**:

1. **Logic Blocks** (`go:: ... ::end`) - For control flow and logic
2. **Echo Expressions** (`@expr`) - For outputting values

This minimal, expressive syntax keeps templates clean and readable while providing all the power you need.

---

## Echo Expressions

Echo expressions output values with automatic HTML escaping.

### Syntax

```html
@expression
```

### Examples

**Simple variable:**
```html
<h1>@.Title</h1>
```

**Object property:**
```html
<p>Hello, @user.Name!</p>
<p>Email: @user.Email</p>
```

**Nested properties:**
```html
<img src="@user.Profile.Avatar" alt="Avatar">
```

**Function calls:**
```html
<p>Price: @formatPrice(product.Price)</p>
<p>@upper(.Title)</p>
```

### HTML Escaping

All echo expressions are **automatically HTML-escaped** for security:

```html
@userInput
<!-- If userInput contains: <script>alert('xss')</script> -->
<!-- Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; -->
```

This prevents XSS (Cross-Site Scripting) attacks by default.

---

## Logic Blocks

Logic blocks control template flow using real Go code.

### Syntax

```html
go:: <go-statement>
  <!-- template content -->
::end
```

### If Statements

**Basic if:**
```html
go:: if .User
  <p>Welcome, @.User.Name!</p>
::end
```

**If/else:**
```html
go:: if .Error
  <div class="error">@.Error</div>
go:: else
  <div class="success">Operation successful!</div>
::end
```

**If with comparison:**
```html
go:: if eq .Role "admin"
  <a href="/admin">Admin Panel</a>
::end
```

### Range Loops

**Iterate over a slice:**
```html
<ul>
go:: range .Items
  <li>@.Name - $@.Price</li>
::end
</ul>
```

**Range with index and value:**
```html
<ol>
go:: range $index, $item := .Items
  <li>Item #@$index: @$item.Name</li>
::end
</ol>
```

**Empty list handling:**
```html
go:: if .Users
<table>
go:: range .Users
  <tr>
    <td>@.ID</td>
    <td>@.Name</td>
    <td>@.Email</td>
  </tr>
::end
</table>
go:: else
<p>No users found.</p>
::end
```

### With Blocks

Set the context to a specific value:

```html
go:: with .User
  <h2>@.Name</h2>
  <p>@.Email</p>
  <p>Role: @.Role</p>
::end
```

### Nested Blocks

You can nest blocks as deeply as needed:

```html
go:: if .Posts
<div class="posts">
  go:: range .Posts
  <article>
    <h2>@.Title</h2>
    <p>@.Content</p>
    go:: if .Comments
    <div class="comments">
      <h3>Comments</h3>
      go:: range .Comments
      <div class="comment">
        <strong>@.Author:</strong> @.Text
      </div>
      ::end
    </div>
    ::end
  </article>
  ::end
</div>
::end
```

---

## Complete Example

Here's a full template showing various features:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>@.Title - goBastion</title>
</head>
<body>
    <header>
        <h1>@.Title</h1>
        go:: if .User
        <p>Welcome, @.User.Name! | <a href="/logout">Logout</a></p>
        go:: else
        <p><a href="/login">Login</a> | <a href="/register">Register</a></p>
        ::end
    </header>

    <main>
        go:: if .Error
        <div class="alert alert-danger">
            @.Error
        </div>
        ::end

        go:: if .Success
        <div class="alert alert-success">
            @.Success
        </div>
        ::end

        go:: if .Users
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                go:: range .Users
                <tr>
                    <td>@.ID</td>
                    <td>@.Name</td>
                    <td>@.Email</td>
                    <td>
                        go:: if eq .Role "admin"
                        <span class="badge badge-danger">Admin</span>
                        go:: else
                        <span class="badge badge-success">User</span>
                        ::end
                    </td>
                    <td>
                        go:: if .IsActive
                        <span class="text-success">Active</span>
                        go:: else
                        <span class="text-muted">Inactive</span>
                        ::end
                    </td>
                </tr>
                ::end
            </tbody>
        </table>
        go:: else
        <p>No users found.</p>
        ::end
    </main>

    <footer>
        <p>&copy; 2025 goBastion</p>
    </footer>
</body>
</html>
```

---

## Template Functions

goBastion provides built-in template functions:

| Function | Description | Example |
|----------|-------------|---------|
| `upper` | Convert to uppercase | `{{ upper .Text }}` |
| `lower` | Convert to lowercase | `{{ lower .Text }}` |
| `title` | Title case | `{{ title .Text }}` |
| `eq` | Equal comparison | `{{ eq .Role "admin" }}` |

You can use these in logic blocks or with the standard `{{ }}` syntax.

---

## Before/After Comparison

### Old PHP-style Syntax (Deprecated)

```html
<?php if ($user != nil) { ?>
  <p>Hello <?= $user->name ?></p>
  <ul>
  <?php foreach ($items as $item) { ?>
    <li><?= $item ?></li>
  <?php } ?>
  </ul>
<?php } ?>
```

### New goBastion Syntax

```html
go:: if .User
  <p>Hello @.User.Name</p>
  <ul>
  go:: range .Items
    <li>@.</li>
  ::end
  </ul>
::end
```

**Benefits of the new syntax:**
- ✅ Cleaner, more Go-like
- ✅ No PHP confusion
- ✅ Easier to read and write
- ✅ Better editor support
- ✅ Same security guarantees

---

## Backward Compatibility

The old PHP-style tags (`<?`, `<?=`, `?>`) are still supported for backward compatibility but are **deprecated**. You should migrate to the new syntax.

Old templates will continue to work, but we recommend updating them to use `go::` and `@` syntax.

---

## Best Practices

### 1. **Keep templates simple**
Complex logic belongs in handlers, not templates.

**❌ Bad:**
```html
go:: if and (gt .User.Age 18) (eq .User.Country "US") (not .User.Banned)
  <button>Access Content</button>
::end
```

**✅ Good (in handler):**
```go
data := map[string]any{
    "CanAccess": user.Age > 18 && user.Country == "US" && !user.Banned,
}
```
```html
go:: if .CanAccess
  <button>Access Content</button>
::end
```

### 2. **Use meaningful variable names**
```html
go:: range $user := .Users
  <li>@$user.Name</li>
::end
```

### 3. **Check for empty collections**
```html
go:: if .Items
  <!-- show items -->
go:: else
  <p>No items available</p>
::end
```

### 4. **Leverage HTML escaping**
Never use raw HTML unless absolutely necessary. The `@` syntax escapes by default.

---

## Security

### Automatic Escaping

All `@expr` outputs are **automatically HTML-escaped**:

```html
@userComment
<!-- Input: <script>alert('xss')</script> -->
<!-- Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; -->
```

### CSRF Protection

goBastion includes built-in CSRF protection. Use CSRF tokens in forms:

```html
<form method="POST" action="/submit">
    go:: if .CSRFToken
    <input type="hidden" name="csrf_token" value="@.CSRFToken">
    ::end

    <!-- form fields -->
    <button type="submit">Submit</button>
</form>
```

---

## Troubleshooting

### Common Errors

**1. Missing `::end`**
```
Error: unexpected EOF, expected {{ end }}
```
Make sure every `go::` has a matching `::end`.

**2. Syntax error in Go statement**
```
Error: unexpected "}", expected expression
```
Check your Go syntax inside `go::` blocks.

**3. Undefined variable**
```
Error: can't evaluate field User in type map[string]interface{}
```
Make sure you're passing the correct data to the template.

---

## Summary

- **`@expr`** - Echo expressions (HTML-escaped)
- **`go:: ... ::end`** - Logic blocks (if, range, with)
- Built on Go's `html/template` for security
- Clean, readable syntax
- Backward compatible with old PHP-style tags

Happy templating! 🎨

---

## 📚 Documentation & Code Comments Summary

This README and codebase have been enhanced with comprehensive documentation to help developers understand what they can safely modify and what they should avoid touching. Here's what was added:

### README Enhancements

#### 1. **Philosophy & Architecture** Section
- Explains goBastion's design principles (security by default, convention over configuration, DX-first)
- Shows clear separation between framework core (`internal/framework/`) and application layer (`internal/app/`)
- Illustrates how the template engine preprocesses syntax
- **Location**: After "What's New" section

#### 2. **Configuration** Section
- Documents the centralized `config/config.json` structure
- Explains all configuration sections (app, server, database, security, frontend, admin, features)
- Shows how to use config in handlers and templates
- Provides security warnings for production settings
- **Location**: After "Template Engine" section

#### 3. **Project Structure & What to Modify** Section
- **✅ SAFE TO MODIFY**: Lists application layer files developers should customize
- **⛔ DO NOT MODIFY**: Lists framework core files that should rarely be touched
- **🔧 MODIFY CAREFULLY**: Explains integration points like `main.go`
- Provides concrete examples of safe vs. unsafe modifications
- **Location**: Expanded existing "Project Structure" section

#### 4. **Extending the Framework** Subsection
- Shows how to add custom middleware
- Shows how to add custom template functions
- Explains when and how to modify framework core (rare cases)
- **Location**: Within "Project Structure" section

### Code Comment Enhancements

#### 1. **view.go** - Template Engine Core
```go
// ⚠️ FRAMEWORK CORE - DO NOT MODIFY
//
// WHEN TO MODIFY:
//   - ❌ NEVER modify for application changes
//   - ❌ DO NOT bypass preprocessing (security risk)
//   - ✅ OK to extend via AddFunctions()
```
- **Added**: Comprehensive package header explaining security implications
- **Added**: Detailed function comments for `NewEngine()` and `Render()`
- **Added**: Usage examples and extension points
- **Added**: Template syntax rules and security notes

#### 2. **config.go** - Configuration Management
```go
// ⚠️ FRAMEWORK CORE - MODIFY CAREFULLY
//
// WHEN TO MODIFY:
//   - ✅ ADD new configuration fields
//   - ❌ DO NOT remove existing fields (breaking)
//   - ❌ DO NOT change JSON field names
```
- **Added**: Package header explaining configuration flow
- **Added**: Guidelines on extending configuration safely
- **Added**: Step-by-step instructions for adding new config sections
- **Added**: Examples of proper config usage

#### 3. **main.go** - Application Entry Point
```go
// ⚠️ APPLICATION LAYER - MODIFY CAREFULLY
//
// INITIALIZATION ORDER (CRITICAL):
//  1. Load configuration
//  2. Initialize database
//  3. Initialize template engine
//  ...
```
- **Added**: Comprehensive bootstrap documentation
- **Added**: Critical initialization order explanation
- **Added**: Safe customization examples (routes, middleware, services)
- **Added**: Clear DO/DON'T guidelines for modifications

### Key Documentation Principles Applied

1. **Clear Visual Markers**
   - ⚠️ Warnings for critical sections
   - ✅ Green checkmarks for safe operations
   - ❌ Red X for forbidden operations
   - 🔧 Wrench for "modify carefully" areas

2. **Explicit "Touch/Don't Touch" Guidance**
   - Every major file/package clearly labeled as framework core or app layer
   - Specific examples of what to modify and what not to modify
   - Explanations of why certain files shouldn't be modified

3. **Security-First Documentation**
   - Security warnings in configuration section
   - XSS/CSRF protection explained in template engine
   - JWT secret and production environment warnings

4. **Practical Examples**
   - Code examples show the RIGHT way to extend functionality
   - BAD examples marked with ❌ to show what NOT to do
   - Real-world use cases (adding routes, middleware, config)

### Where to Find Documentation

- **README.md**: Complete framework overview, quickstart, configuration, project structure
- **TEMPLATE_SYNTAX.md**: Detailed template syntax reference
- **Code Comments**: In-file documentation for all framework core files
- **index.html**: Framework documentation website (see root directory)
- **config/config.json**: Configuration file with inline comments

### Documentation Best Practices Used

✅ **Preservation**: All existing content preserved and enhanced, not replaced
✅ **Clarity**: Clear, professional language similar to Django/FastAPI/Next.js docs
✅ **Actionable**: Specific examples and instructions, not just theory
✅ **Security-Conscious**: Warnings about security-critical code
✅ **Developer Experience**: Helps developers succeed without breaking things

### Next Steps for Developers

1. **Read** the "Philosophy & Architecture" section to understand the framework
2. **Review** the "Project Structure & What to Modify" section to know where to work
3. **Check** the "Configuration" section to customize your application
4. **Explore** code comments in `view.go`, `config.go`, and `main.go` for deep understanding
5. **Follow** the "Extending the Framework" examples when adding new features

**Remember**: When in doubt, check the comments in the file you're modifying. Every critical file now has clear guidance on when and how to modify it safely.

---

## 🤝 Contributing

Contributions are welcome! When contributing:

1. **Preserve existing functionality** - Don't delete features
2. **Follow the documentation patterns** - Use ✅/❌/⚠️ markers consistently
3. **Add comments for new code** - Especially "touch/don't touch" guidance
4. **Test thoroughly** - Run tests and verify security features still work
5. **Update README** - Document new features in appropriate sections

Please feel free to submit a Pull Request with improvements to code or documentation!

