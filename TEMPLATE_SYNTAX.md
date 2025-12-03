# goBastion Template Syntax

goBastion uses a custom template engine built on top of Go's `html/template` package. The engine provides a clean, intuitive syntax while maintaining all the security benefits of Go templates, including automatic HTML escaping.

## Overview

The goBastion template syntax uses **only two constructs**:

1. **Logic Blocks** (`go:: ... ::end`) - For control flow and logic
2. **Echo Expressions** (`@expr`) - For outputting values

This minimal, expressive syntax keeps templates clean and readable while providing all the power you need.

---

## Echo Expressions

Echo expressions output values with automatic HTML escaping.

### Syntax

```html
@expression
```

### Examples

**Simple variable:**
```html
<h1>@.Title</h1>
```

**Object property:**
```html
<p>Hello, @user.Name!</p>
<p>Email: @user.Email</p>
```

**Nested properties:**
```html
<img src="@user.Profile.Avatar" alt="Avatar">
```

**Function calls:**
```html
<p>Price: @formatPrice(product.Price)</p>
<p>@upper(.Title)</p>
```

### HTML Escaping

All echo expressions are **automatically HTML-escaped** for security:

```html
@userInput
<!-- If userInput contains: <script>alert('xss')</script> -->
<!-- Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; -->
```

This prevents XSS (Cross-Site Scripting) attacks by default.

---

## Logic Blocks

Logic blocks control template flow using real Go code.

### Syntax

```html
go:: <go-statement>
  <!-- template content -->
::end
```

### If Statements

**Basic if:**
```html
go:: if .User
  <p>Welcome, @.User.Name!</p>
::end
```

**If/else:**
```html
go:: if .Error
  <div class="error">@.Error</div>
go:: else
  <div class="success">Operation successful!</div>
::end
```

**If with comparison:**
```html
go:: if eq .Role "admin"
  <a href="/admin">Admin Panel</a>
::end
```

### Range Loops

**Iterate over a slice:**
```html
<ul>
go:: range .Items
  <li>@.Name - $@.Price</li>
::end
</ul>
```

**Range with index and value:**
```html
<ol>
go:: range $index, $item := .Items
  <li>Item #@$index: @$item.Name</li>
::end
</ol>
```

**Empty list handling:**
```html
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
<p>No users found.</p>
::end
```

### With Blocks

Set the context to a specific value:

```html
go:: with .User
  <h2>@.Name</h2>
  <p>@.Email</p>
  <p>Role: @.Role</p>
::end
```

### Nested Blocks

You can nest blocks as deeply as needed:

```html
go:: if .Posts
<div class="posts">
  go:: range .Posts
  <article>
    <h2>@.Title</h2>
    <p>@.Content</p>
    go:: if .Comments
    <div class="comments">
      <h3>Comments</h3>
      go:: range .Comments
      <div class="comment">
        <strong>@.Author:</strong> @.Text
      </div>
      ::end
    </div>
    ::end
  </article>
  ::end
</div>
::end
```

---

## Complete Example

Here's a full template showing various features:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>@.Title - goBastion</title>
</head>
<body>
    <header>
        <h1>@.Title</h1>
        go:: if .User
        <p>Welcome, @.User.Name! | <a href="/logout">Logout</a></p>
        go:: else
        <p><a href="/login">Login</a> | <a href="/register">Register</a></p>
        ::end
    </header>

    <main>
        go:: if .Error
        <div class="alert alert-danger">
            @.Error
        </div>
        ::end

        go:: if .Success
        <div class="alert alert-success">
            @.Success
        </div>
        ::end

        go:: if .Users
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                go:: range .Users
                <tr>
                    <td>@.ID</td>
                    <td>@.Name</td>
                    <td>@.Email</td>
                    <td>
                        go:: if eq .Role "admin"
                        <span class="badge badge-danger">Admin</span>
                        go:: else
                        <span class="badge badge-success">User</span>
                        ::end
                    </td>
                    <td>
                        go:: if .IsActive
                        <span class="text-success">Active</span>
                        go:: else
                        <span class="text-muted">Inactive</span>
                        ::end
                    </td>
                </tr>
                ::end
            </tbody>
        </table>
        go:: else
        <p>No users found.</p>
        ::end
    </main>

    <footer>
        <p>&copy; 2025 goBastion</p>
    </footer>
