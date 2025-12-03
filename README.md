# goBastion

goBastion is a modern, production-ready Go framework designed for rapid development of web applications and APIs. This repository serves as a template for new projects, providing a solid foundation with a pre-configured structure, authentication, database integration, and a powerful custom template engine.

## ✨ What's New

- **🎨 Modern Template Syntax**: Clean, Go-like template syntax with `go::` / `@` constructs
- **💅 Tailwind-Styled Templates**: Beautiful, responsive UI out of the box
- **🏠 Professional Home Page**: Next.js/Django-style boilerplate landing page at `/`
- **🔒 Security First**: Auto HTML escaping, CSRF protection, JWT auth, rate limiting
- **✅ Comprehensive Tests**: Full test coverage for template engine and routes

## Project Generator CLI

A command-line interface (CLI) tool, `go-bastion`, has been developed to streamline the creation of new projects based on this template. This CLI automates the process of cloning the repository, configuring module names, and setting up the initial project structure.

### How to Build and Install the CLI

To build and install the `go-bastion` CLI, making it available globally in your system's PATH, execute the following command from the root of this repository:

```bash
go install ./cmd/go-bastion
```

This will compile the CLI and place the `go-bastion` executable in your Go binary path (e.g., `$GOPATH/bin` or `$HOME/go/bin`), allowing you to run it from any directory.

### How to Use the CLI

Once installed, you can use the `go-bastion` CLI in two modes:

#### 1. Non-Interactive Mode

To create a new project with a specified name directly, use the following command:

```bash
go-bastion my-new-app
```

Replace `my-new-app` with your desired project name. The CLI will clone the template into a new directory named `my-new-app`, configure the Go module, and prepare the project for development.

#### 2. Interactive Mode

If you prefer to be prompted for the project name, simply run the CLI without any arguments:

```bash
go-bastion
```

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
```

### 2. Logic Blocks (`go:: ... ::end`)

Control flow using real Go code:

```html
go:: if user != nil {
  <p>Welcome, @user.Name!</p>
::end

go:: range .Items
  <li>@.Name - $@.Price</li>
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
```

### Key Features

- ✅ **Auto HTML Escaping**: All `@expr` outputs are automatically escaped
- ✅ **Clean Syntax**: Only two constructs to learn (`go::` and `@`)
- ✅ **Type Safe**: Compiles to Go's `html/template`
- ✅ **Backward Compatible**: Old PHP-style tags still work (deprecated)
- ✅ **Security First**: Built-in CSRF protection and XSS prevention

### Full Documentation

For complete details on the template syntax, including advanced features, best practices, and security guidelines, see:

📚 **[TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md)** - Complete template syntax reference

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
```

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
```

📖 **[Full Vim/Neovim Guide](cmd/gobastion-vim-installer/README.md)**

### VS Code (using `.vsix` extension)

Install the pre-packaged VS Code extension:

#### Option 1: Install from Pre-built VSIX

```bash
# The extension is already packaged
code --install-extension goBastionTemplates/gobastion-templates-0.1.0.vsix
```

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
```

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
```

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
```

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
```

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
```

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
```

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
```

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
```

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

```
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

## 🚀 Quick Start

1. **Install the CLI**:
   ```bash
   go install github.com/AlejandroMBJS/goBastion/cmd/go-bastion@latest
   ```

2. **Create a new project**:
   ```bash
   go-bastion my-new-app
   cd my-new-app
   ```

3. **Start the server**:
   ```bash
   go run ./cmd/server/ serve
   ```

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

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the MIT License.
