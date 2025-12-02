package docs

import (
	"net/http"

	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
)

// OpenAPIJSON contains the OpenAPI 3.0 specification
var OpenAPIJSON = []byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "Go Native FastAPI",
    "description": "A production-ready web API framework built with pure Go standard library",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Development server"
    }
  ],
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "User": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "name": { "type": "string" },
          "email": { "type": "string", "format": "email" },
          "role": { "type": "string", "enum": ["user", "admin"] },
          "is_active": { "type": "boolean" },
          "is_staff": { "type": "boolean" },
          "is_superuser": { "type": "boolean" }
        }
      },
      "UserInput": {
        "type": "object",
        "required": ["name", "email"],
        "properties": {
          "name": { "type": "string", "minLength": 2, "maxLength": 100 },
          "email": { "type": "string", "format": "email", "maxLength": 255 },
          "role": { "type": "string", "enum": ["user", "admin"] }
        }
      },
      "RegisterInput": {
        "type": "object",
        "required": ["name", "email", "password"],
        "properties": {
          "name": { "type": "string", "minLength": 2, "maxLength": 100 },
          "email": { "type": "string", "format": "email", "maxLength": 255 },
          "password": { "type": "string", "minLength": 8, "maxLength": 100 },
          "role": { "type": "string", "enum": ["user", "admin"] }
        }
      },
      "LoginInput": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": { "type": "string", "format": "email" },
          "password": { "type": "string" }
        }
      },
      "RefreshInput": {
        "type": "object",
        "required": ["refresh_token"],
        "properties": {
          "refresh_token": { "type": "string" }
        }
      },
      "TokenResponse": {
        "type": "object",
        "properties": {
          "access_token": { "type": "string" },
          "refresh_token": { "type": "string" },
          "token_type": { "type": "string" }
        }
      },
      "Error": {
        "type": "object",
        "properties": {
          "error": { "type": "string" }
        }
      }
    }
  },
  "paths": {
    "/api/v1/auth/register": {
      "post": {
        "summary": "Register a new user",
        "tags": ["Authentication"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RegisterInput" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User registered successfully",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "user": { "$ref": "#/components/schemas/User" },
                    "access_token": { "type": "string" },
                    "refresh_token": { "type": "string" },
                    "token_type": { "type": "string" }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Invalid input",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          },
          "409": {
            "description": "User already exists",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "summary": "Login with email and password",
        "tags": ["Authentication"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/LoginInput" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login successful",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "user": { "$ref": "#/components/schemas/User" },
                    "access_token": { "type": "string" },
                    "refresh_token": { "type": "string" },
                    "token_type": { "type": "string" }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Invalid credentials",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/auth/refresh": {
      "post": {
        "summary": "Refresh access token",
        "tags": ["Authentication"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RefreshInput" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Token refreshed successfully",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/TokenResponse" }
              }
            }
          },
          "401": {
            "description": "Invalid or expired token",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/auth/me": {
      "get": {
        "summary": "Get current user profile",
        "tags": ["Authentication"],
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": {
            "description": "Current user profile",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/User" }
              }
            }
          },
          "401": {
            "description": "Unauthorized",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/users": {
      "get": {
        "summary": "List all users",
        "tags": ["Users"],
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": {
            "description": "List of users",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/User" }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a new user",
        "tags": ["Users"],
        "security": [{ "bearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RegisterInput" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/User" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/users/{id}": {
      "get": {
        "summary": "Get a user by ID",
        "tags": ["Users"],
        "security": [{ "bearerAuth": [] }],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": { "type": "integer" }
          }
        ],
        "responses": {
          "200": {
            "description": "User details",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/User" }
              }
            }
          },
          "404": {
            "description": "User not found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      },
      "put": {
        "summary": "Update a user",
        "tags": ["Users"],
        "security": [{ "bearerAuth": [] }],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": { "type": "integer" }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/UserInput" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "User updated",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/User" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          },
          "404": {
            "description": "User not found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      },
      "delete": {
        "summary": "Delete a user",
        "tags": ["Users"],
        "security": [{ "bearerAuth": [] }],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": { "type": "integer" }
          }
        ],
        "responses": {
          "204": { "description": "User deleted" },
          "404": {
            "description": "User not found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    }
  }
}`)

// HandlerJSON returns the OpenAPI JSON specification
func HandlerJSON(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(OpenAPIJSON)
}

// HandlerUI returns the Swagger UI HTML
func HandlerUI(w http.ResponseWriter, r *http.Request, params map[string]string) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/docs/openapi.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// RegisterRoutes registers documentation routes
func RegisterRoutes(r *frameworkrouter.Router) {
	r.Handle("GET", "/docs/openapi.json", HandlerJSON)
	r.Handle("GET", "/docs", HandlerUI)
}
