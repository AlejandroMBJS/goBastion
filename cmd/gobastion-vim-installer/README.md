# goBastion Vim/Neovim Syntax Installer

Automated installer for goBastion template syntax highlighting in Vim and Neovim.

## Features

- 🎨 **Full Syntax Highlighting** for goBastion templates
- 🔍 **Auto-detects** installed Vim/Neovim instances
- 🖥️ **Cross-platform** (Linux, macOS, Windows)
- ⚡ **Zero dependencies** - single binary installation
- 🎯 **Smart Detection** - automatically finds your editor config

## Supported Editors

- ✅ Neovim (Linux/macOS: `~/.config/nvim`, Windows: `%LOCALAPPDATA%\nvim`)
- ✅ Vim (Linux/macOS: `~/.vim`, Windows: `~/vimfiles`)
- ✅ LazyVim and other Neovim distributions

## Supported File Extensions

The installer configures syntax highlighting for:

- `*.gb.html` - Primary goBastion template extension
- `*.gobastion.html` - Alternative extension
- `*.bastion.html` - Alternative extension
- `*.gb.tmpl` - Template extension

## Installation

### Build from Source

```bash
# From the goBastion root directory
go build -o gobastion-vim-installer ./cmd/gobastion-vim-installer/

# Run the installer
./gobastion-vim-installer
```

### Install Globally

```bash
# Install to your Go bin directory
go install ./cmd/gobastion-vim-installer

# Run from anywhere
gobastion-vim-installer
```

## Usage

### Interactive Installation

Simply run the installer and follow the prompts:

```bash
./gobastion-vim-installer
```

The installer will:
1. Detect all installed Vim/Neovim instances
2. Let you choose which editor to install to
3. Create necessary directories
4. Install syntax files
5. Verify the installation

### Command-line Options

```bash
# Show help
gobastion-vim-installer --help

# Show version
gobastion-vim-installer --version
```

## What Gets Installed

### For Neovim

**Filetype Detection** (`~/.config/nvim/ftdetect/gobastion.lua`):
- Lua-based filetype detection
- Automatically sets `gobastion` filetype for template files

**Syntax Highlighting** (`~/.config/nvim/syntax/gobastion.vim`):
- Full HTML syntax support
- `go::` keyword highlighting
- `::end` keyword highlighting
- `@expression` echo syntax highlighting
- Embedded Go syntax in logic blocks

### For Vim

**Filetype Detection** (`~/.vim/ftdetect/gobastion.vim`):
- Vimscript-based filetype detection
- Autocmd for goBastion file extensions

**Syntax Highlighting** (`~/.vim/syntax/gobastion.vim`):
- Same features as Neovim
- Compatible with classic Vim

## Syntax Highlighting Features

### 1. Logic Blocks

```html
go:: if user != nil {
  <p>Hello @user.Name</p>
::end
```

- `go::` highlighted as **Keyword**
- Go code syntax highlighted
- `::end` highlighted as **Keyword**

### 2. Echo Expressions

```html
<h1>@.Title</h1>
<p>Email: @user.Email</p>
<p>@formatPrice(product.Price)</p>
```

- `@` highlighted as **Operator**
- Variable/expression highlighted as **Constant**

### 3. HTML Support

Full HTML syntax highlighting including:
- Tags
- Attributes
- Strings
- Comments

## Testing the Installation

After installation:

1. **Restart your editor**
2. **Open a test file**:
   ```bash
   # Create a test file
   cat > test.gb.html << 'EOF'
   <!DOCTYPE html>
   <html>
   <body>
     <h1>@.Title</h1>
     go:: if .User
       <p>Hello @.User.Name</p>
     ::end
   </body>
   </html>
   EOF

   # Open in your editor
   nvim test.gb.html
   ```

3. **Verify syntax highlighting** works for:
   - HTML tags
   - `go::` keywords
   - `@` expressions
   - `::end` keywords

## Troubleshooting

### Editor Not Detected

**Problem**: Installer says "No Vim or Neovim installation detected"

**Solutions**:
- Ensure Vim/Neovim is installed: `which vim` or `which nvim`
- Check that config directory exists:
  - Neovim: `~/.config/nvim/` (create with `mkdir -p ~/.config/nvim`)
  - Vim: `~/.vim/` (create with `mkdir -p ~/.vim`)

### Syntax Not Working

**Problem**: Syntax highlighting doesn't appear

**Solutions**:
1. **Restart your editor** completely
2. **Check filetype**: `:set filetype?` should show `gobastion`
3. **Manually set filetype**: `:set filetype=gobastion`
4. **Verify files exist**:
   ```bash
   # For Neovim
   ls ~/.config/nvim/ftdetect/gobastion.lua
   ls ~/.config/nvim/syntax/gobastion.vim

   # For Vim
   ls ~/.vim/ftdetect/gobastion.vim
   ls ~/.vim/syntax/gobastion.vim
   ```
5. **Check syntax loading**: `:syntax` should show goBastion syntax rules

### Wrong File Extension

**Problem**: Only want `.gb.html` detection, not all extensions

**Solution**: Edit the ftdetect file and remove unwanted patterns:
```bash
# For Neovim
nvim ~/.config/nvim/ftdetect/gobastion.lua

# For Vim
vim ~/.vim/ftdetect/gobastion.vim
```

## Manual Installation

If you prefer manual installation, see the shell script:
```bash
./goBastionTemplates/gBTemplatesNvim.sh
```

## Uninstallation

To remove goBastion syntax highlighting:

```bash
# For Neovim
rm ~/.config/nvim/ftdetect/gobastion.lua
rm ~/.config/nvim/syntax/gobastion.vim

# For Vim
rm ~/.vim/ftdetect/gobastion.vim
rm ~/.vim/syntax/gobastion.vim
```

## Development

### Building

```bash
# Build for current platform
go build -o gobastion-vim-installer ./cmd/gobastion-vim-installer/

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o gobastion-vim-installer-linux ./cmd/gobastion-vim-installer/
GOOS=darwin GOARCH=amd64 go build -o gobastion-vim-installer-macos ./cmd/gobastion-vim-installer/
GOOS=windows GOARCH=amd64 go build -o gobastion-vim-installer.exe ./cmd/gobastion-vim-installer/
```

### Testing

```bash
# Test the binary
./gobastion-vim-installer --help
./gobastion-vim-installer --version

# Run with dry-run (not yet implemented)
# ./gobastion-vim-installer --dry-run
```

## VS Code Extension

For VS Code users, see the VS Code extension in `goBastionTemplates/`:

```bash
cd goBastionTemplates
# Install vsce
npm install -g @vscode/vsce

# Package extension
vsce package

# Install the .vsix file via VS Code
```

## License

MIT License - Free to use in personal and commercial projects.

## Support

- 📚 **Documentation**: See [TEMPLATE_SYNTAX.md](../../TEMPLATE_SYNTAX.md)
- 🐛 **Issues**: Report bugs on GitHub
- 💬 **Questions**: Check the main goBastion README

---

**Built with ❤️ for goBastion**
