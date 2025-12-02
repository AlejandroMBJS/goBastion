package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/bcrypt"

	"github.com/AlejandroMBJS/goBastion/internal/app/models"
	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
)

type createAdminModel struct {
	email    string
	name     string
	status   string
	done     bool
	err      error
	quitting bool
}

func (m createAdminModel) Init() tea.Cmd {
	return runCreateAdminCmd(m.email, m.name)
}

func (m createAdminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case createAdminResultMsg:
		m.done = true
		m.err = msg.err
		if msg.err == nil {
			m.status = "created successfully"
		} else {
			m.status = fmt.Sprintf("failed: %v", msg.err)
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m createAdminModel) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		Width(50).
		Align(lipgloss.Center)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")).
		Bold(true)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	var s string
	s += titleStyle.Render("CREATE ADMIN USER") + "\n\n"

	if !m.done {
		s += statusStyle.Render("Creating admin user...") + "\n"
		s += "\n  • Hashing password\n"
		s += "  • Inserting into database\n"
	} else {
		if m.err == nil {
			s += successStyle.Render("✓ Admin user "+m.status) + "\n\n"
			s += "  " + infoStyle.Render("Admin user details:") + "\n"
			s += fmt.Sprintf("    Name:  %s\n", m.name)
			s += fmt.Sprintf("    Email: %s\n", m.email)
			s += "    Role:  admin\n"
		} else {
			s += errorStyle.Render("✗ Admin user creation "+m.status) + "\n"
		}
	}

	return s
}

type createAdminResultMsg struct {
	err error
}

func runCreateAdminCmd(email, name string) func() tea.Msg {
	return func() tea.Msg {
		cfg, err := config.Load("config/config.json")
		if err != nil {
			return createAdminResultMsg{err: err}
		}

		if err := db.Init(cfg.Database); err != nil {
			return createAdminResultMsg{err: err}
		}

		// The password is passed from the flag
		return createAdminResultMsg{err: nil}
	}
}

func runCreateAdmin(email, password, name string) {
	// Validate inputs
	if len(password) < 8 {
		log.Fatal("Password must be at least 8 characters long")
	}

	fmt.Println("Creating admin user...")

	// Load config and initialize database
	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create the user
	input := models.RegisterInput{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     "admin",
	}

	_, err = db.CreateUser(context.Background(), input, string(hashedPassword))
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	// Show success message
	fmt.Println("✓ Admin user created successfully")
	fmt.Println("  Admin user details:")
	fmt.Printf("    Name:  %s\n", name)
	fmt.Printf("    Email: %s\n", email)
	fmt.Println("    Role:  admin")
}
