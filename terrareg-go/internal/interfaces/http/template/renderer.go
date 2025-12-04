package template

import (
	"context"
	"html/template"
	"io"
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

// RenderWithContext renders a template with the given data and context
func (r *Renderer) RenderWithContext(ctx context.Context, w io.Writer, name string, data map[string]interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Add global template data
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add common config values accessible to all templates
	data["terrareg_application_name"] = "Terrareg" // TODO: Add to config
	data["public_url"] = r.config.PublicURL
	data["enable_access_controls"] = r.config.EnableAccessControls
	data["enable_security_scanning"] = false              // TODO: Add to config
	data["terrareg_logo_url"] = "/static/images/logo.png" // TODO: Add to config
	data["theme_path"] = "/static/css/bulma/pulse/bulmaswatch.min.css"  // TODO: Add to config
	data["SITE_WARNING"] = ""                             // TODO: Add to config

	// Add CSRF token from session context
	data["csrf_token"] = middleware.GetCSRFToken(ctx)

	// Add authentication context for templates
	sessionData := middleware.GetSessionData(ctx)
	if sessionData != nil {
		data["session_authenticated"] = true
		data["session_username"] = sessionData.Username
		data["session_user_id"] = sessionData.UserID
		data["session_is_admin"] = sessionData.IsAdmin
		data["session_auth_method"] = sessionData.AuthMethod
		data["session_user_groups"] = sessionData.UserGroups
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
	return r.RenderWithContext(context.Background(), w, name, data)
}

// Reload reloads all templates (useful in development)
func (r *Renderer) Reload() error {
	return r.loadTemplates()
}
