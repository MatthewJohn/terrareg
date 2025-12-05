package template

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// Renderer handles HTML template rendering
type Renderer struct {
	templates *template.Template
	config    *config.Config
	mu        sync.RWMutex
}

// NewRenderer creates a new template renderer
func NewRenderer(cfg *config.Config) (*Renderer, error) {
	r := &Renderer{
		config: cfg,
	}

	if err := r.loadTemplates(); err != nil {
		return nil, err
	}

	return r, nil
}

// loadTemplates loads all HTML templates and handles inheritance
func (r *Renderer) loadTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load base template first
	baseTemplate := template.New("")

	// Parse base template (template.html) only
	baseTemplate, err := baseTemplate.ParseFiles("templates/template.html")
	if err != nil {
		return err
	}

	r.templates = baseTemplate
	return nil
}

// getThemePath returns CSS path for theme based on theme cookie or session
func (r *Renderer) getThemePath(ctx context.Context, request *http.Request) string {
	// Try to get theme from session first - Note: Theme is not in new SessionData
	// We can add it later if needed, for now fall back to cookie
	/*
	if sessionData := middleware.GetSessionData(ctx); sessionData != nil && sessionData.Theme != "" {
		if r.isValidTheme(sessionData.Theme) {
			return r.buildThemePath(sessionData.Theme)
		}
	}
	*/

	// Fall back to cookie (like Python version)
	if request != nil {
		cookie, err := request.Cookie("theme")
		if err == nil && r.isValidTheme(cookie.Value) {
			return r.buildThemePath(cookie.Value)
		}
	}

	// Default theme
	return r.buildThemePath("default")
}

// isValidTheme checks if the theme is valid
func (r *Renderer) isValidTheme(theme string) bool {
	validThemes := []string{"default", "lux", "pulse", "cherry-dark"}
	for _, validTheme := range validThemes {
		if theme == validTheme {
			return true
		}
	}
	return false
}

// buildThemePath builds the theme path based on theme name
func (r *Renderer) buildThemePath(theme string) string {
	baseURL := "/static/css/bulma/"

	// Use bulmaswatch themes for specific themes
	if theme == "lux" || theme == "pulse" || theme == "cherry-dark" {
		return baseURL + theme + "/bulmaswatch.min.css"
	}

	// Default theme uses standard Bulma CSS
	return baseURL + "bulma-0.9.3.min.css"
}

// RenderWithContext renders a template with the given data and context
func (r *Renderer) RenderWithContext(ctx context.Context, w io.Writer, name string, data map[string]interface{}, request *http.Request) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Add global template data
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add common config values accessible to all templates (from configuration)
	data["terrareg_application_name"] = r.config.ApplicationName
	data["public_url"] = r.config.PublicURL
	data["enable_access_controls"] = r.config.EnableAccessControls
	data["enable_security_scanning"] = r.config.EnableSecurityScanning
	data["terrareg_logo_url"] = r.config.LogoURL
	data["theme_path"] = r.getThemePath(ctx, request)
	data["SITE_WARNING"] = r.config.SiteWarning

	// Add CSRF token from session context
	data["csrf_token"] = middleware.GetCSRFToken(ctx)

	// Add authentication context for templates
	sessionData := middleware.GetSessionData(ctx)
	if sessionData != nil {
		data["session_authenticated"] = true
		data["session_username"] = sessionData.Username
		data["session_user_id"] = sessionData.SessionID // Use SessionID as fallback since UserID is not in new SessionData
		data["session_is_admin"] = sessionData.IsAdmin
		data["session_auth_method"] = sessionData.AuthMethod
		data["session_user_groups"] = []string{} // Empty slice since UserGroups is not in new SessionData
	} else {
		data["session_authenticated"] = false
		data["session_username"] = ""
		data["session_user_id"] = ""
		data["session_is_admin"] = false
		data["session_auth_method"] = ""
		data["session_user_groups"] = []string{}
	}

	// Load the requested template individually to avoid conflicts
	templatePath := filepath.Join("templates", name)
	tmpl, err := r.templates.Clone()
	if err != nil {
		return err
	}

	// Parse the specific template file with the base template
	tmpl, err = tmpl.ParseFiles("templates/template.html", templatePath)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

// Render renders a template with the given data (maintains backward compatibility)
func (r *Renderer) Render(w io.Writer, name string, data map[string]interface{}) error {
	return r.RenderWithContext(context.Background(), w, name, data, nil)
}

// RenderWithRequest renders a template with the given data, context, and HTTP request
func (r *Renderer) RenderWithRequest(ctx context.Context, w io.Writer, name string, data map[string]interface{}, request *http.Request) error {
	return r.RenderWithContext(ctx, w, name, data, request)
}

// Reload reloads all templates (useful in development)
func (r *Renderer) Reload() error {
	return r.loadTemplates()
}
