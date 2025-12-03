package router

import (
	"net/http"

	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

// RegisterHomeRoutes registers the home page route
func RegisterHomeRoutes(r *frameworkrouter.Router, views *view.Engine) {
	r.Handle("GET", "/", handleHomePage(views))
}

// handleHomePage renders the home page
func handleHomePage(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		data := map[string]any{
			"Title": "Welcome to goBastion",
		}

		if err := views.Render(w, "home", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}
