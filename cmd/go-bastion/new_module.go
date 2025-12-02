package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type newModuleModel struct {
	moduleName string
	files      []string
	done       bool
	err        error
	quitting   bool
}

func (m newModuleModel) Init() tea.Cmd {
	return createModuleCmd(m.moduleName)
}

func (m newModuleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case moduleResultMsg:
		m.done = true
		m.err = msg.err
		m.files = msg.files
		return m, tea.Quit
	}
	return m, nil
}

func (m newModuleModel) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		Width(60).
		Align(lipgloss.Center)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	var s string
	s += titleStyle.Render("CREATE NEW MODULE") + "\n\n"

	if !m.done {
		s += fmt.Sprintf("Creating module: %s\n", m.moduleName)
		s += "\n  • Generating model file\n"
		s += "  • Generating router file\n"
		s += "  • Setting up structure\n"
	} else {
		if m.err == nil {
			s += successStyle.Render(fmt.Sprintf("✓ Module '%s' created successfully", m.moduleName)) + "\n\n"
			s += "Files created:\n"
			for _, file := range m.files {
				s += "  " + fileStyle.Render(file) + "\n"
			}
			s += "\n"
			s += "Next steps:\n"
			s += "  1. Implement your business logic in the model file\n"
			s += "  2. Add custom routes and handlers in the router file\n"
			s += "  3. Register routes in cmd/server/main.go\n"
		} else {
			s += errorStyle.Render("✗ Failed to create module") + "\n"
			s += fmt.Sprintf("Error: %v\n", m.err)
		}
	}

	return s
}

type moduleResultMsg struct {
	files []string
	err   error
}

func createModuleCmd(moduleName string) tea.Cmd {
	return func() tea.Msg {
		files := []string{}

		// Create directory
		moduleDir := filepath.Join("internal", "app", moduleName)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			return moduleResultMsg{err: err}
		}

		// Generate model file
		modelFile := filepath.Join(moduleDir, "models.go")
		modelContent := generateModelFile(moduleName)
		if err := os.WriteFile(modelFile, []byte(modelContent), 0644); err != nil {
			return moduleResultMsg{err: err}
		}
		files = append(files, modelFile)

		// Generate router file
		routerFile := filepath.Join(moduleDir, "router.go")
		routerContent := generateRouterFile(moduleName)
		if err := os.WriteFile(routerFile, []byte(routerContent), 0644); err != nil {
			return moduleResultMsg{err: err}
		}
		files = append(files, routerFile)

		return moduleResultMsg{files: files, err: nil}
	}
}

func generateModelFile(moduleName string) string {
	singular := strings.Title(moduleName)
	if strings.HasSuffix(singular, "s") {
		singular = strings.TrimSuffix(singular, "s")
	}

	return fmt.Sprintf(`package %s

import (
	"errors"
	"strings"
	"time"
)

// %s represents a %s entity
type %s struct {
	ID        int       `+"`json:\"id\"`"+`
	Name      string    `+"`json:\"name\"`"+`
	CreatedAt time.Time `+"`json:\"created_at\"`"+`
	UpdatedAt time.Time `+"`json:\"updated_at\"`"+`
}

// %sInput represents the input for creating/updating a %s
type %sInput struct {
	Name string `+"`json:\"name\"`"+`
}

// Validate validates the %s input
func (i %sInput) Validate() error {
	if len(strings.TrimSpace(i.Name)) == 0 {
		return errors.New("name is required")
	}
	if len(i.Name) < 2 {
		return errors.New("name must be at least 2 characters long")
	}
	if len(i.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}
	return nil
}
`, moduleName, singular, strings.ToLower(singular), singular,
		singular, strings.ToLower(singular), singular,
		strings.ToLower(singular), singular)
}

