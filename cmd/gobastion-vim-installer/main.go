package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	version = "1.0.0"

	// Syntax highlighting content
	ftdetectContent = `vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = { "*.gb.html", "*.gobastion.html", "*.bastion.html", "*.gb.tmpl" },
  callback = function()
    vim.bo.filetype = "gobastion"
  end,
})
`

	syntaxContent = `" goBastion Template Syntax Highlighting
" Place this in ~/.config/nvim/syntax/gobastion.vim or ~/.vim/syntax/gobastion.vim

" 1. Load the base HTML syntax first
runtime! syntax/html.vim
unlet b:current_syntax

" 2. Define goBastion Regions and Matches

" Match the ::end keyword
syn match goBastionEnd "^\s*::end" containedin=ALL

" Match the @expression (simple regex for variables/functions)
syn match goBastionEcho "@[a-zA-Z0-9_.]\+\(\(.*?\)\)\?" containedin=htmlString,htmlTag,htmlText

" Match the go:: logic line
" We use a region to capture the whole line starting with go::
syn region goBastionLogicLine start="^\s*go::" end="$" keepend contains=goBastionKeyword,goBastionGoCode

" Highlight 'go::' specifically
syn match goBastionKeyword "go::" contained

" Attempt to include basic Go syntax inside the line (optional/advanced)
" This loads Go syntax but limits it to the logic line region
syn include @GoSyntax syntax/go.vim
syn region goBastionGoCode start="." end="$" contained contains=@GoSyntax

" 3. Link to standard Highlight groups (LazyVim themes use these)
hi def link goBastionKeyword  Keyword
hi def link goBastionEnd      Keyword
hi def link goBastionEcho     Constant
hi def link goBastionLogicLine Normal

let b:current_syntax = "gobastion"
`
)

// EditorConfig represents configuration for an editor
type EditorConfig struct {
	Name          string
	ConfigDir     string
	FtdetectFile  string
	SyntaxFile    string
	DetectionType string // "lua" or "vim"
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║  goBastion Vim/Neovim Syntax Installer          ║")
	fmt.Printf("║  Version %-39s ║\n", version)
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	// Detect editors
	editors := detectEditors()

	if len(editors) == 0 {
		fmt.Println("❌ No Vim or Neovim installation detected.")
		fmt.Println("   Please install Vim or Neovim first.")
		os.Exit(1)
	}

	fmt.Printf("✓ Detected %d editor(s):\n", len(editors))
	for i, editor := range editors {
		fmt.Printf("  %d. %s (%s)\n", i+1, editor.Name, editor.ConfigDir)
	}
	fmt.Println()

	// Ask user which editor to install to
	var choice int
	if len(editors) == 1 {
		choice = 1
		fmt.Printf("Installing to %s...\n\n", editors[0].Name)
	} else {
		fmt.Print("Select editor (enter number): ")
		_, err := fmt.Scanf("%d", &choice)
		if err != nil || choice < 1 || choice > len(editors) {
			fmt.Println("❌ Invalid choice")
			os.Exit(1)
		}
		fmt.Println()
	}

	selectedEditor := editors[choice-1]

	// Install syntax files
	if err := installSyntaxFiles(selectedEditor); err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n╔══════════════════════════════════════════════════╗")
	fmt.Println("║  ✓ Installation Complete!                       ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart your editor")
	fmt.Println("  2. Open any .gb.html file")
	fmt.Println("  3. Enjoy syntax highlighting! 🎨")
	fmt.Println()
}

