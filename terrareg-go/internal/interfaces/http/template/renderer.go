package template

import (
	"html/template"
	"io"
	"sync"

	"github.com/terrareg/terrareg/internal/config"
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

// loadTemplates loads all HTML templates
func (r *Renderer) loadTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// IMPORTANT: The templates contain Jinja2 syntax which is incompatible with Go templates.
	// Since the frontend is entirely JavaScript-based and calls API endpoints for data,
	// we don't actually need server-side template rendering.
	// For now, create a minimal template set - templates will be served as static files.

	// Create a dummy template to satisfy the interface
	templates := template.New("")
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

	return r.templates.ExecuteTemplate(w, name, data)
}

// Reload reloads all templates (useful in development)
func (r *Renderer) Reload() error {
	return r.loadTemplates()
}
