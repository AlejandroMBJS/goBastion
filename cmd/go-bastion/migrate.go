package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
)

type migrateModel struct {
	status   string
	done     bool
	err      error
	quitting bool
}

func (m migrateModel) Init() tea.Cmd {
	return runMigrationCmd()
}

func (m migrateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case migrationResultMsg:
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

func (m migrateModel) View() string {
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

	var s string
	s += titleStyle.Render("DATABASE MIGRATION") + "\n\n"

	if !m.done {
		s += statusStyle.Render("Running migrations...") + "\n"
		s += "\n  • Creating tables\n"
		s += "  • Setting up indexes\n"
		s += "  • Initializing data\n"
	} else {
		if m.err == nil {
			s += successStyle.Render("✓ Migration "+m.status) + "\n\n"
			s += "  • Users table created\n"
			s += "  • Indexes created\n"
			s += "  • Database ready\n"
		} else {
			s += errorStyle.Render("✗ Migration "+m.status) + "\n"
		}
	}

	return s
}

type migrationResultMsg struct {
	err error
}

func runMigrationCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load("config/config.json")
		if err != nil {
			return migrationResultMsg{err: err}
		}

		// Init automatically creates tables
		if err := db.Init(cfg.Database); err != nil {
			return migrationResultMsg{err: err}
		}

		return migrationResultMsg{err: nil}
	}
}

func runMigration() {
	fmt.Println("Running database migration...")

	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✓ Migration completed successfully")
	fmt.Println("  • Users table created")
	fmt.Println("  • Indexes created")
	fmt.Println("  • Database ready")
}
