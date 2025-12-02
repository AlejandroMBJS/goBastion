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

type seedModel struct {
	status   string
	done     bool
	err      error
	quitting bool
}

func (m seedModel) Init() tea.Cmd {
	return runSeedCmd()
}

func (m seedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case seedResultMsg:
		m.done = true
		m.err = msg.err
		if msg.err == nil {
			m.status = "completed successfully"
		} else {
			m.status = fmt.Sprintf("failed: %v", msg.err)
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m seedModel) View() string {
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
	s += titleStyle.Render("DATABASE SEEDING") + "\n\n"

	if !m.done {
		s += statusStyle.Render("Seeding database...") + "\n"
		s += "\n  • Creating default admin user\n"
	} else {
		if m.err == nil {
			s += successStyle.Render("✓ Seeding "+m.status) + "\n\n"
			s += "  " + infoStyle.Render("Default admin user created:") + "\n"
			s += "    Email:    admin@example.com\n"
			s += "    Password: AdminPass123\n"
			s += "    Role:     admin\n"
		} else {
			s += errorStyle.Render("✗ Seeding "+m.status) + "\n"
		}
	}

	return s
}

type seedResultMsg struct {
	err error
}

func runSeedCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load("config/config.json")
		if err != nil {
			return seedResultMsg{err: err}
		}

		if err := db.Init(cfg.Database); err != nil {
			return seedResultMsg{err: err}
		}

		// Create default admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("AdminPass123"), bcrypt.DefaultCost)
		if err != nil {
			return seedResultMsg{err: err}
		}

		input := models.RegisterInput{
			Name:     "Admin",
			Email:    "admin@example.com",
			Password: "AdminPass123",
			Role:     "admin",
		}

		_, err = db.CreateUser(context.Background(), input, string(hashedPassword))
		if err != nil {
			// Check if user already exists
			return seedResultMsg{err: fmt.Errorf("admin user may already exist or error: %v", err)}
		}

		return seedResultMsg{err: nil}
	}
}

func runSeed() {
	fmt.Println("Seeding database...")

	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create default admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("AdminPass123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	input := models.RegisterInput{
		Name:     "Admin",
		Email:    "admin@example.com",
		Password: "AdminPass123",
		Role:     "admin",
	}

	_, err = db.CreateUser(context.Background(), input, string(hashedPassword))
	if err != nil {
		log.Fatalf("Seeding failed (admin user may already exist): %v", err)
	}

	fmt.Println("✓ Seeding completed successfully")
	fmt.Println("  Default admin user created:")
	fmt.Println("    Email:    admin@example.com")
	fmt.Println("    Password: AdminPass123")
	fmt.Println("    Role:     admin")
}
