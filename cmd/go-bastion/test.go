package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type testModel struct {
	verbose  bool
	status   string
	output   string
	done     bool
	err      error
	quitting bool
}

func (m testModel) Init() tea.Cmd {
	return runTestCmd(m.verbose)
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
	case testResultMsg:
		m.done = true
		m.err = msg.err
		m.output = msg.output
		if msg.err == nil {
			m.status = "completed successfully"
		} else {
			m.status = "completed with failures"
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m testModel) View() string {
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
	s += titleStyle.Render("RUNNING TESTS") + "\n\n"

	if !m.done {
		s += statusStyle.Render("Running go test ./...") + "\n"
		s += "\n  • Discovering test files\n"
		s += "  • Running test suites\n"
		s += "  • Collecting results\n"
	} else {
		if m.err == nil {
			s += successStyle.Render("✓ Tests "+m.status) + "\n\n"
		} else {
			s += errorStyle.Render("✗ Tests "+m.status) + "\n\n"
		}

		if m.output != "" {
			s += "Output:\n"
			s += m.output + "\n"
		}
	}

	return s
}

type testResultMsg struct {
	output string
	err    error
}

func runTestCmd(verbose bool) tea.Cmd {
	return func() tea.Msg {
		args := []string{"test", "./..."}
		if verbose {
			args = append(args, "-v")
		}

		cmd := exec.Command("go", args...)
		output, err := cmd.CombinedOutput()

		return testResultMsg{
			output: string(output),
			err:    err,
		}
	}
}

func runTests(verbose bool) {
	p := tea.NewProgram(testModel{verbose: verbose, status: "starting"})
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running tests: %v", err)
	}

	m := finalModel.(testModel)

	// Print summary
	if m.err != nil {
		// Check if there are no test files
		if strings.Contains(m.output, "no test files") || strings.Contains(m.output, "?") {
			fmt.Println("\nNote: No test files found in the project")
			fmt.Println("Consider adding unit tests for your application")
		} else {
			log.Fatalf("Tests failed: %v", m.err)
		}
	}
}
