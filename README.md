# goBastion

goBastion is a Go framework designed for rapid development of web applications and APIs. This repository serves as a template for new projects, providing a solid foundation with a pre-configured structure, authentication, and database integration.

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