</body>
</html>
```

---

## Template Functions

goBastion provides built-in template functions:

| Function | Description | Example |
|----------|-------------|---------|
| `upper` | Convert to uppercase | `{{ upper .Text }}` |
| `lower` | Convert to lowercase | `{{ lower .Text }}` |
| `title` | Title case | `{{ title .Text }}` |
| `eq` | Equal comparison | `{{ eq .Role "admin" }}` |

You can use these in logic blocks or with the standard `{{ }}` syntax.

---

## Before/After Comparison

### Old PHP-style Syntax (Deprecated)

```html
<?php if ($user != nil) { ?>
  <p>Hello <?= $user->name ?></p>
  <ul>
  <?php foreach ($items as $item) { ?>
    <li><?= $item ?></li>
  <?php } ?>
  </ul>
<?php } ?>
```

### New goBastion Syntax

```html
go:: if .User
  <p>Hello @.User.Name</p>
  <ul>
  go:: range .Items
    <li>@.</li>
  ::end
  </ul>
::end
```

**Benefits of the new syntax:**
- ✅ Cleaner, more Go-like
- ✅ No PHP confusion
- ✅ Easier to read and write
- ✅ Better editor support
- ✅ Same security guarantees

---

## Backward Compatibility

The old PHP-style tags (`<?`, `<?=`, `?>`) are still supported for backward compatibility but are **deprecated**. You should migrate to the new syntax.

Old templates will continue to work, but we recommend updating them to use `go::` and `@` syntax.

---

## Best Practices

### 1. **Keep templates simple**
Complex logic belongs in handlers, not templates.

**❌ Bad:**
```html
go:: if and (gt .User.Age 18) (eq .User.Country "US") (not .User.Banned)
  <button>Access Content</button>
::end
```

**✅ Good (in handler):**
```go
data := map[string]any{
    "CanAccess": user.Age > 18 && user.Country == "US" && !user.Banned,
}
```
```html
go:: if .CanAccess
  <button>Access Content</button>
::end
```

### 2. **Use meaningful variable names**
```html
go:: range $user := .Users
  <li>@$user.Name</li>
::end
```

### 3. **Check for empty collections**
```html
go:: if .Items
  <!-- show items -->
go:: else
  <p>No items available</p>
::end
```

### 4. **Leverage HTML escaping**
Never use raw HTML unless absolutely necessary. The `@` syntax escapes by default.

---

## Security

### Automatic Escaping

All `@expr` outputs are **automatically HTML-escaped**:

```html
@userComment
<!-- Input: <script>alert('xss')</script> -->
<!-- Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; -->
```

### CSRF Protection

goBastion includes built-in CSRF protection. Use CSRF tokens in forms:

```html
<form method="POST" action="/submit">
    go:: if .CSRFToken
    <input type="hidden" name="csrf_token" value="@.CSRFToken">
    ::end

    <!-- form fields -->
    <button type="submit">Submit</button>
</form>
```

---

## Troubleshooting

### Common Errors

**1. Missing `::end`**
```
Error: unexpected EOF, expected {{ end }}
```
Make sure every `go::` has a matching `::end`.

**2. Syntax error in Go statement**
```
Error: unexpected "}", expected expression
```
Check your Go syntax inside `go::` blocks.

**3. Undefined variable**
```
Error: can't evaluate field User in type map[string]interface{}
```
Make sure you're passing the correct data to the template.

---

## Summary

- **`@expr`** - Echo expressions (HTML-escaped)
- **`go:: ... ::end`** - Logic blocks (if, range, with)
- Built on Go's `html/template` for security
- Clean, readable syntax
- Backward compatible with old PHP-style tags

Happy templating! 🎨