func generateRouterFile(moduleName string) string {
	singular := strings.Title(moduleName)
	if strings.HasSuffix(singular, "s") {
		singular = strings.TrimSuffix(singular, "s")
	}
	plural := pluralize(moduleName)

	return fmt.Sprintf(`package %s

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go-native-fastapi/internal/framework/config"
	"go-native-fastapi/internal/framework/db"
	"go-native-fastapi/internal/framework/middleware"
	frameworkrouter "go-native-fastapi/internal/framework/router"
)

// Register%sRoutes registers all %s routes
func Register%sRoutes(r *frameworkrouter.Router, securityCfg config.SecurityConfig) {
	// Apply JWT middleware to all routes
	jwtMiddleware := middleware.JWTMiddleware(securityCfg.JWTSecret)

	// List %s - GET /api/v1/%s
	r.Handle("GET", "/api/v1/%s",
		jwtMiddleware(list%s))

	// Create %s - POST /api/v1/%s
	r.Handle("POST", "/api/v1/%s",
		jwtMiddleware(create%s))

	// Get %s by ID - GET /api/v1/%s/{id}
	r.Handle("GET", "/api/v1/%s/{id}",
		jwtMiddleware(get%s))

	// Update %s - PUT /api/v1/%s/{id}
	r.Handle("PUT", "/api/v1/%s/{id}",
		jwtMiddleware(update%s))

	// Delete %s - DELETE /api/v1/%s/{id}
	r.Handle("DELETE", "/api/v1/%s/{id}",
		jwtMiddleware(delete%s))
}

// list%s returns all %s
func list%s(w http.ResponseWriter, r *http.Request, params map[string]string) {
	// TODO: Implement database query
	// Example:
	// rows, err := db.GetDB().Query("SELECT id, name, created_at, updated_at FROM %s")
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     return
	// }
	// defer rows.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "List %s - TODO: implement database query",
		"%s":      []%s{},
	})
}

// create%s creates a new %s
func create%s(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var input %sInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := input.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// TODO: Implement database insert
	// Example:
	// result, err := db.GetDB().Exec(
	//     "INSERT INTO %s (name, created_at, updated_at) VALUES (?, ?, ?)",
	//     input.Name, time.Now(), time.Now(),
	// )
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     return
	// }
	// id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Create %s - TODO: implement database insert",
		"input":   input,
	})
}

// get%s retrieves a single %s by ID
func get%s(w http.ResponseWriter, r *http.Request, params map[string]string) {
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement database query
	// Example:
	// var item %s
	// err := db.GetDB().QueryRow(
	//     "SELECT id, name, created_at, updated_at FROM %s WHERE id = ?",
	//     id,
	// ).Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
	// if err == sql.ErrNoRows {
	//     http.Error(w, "%s not found", http.StatusNotFound)
	//     return
	// }
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Get %s %%d - TODO: implement database query", id),
	})
}

// update%s updates a %s by ID
func update%s(w http.ResponseWriter, r *http.Request, params map[string]string) {
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input %sInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := input.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// TODO: Implement database update
	// Example:
	// result, err := db.GetDB().Exec(
	//     "UPDATE %s SET name = ?, updated_at = ? WHERE id = ?",
	//     input.Name, time.Now(), id,
	// )
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     return
	// }
	// rowsAffected, _ := result.RowsAffected()
	// if rowsAffected == 0 {
	//     http.Error(w, "%s not found", http.StatusNotFound)
	//     return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Update %s %%d - TODO: implement database update", id),
		"input":   input,
	})
}

// delete%s deletes a %s by ID
func delete%s(w http.ResponseWriter, r *http.Request, params map[string]string) {
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement database delete
	// Example:
	// result, err := db.GetDB().Exec("DELETE FROM %s WHERE id = ?", id)
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     return
	// }
	// rowsAffected, _ := result.RowsAffected()
	// if rowsAffected == 0 {
	//     http.Error(w, "%s not found", http.StatusNotFound)
	//     return
	// }

	w.WriteHeader(http.StatusNoContent)
}
`, moduleName,
		plural, singular,
		plural, plural, plural, singular,
		singular, plural, plural, singular,
		singular, plural, plural, singular,
		singular, plural, plural, singular,
		singular, plural, plural, singular,
		singular, plural, singular,
		plural, plural, plural, singular,
		singular, plural, plural, singular,
		singular, singular, singular,
		singular,
		plural,
		singular,
		singular, singular, singular,
		singular,
		plural,
		singular,
		singular, singular, singular,
		plural, singular, singular,
		singular, singular, singular,
		singular,
		plural,
		singular,
		singular, singular, singular,
		plural, plural)
}

func runNewModule(moduleName string) {
	p := tea.NewProgram(newModuleModel{moduleName: moduleName})
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error creating module: %v", err)
	}

	m := finalModel.(newModuleModel)
	if m.err != nil {
		log.Fatalf("Module creation failed: %v", m.err)
	}
}
