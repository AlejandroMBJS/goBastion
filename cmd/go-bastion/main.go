package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		handleServe()
	case "migrate":
		handleMigrate()
	case "seed":
		handleSeed()
	case "create-admin":
		handleCreateAdmin()
	case "test":
		handleTest()
	case "doctor":
		handleDoctor()
	case "new-module":
		handleNewModule()
	case "version", "-v", "--version":
		fmt.Printf("go-bastion v%s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	help := `
╔══════════════════════════════════════════════════════════════╗
║                       GO-BASTION CLI                         ║
║          Production-Ready Go Web Framework Manager           ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
    go-bastion <command> [options]

COMMANDS:
    serve            Start the HTTP server
    migrate          Run database migrations
    seed             Insert default admin user
    create-admin     Create a specific admin user
    test             Run all tests (go test ./...)
    doctor           Run health checks on the system
    new-module       Scaffold a new API module
    version          Show version information
    help             Show this help message

EXAMPLES:
    go-bastion serve --port :8080
    go-bastion migrate
    go-bastion seed
    go-bastion create-admin --email admin@example.com --password Secret123 --name "Admin User"
    go-bastion test
    go-bastion doctor
    go-bastion new-module posts

For more information on a specific command:
    go-bastion <command> --help

`
	fmt.Print(help)
}

func handleServe() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", "", "Host to bind to (overrides config)")
	port := fs.String("port", "", "Port to bind to (overrides config, e.g., :8080)")
	env := fs.String("env", "", "Environment (dev, prod)")

	fs.Parse(os.Args[2:])

	// Override environment variables if flags are provided
	if *port != "" {
		os.Setenv("APP_PORT", *port)
	}
	if *host != "" {
		os.Setenv("APP_HOST", *host)
	}

	// Display configuration using Bubble Tea
	displayConfigAndServe(*env)
}

func handleMigrate() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	runMigration()
}

func handleSeed() {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	runSeed()
}

func handleCreateAdmin() {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	email := fs.String("email", "", "Admin email address")
	password := fs.String("password", "", "Admin password")
	name := fs.String("name", "Admin", "Admin name")

	fs.Parse(os.Args[2:])

	if *email == "" || *password == "" {
		fmt.Println("Error: --email and --password are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	runCreateAdmin(*email, *password, *name)
}

func handleTest() {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	verbose := fs.Bool("v", false, "Verbose output")
	fs.Parse(os.Args[2:])

	runTests(*verbose)
}

func handleDoctor() {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	runDoctor()
}

func handleNewModule() {
	if len(os.Args) < 3 {
		fmt.Println("Error: module name is required")
		fmt.Println("Usage: go-bastion new-module <module-name>")
		os.Exit(1)
	}

	moduleName := os.Args[2]
	if !isValidModuleName(moduleName) {
		fmt.Println("Error: invalid module name. Use lowercase letters, numbers, and hyphens only")
		os.Exit(1)
	}

	runNewModule(moduleName)
}

func isValidModuleName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

func pluralize(name string) string {
	if strings.HasSuffix(name, "s") {
		return name + "es"
	}
	if strings.HasSuffix(name, "y") {
		return strings.TrimSuffix(name, "y") + "ies"
	}
	return name + "s"
}
