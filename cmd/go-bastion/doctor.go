package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

type checkResult struct {
	name   string
	status string
	err    error
}

type doctorModel struct {
	checks   []checkResult
	current  int
	done     bool
	quitting bool
}

func (m doctorModel) Init() tea.Cmd {
	return runHealthChecks()
}

func (m doctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case doctorResultMsg:
		m.checks = msg.checks
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m doctorModel) View() string {
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

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")).
		Width(30)

	var s string
	s += titleStyle.Render("SYSTEM HEALTH CHECK") + "\n\n"

	if !m.done {
		s += "Running health checks...\n"
	} else {
		allPassed := true
		for _, check := range m.checks {
			statusStr := ""
			if check.err == nil {
				statusStr = successStyle.Render("✓ PASS")
			} else {
				statusStr = errorStyle.Render("✗ FAIL")
				allPassed = false
			}

			s += labelStyle.Render(check.name+":") + " " + statusStr
			if check.err != nil {
				s += fmt.Sprintf(" (%v)", check.err)
			}
			s += "\n"
		}

		s += "\n"
		if allPassed {
			s += successStyle.Render("All checks passed! System is healthy.") + "\n"
		} else {
			s += errorStyle.Render("Some checks failed. Please review the errors above.") + "\n"
		}
	}

	return s
}

type doctorResultMsg struct {
	checks []checkResult
}

func runHealthChecks() tea.Cmd {
	return func() tea.Msg {
		checks := []checkResult{}

		// Check 1: Configuration file
		check1 := checkResult{name: "Configuration file"}
		var cfg config.Config
		var err error
		cfg, err = config.Load("config/config.json")
		if err != nil {
			check1.status = "FAIL"
			check1.err = err
		} else {
			check1.status = "PASS"
		}
		checks = append(checks, check1)

		// Check 2: Database connectivity
		check2 := checkResult{name: "Database connectivity"}
		if err == nil {
			if err := db.Init(cfg.Database); err != nil {
				check2.status = "FAIL"
				check2.err = err
			} else {
				check2.status = "PASS"
			}
		} else {
			check2.status = "SKIP"
			check2.err = fmt.Errorf("config not loaded")
		}
		checks = append(checks, check2)

		// Check 3: Database migrations
		check3 := checkResult{name: "Database schema"}
		if check2.err == nil {
			// Check if users table exists
			_, err := db.ListUsers(context.Background())
			if err != nil {
				check3.status = "FAIL"
				check3.err = fmt.Errorf("users table not found, run 'go-bastion migrate'")
			} else {
				check3.status = "PASS"
			}
		} else {
			check3.status = "SKIP"
			check3.err = fmt.Errorf("database not available")
		}
		checks = append(checks, check3)

		// Check 4: Template directory
		check4 := checkResult{name: "Template directory"}
		if _, err := os.Stat("templates"); os.IsNotExist(err) {
			check4.status = "FAIL"
			check4.err = fmt.Errorf("templates directory not found")
		} else {
			check4.status = "PASS"
		}
		checks = append(checks, check4)

		// Check 5: Template engine
		check5 := checkResult{name: "Template engine"}
		_, err = view.NewEngine("templates")
		if err != nil {
			check5.status = "FAIL"
			check5.err = err
		} else {
			check5.status = "PASS"
		}
		checks = append(checks, check5)

		// Check 6: JWT secret configuration
		check6 := checkResult{name: "JWT secret"}
		if err == nil && cfg.Security.JWTSecret != "" {
			check6.status = "PASS"
		} else {
			check6.status = "WARN"
			check6.err = fmt.Errorf("JWT secret is empty or default")
		}
		checks = append(checks, check6)

		return doctorResultMsg{checks: checks}
	}
}

func runDoctor() {
	p := tea.NewProgram(doctorModel{})
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running doctor: %v", err)
	}

	m := finalModel.(doctorModel)

	// Check if any critical checks failed
	criticalFailed := false
	for _, check := range m.checks {
		if check.err != nil && check.status == "FAIL" {
			criticalFailed = true
			break
		}
	}

	if criticalFailed {
		os.Exit(1)
	}
}