// detectEditors detects installed Vim and Neovim
func detectEditors() []EditorConfig {
	var editors []EditorConfig
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return editors
	}

	// Check for Neovim
	nvimConfigDir := filepath.Join(homeDir, ".config", "nvim")
	if _, err := os.Stat(nvimConfigDir); err == nil {
		// Check if nvim is in PATH
		if _, err := exec.LookPath("nvim"); err == nil {
			editors = append(editors, EditorConfig{
				Name:          "Neovim",
				ConfigDir:     nvimConfigDir,
				FtdetectFile:  filepath.Join(nvimConfigDir, "ftdetect", "gobastion.lua"),
				SyntaxFile:    filepath.Join(nvimConfigDir, "syntax", "gobastion.vim"),
				DetectionType: "lua",
			})
		}
	}

	// Check for Vim
	vimConfigDir := filepath.Join(homeDir, ".vim")
	if _, err := os.Stat(vimConfigDir); err == nil {
		// Check if vim is in PATH
		if _, err := exec.LookPath("vim"); err == nil {
			editors = append(editors, EditorConfig{
				Name:          "Vim",
				ConfigDir:     vimConfigDir,
				FtdetectFile:  filepath.Join(vimConfigDir, "ftdetect", "gobastion.vim"),
				SyntaxFile:    filepath.Join(vimConfigDir, "syntax", "gobastion.vim"),
				DetectionType: "vim",
			})
		}
	}

	// Windows-specific paths
	if runtime.GOOS == "windows" {
		// Check for Neovim on Windows
		nvimConfigDirWin := filepath.Join(homeDir, "AppData", "Local", "nvim")
		if _, err := os.Stat(nvimConfigDirWin); err == nil {
			if _, err := exec.LookPath("nvim"); err == nil {
				editors = append(editors, EditorConfig{
					Name:          "Neovim (Windows)",
					ConfigDir:     nvimConfigDirWin,
					FtdetectFile:  filepath.Join(nvimConfigDirWin, "ftdetect", "gobastion.lua"),
					SyntaxFile:    filepath.Join(nvimConfigDirWin, "syntax", "gobastion.vim"),
					DetectionType: "lua",
				})
			}
		}

		// Check for Vim on Windows
		vimConfigDirWin := filepath.Join(homeDir, "vimfiles")
		if _, err := os.Stat(vimConfigDirWin); err == nil {
			if _, err := exec.LookPath("vim"); err == nil {
				editors = append(editors, EditorConfig{
					Name:          "Vim (Windows)",
					ConfigDir:     vimConfigDirWin,
					FtdetectFile:  filepath.Join(vimConfigDirWin, "ftdetect", "gobastion.vim"),
					SyntaxFile:    filepath.Join(vimConfigDirWin, "syntax", "gobastion.vim"),
					DetectionType: "vim",
				})
			}
		}
	}

	return editors
}

// installSyntaxFiles installs syntax files for the selected editor
func installSyntaxFiles(config EditorConfig) error {
	// Create directories
	ftdetectDir := filepath.Dir(config.FtdetectFile)
	syntaxDir := filepath.Dir(config.SyntaxFile)

	fmt.Printf("📁 Creating directories...\n")
	if err := os.MkdirAll(ftdetectDir, 0755); err != nil {
		return fmt.Errorf("failed to create ftdetect directory: %w", err)
	}
	fmt.Printf("   ✓ %s\n", ftdetectDir)

	if err := os.MkdirAll(syntaxDir, 0755); err != nil {
		return fmt.Errorf("failed to create syntax directory: %w", err)
	}
	fmt.Printf("   ✓ %s\n", syntaxDir)

	// Write filetype detection file
	fmt.Printf("\n📝 Installing filetype detection...\n")
	var ftContent string
	if config.DetectionType == "lua" {
		ftContent = ftdetectContent
	} else {
		// For classic Vim, use vimscript
		ftContent = `" goBastion filetype detection
au BufRead,BufNewFile *.gb.html,*.gobastion.html,*.bastion.html,*.gb.tmpl set filetype=gobastion
`
	}

	if err := os.WriteFile(config.FtdetectFile, []byte(ftContent), 0644); err != nil {
		return fmt.Errorf("failed to write ftdetect file: %w", err)
	}
	fmt.Printf("   ✓ %s\n", config.FtdetectFile)

	// Write syntax file
	fmt.Printf("\n🎨 Installing syntax highlighting...\n")
	if err := os.WriteFile(config.SyntaxFile, []byte(syntaxContent), 0644); err != nil {
		return fmt.Errorf("failed to write syntax file: %w", err)
	}
	fmt.Printf("   ✓ %s\n", config.SyntaxFile)

	// Verify installation
	fmt.Printf("\n🔍 Verifying installation...\n")
	if !fileExists(config.FtdetectFile) {
		return fmt.Errorf("ftdetect file not found after installation")
	}
	if !fileExists(config.SyntaxFile) {
		return fmt.Errorf("syntax file not found after installation")
	}
	fmt.Printf("   ✓ All files installed successfully\n")

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getVersion returns the version string
func getVersion() string {
	return version
}

// showHelp displays help information
func showHelp() {
	fmt.Println("goBastion Vim/Neovim Syntax Installer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gobastion-vim-installer        Install syntax highlighting")
	fmt.Println("  gobastion-vim-installer -v     Show version")
	fmt.Println("  gobastion-vim-installer -h     Show this help")
	fmt.Println()
	fmt.Println("Supported file extensions:")
	fmt.Println("  • .gb.html")
	fmt.Println("  • .gobastion.html")
	fmt.Println("  • .bastion.html")
	fmt.Println("  • .gb.tmpl")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  • Syntax highlighting for go:: logic blocks")
	fmt.Println("  • Highlighting for @expression echo syntax")
	fmt.Println("  • Highlighting for ::end keywords")
	fmt.Println("  • Full HTML syntax support")
	fmt.Println("  • Go syntax highlighting inside logic blocks")
	fmt.Println()
}

func init() {
	// Check for flags
	args := os.Args[1:]
	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "-v", "--version":
			fmt.Printf("gobastion-vim-installer version %s\n", version)
			os.Exit(0)
		case "-h", "--help":
			showHelp()
			os.Exit(0)
		}
	}
}
