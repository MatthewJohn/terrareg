package template

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/csrf"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
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

	// Load base template first with custom functions
	baseTemplate := template.New("").Funcs(template.FuncMap{
		"csrf_token": func() string {
			// This is a placeholder - the actual CSRF token will be provided in the template data
			return ""
		},
		// Add any other custom template functions here if needed
	})

	// Parse base template (template.html)
	baseTemplate, err := baseTemplate.ParseFiles("templates/template.html")
	if err != nil {
		return err
	}

	// Load all other templates
	pattern := filepath.Join("templates", "*.html")
	templates, err := baseTemplate.ParseGlob(pattern)
	if err != nil {
		return err
	}

	r.templates = templates
	return nil
}

// Render renders a template with the given data
func (r *Renderer) Render(w io.Writer, name string, data map[string]interface{}) error {
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
	data["enable_security_scanning"] = false // TODO: Add to config
	data["terrareg_logo_url"] = "/static/images/logo.png" // TODO: Add to config
	data["theme_path"] = "/static/css/terrareg.css" // TODO: Add to config
	data["SITE_WARNING"] = "" // TODO: Add to config

	return r.templates.ExecuteTemplate(w, name, data)
}

// Reload reloads all templates (useful in development)
func (r *Renderer) Reload() error {
	return r.loadTemplates()
}
