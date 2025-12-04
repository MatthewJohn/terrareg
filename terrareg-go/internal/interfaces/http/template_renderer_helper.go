package http

import (
	"net/http"
)

// RenderTemplateWithContext is a helper function to render templates with session context
func (s *Server) RenderTemplateWithContext(w http.ResponseWriter, r *http.Request, templateName string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if data == nil {
		data = make(map[string]interface{})
	}

	// Add template name if not provided
	if _, exists := data["TEMPLATE_NAME"]; !exists {
		data["TEMPLATE_NAME"] = templateName
	}

	err := s.templateRenderer.RenderWithContext(r.Context(), w, templateName, data)
	if err != nil {
		s.logger.Error().Err(err).Str("template", templateName).Msg("Failed to render template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
