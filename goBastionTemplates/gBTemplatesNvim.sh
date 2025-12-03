#!/bin/bash

# Configuration
NVIM_CONFIG_DIR="$HOME/.config/nvim"
FTDETECT_DIR="$NVIM_CONFIG_DIR/ftdetect"
SYNTAX_DIR="$NVIM_CONFIG_DIR/syntax"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== goBastion LazyVim/Neovim Installer ===${NC}"

# Check if Neovim config exists
if [ ! -d "$NVIM_CONFIG_DIR" ]; then
  echo "Error: Neovim config directory ($NVIM_CONFIG_DIR) not found."
  echo "Please ensure Neovim/LazyVim is installed."
  exit 1
fi

# Create directories
echo -e "Creating directories..."
mkdir -p "$FTDETECT_DIR"
mkdir -p "$SYNTAX_DIR"

# 1. Create FileType Detection (Lua)
echo -e "Installing filetype detection to ${BLUE}$FTDETECT_DIR/gobastion.lua${NC}..."
cat <<EOF >"$FTDETECT_DIR/gobastion.lua"
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = { "*.gb.html", "*.gobastion.html", "*.bastion.html", "*.gb.tmpl" },
  callback = function()
    vim.bo.filetype = "gobastion"
  end,
})
EOF

# 2. Create Syntax Highlighting (Vimscript)
echo -e "Installing syntax highlighting to ${BLUE}$SYNTAX_DIR/gobastion.vim${NC}..."
cat <<EOF >"$SYNTAX_DIR/gobastion.vim"
" Place this in ~/.config/nvim/syntax/gobastion.vim

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
EOF

echo -e "${GREEN}Success! goBastion support installed.${NC}"
echo -e "Restart Neovim and open a ${BLUE}.gb.html${NC} file to see changes."
