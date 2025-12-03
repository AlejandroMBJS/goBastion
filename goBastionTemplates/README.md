# goBastion Templates - VS Code Extension

Syntax highlighting for goBastion template files in Visual Studio Code.

## Features

- 🎨 **Full Syntax Highlighting** for goBastion template syntax
- 🔍 **File Extension Support** for `.gb.html`, `.gobastion.html`, `.bastion.html`, and `.gb.tmpl`
- 🏷️ **Smart Highlighting** for:
  - `go::` logic blocks with embedded Go syntax
  - `::end` closing keywords
  - `@expression` echo syntax
  - Full HTML support as base language

## Supported File Extensions

This extension activates for the following file types:

- `*.gb.html` - Primary goBastion template extension
- `*.gobastion.html` - Alternative extension
- `*.bastion.html` - Alternative extension
- `*.gb.tmpl` - Template extension

## Syntax Examples

### Echo Expressions

```html
<h1>@.Title</h1>
<p>Hello, @user.Name!</p>
<p>Email: @user.Email</p>
```

### Logic Blocks

```html
go:: if user != nil {
  <p>Welcome, @user.Name!</p>
::end

go:: range .Items
  <li>@.Name - $@.Price</li>
::end
```

### Complete Example

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>@.Title</title>
</head>
<body>
  <h1>@.Title</h1>

  go:: if .Error
  <div class="error">@.Error</div>
  ::end

  go:: if .Users
  <table>
    go:: range .Users
    <tr>
      <td>@.ID</td>
      <td>@.Name</td>
      <td>@.Email</td>
    </tr>
    ::end
  </table>
  go:: else
  <p>No users found</p>
  ::end
</body>
</html>
```

## Installation

### From VSIX File

1. Download the `.vsix` file
2. Open VS Code
3. Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on macOS)
4. Type "Extensions: Install from VSIX..."
5. Select the downloaded `.vsix` file
6. Reload VS Code

### From Source

```bash
# Install vsce
npm install -g @vscode/vsce

# Package the extension
cd goBastionTemplates
vsce package

# Install the generated .vsix file
code --install-extension gobastion-templates-0.1.0.vsix
```

## Requirements

- Visual Studio Code version 1.80.0 or higher
- Go extension (recommended for embedded Go syntax highlighting)

## Usage

1. Open any file with a supported extension (`.gb.html`, etc.)
2. Syntax highlighting will activate automatically
3. Start writing goBastion templates with full highlighting support

## Language Features

### Auto-Closing Pairs

The extension provides automatic closing for:
- HTML tags: `<` → `</>`
- Braces: `{` → `}`
- Parentheses: `(` → `)`
- Brackets: `[` → `]`
- Quotes: `"` and `'`

### Bracket Matching

Matching pairs are highlighted when your cursor is next to them.

### Comments

HTML-style comments are supported:
```html
<!-- This is a comment -->
```

## About goBastion

goBastion is a modern Go web framework with a custom template engine that combines:
- Clean, Go-like syntax (`go::` / `@` constructs)
- Automatic HTML escaping for security
- Full power of Go's `html/template`
- Beautiful, Tailwind-styled UI components

Learn more at: [goBastion GitHub Repository](https://github.com/AlejandroMBJS/goBastion)

## Extension Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/AlejandroMBJS/goBastion.git
cd goBastion/goBastionTemplates

# Install dependencies (if any)
npm install

# Package the extension
vsce package
```

### Testing

1. Open this folder in VS Code
2. Press `F5` to launch Extension Development Host
3. Open `sample.gb.html` to test syntax highlighting

### File Structure

```
goBastionTemplates/
├── package.json                      # Extension manifest
├── language-configuration.json       # Language configuration
├── syntaxes/
│   └── gobastion.tmLanguage.json    # TextMate grammar
├── sample.gb.html                    # Example file
└── README.md                         # This file
```

## Known Issues

None at this time. Please report issues on GitHub.

## Release Notes

### 0.1.0

Initial release:
- Syntax highlighting for goBastion templates
- Support for `.gb.html`, `.gobastion.html`, `.bastion.html`, and `.gb.tmpl` files
- Auto-closing pairs and bracket matching
- Embedded Go syntax highlighting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - Free to use in personal and commercial projects.

---

**Enjoy building with goBastion!** 🎨
