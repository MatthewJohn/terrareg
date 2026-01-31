package terrareg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider_source"
	provider_source_query "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider_source"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	terrareg_middleware "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// ProviderSourceAPIHandler handles provider source API requests
type ProviderSourceAPIHandler struct {
	getOrganizationsQuery   *provider_source_query.GetOrganizationsQuery
	getRepositoriesQuery    *provider_source_query.GetRepositoriesQuery
	refreshNamespaceCommand *provider_source.RefreshNamespaceCommand
	publishProviderCommand  *provider_source.PublishProviderCommand
	repositoryRepo          repository.RepositoryRepository
}

// NewProviderSourceAPIHandler creates a new ProviderSourceAPIHandler
func NewProviderSourceAPIHandler(
	getOrganizationsQuery *provider_source_query.GetOrganizationsQuery,
	getRepositoriesQuery *provider_source_query.GetRepositoriesQuery,
	refreshNamespaceCommand *provider_source.RefreshNamespaceCommand,
	publishProviderCommand *provider_source.PublishProviderCommand,
	repositoryRepo repository.RepositoryRepository,
) (*ProviderSourceAPIHandler, error) {
	if getOrganizationsQuery == nil {
		return nil, fmt.Errorf("getOrganizationsQuery is required")
	}
	if getRepositoriesQuery == nil {
		return nil, fmt.Errorf("getRepositoriesQuery is required")
	}
	if refreshNamespaceCommand == nil {
		return nil, fmt.Errorf("refreshNamespaceCommand is required")
	}
	if publishProviderCommand == nil {
		return nil, fmt.Errorf("publishProviderCommand is required")
	}
	if repositoryRepo == nil {
		return nil, fmt.Errorf("repositoryRepo is required")
	}

	return &ProviderSourceAPIHandler{
		getOrganizationsQuery:   getOrganizationsQuery,
		getRepositoriesQuery:    getRepositoriesQuery,
		refreshNamespaceCommand: refreshNamespaceCommand,
		publishProviderCommand:  publishProviderCommand,
		repositoryRepo:          repositoryRepo,
	}, nil
}

// HandleGetOrganizations handles GET /{provider_source}/organizations
// Python reference: github_organisations.py::GithubOrganisations
func (h *ProviderSourceAPIHandler) HandleGetOrganizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider source from path
	providerSource := chi.URLParam(r, "provider_source")

	// Get session ID from authentication context
	sessionID := getSessionIDFromContext(ctx)

	// Execute query
	orgs, err := h.getOrganizationsQuery.Execute(ctx, provider_source_query.GetOrganizationsRequest{
		ProviderSource: providerSource,
		SessionID:      sessionID,
	})
	if err != nil {
		// Check if provider source doesn't exist
		if errors.Is(err, shared.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "Provider source not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response DTO
	response := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = dto.OrganizationResponse{
			Name:                org.GetName(),
			Type:                org.GetType(),
			CanPublishProviders: org.CanPublish(),
		}
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleGetRepositories handles GET /{provider_source}/repositories
// Python reference: github_repositories.py::GithubRepositories
func (h *ProviderSourceAPIHandler) HandleGetRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider source from path
	providerSource := chi.URLParam(r, "provider_source")

	// Get session ID from authentication context
	sessionID := getSessionIDFromContext(ctx)

	// Check if user is admin
	isAdmin := getIsAdminFromContext(ctx)

	// Execute query
	repos, err := h.getRepositoriesQuery.Execute(ctx, provider_source_query.GetRepositoriesRequest{
		ProviderSource: providerSource,
		SessionID:      sessionID,
		IsAdmin:        isAdmin,
	})
	if err != nil {
		// Check if provider source doesn't exist
		if errors.Is(err, shared.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "Provider source not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response DTO
	response := make([]dto.RepositoryResponse, len(repos))
	for i, repo := range repos {
		response[i] = dto.RepositoryResponse{
			ID:          repo.GetID(),
			FullName:    repo.GetFullName(),
			OwnerLogin:  repo.GetOwnerLogin(),
			OwnerType:   repo.GetOwnerType(),
			Kind:        repo.GetKind(),
			PublishedID: repo.GetPublishedID(),
		}
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleRefreshNamespace handles POST /{provider_source}/refresh-namespace
// Python reference: github_refresh_namespace.py::GithubRefreshNamespace
func (h *ProviderSourceAPIHandler) HandleRefreshNamespace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider source from path
	providerSource := chi.URLParam(r, "provider_source")

	// Parse request body
	var req dto.RefreshNamespaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Execute command
	if err := h.refreshNamespaceCommand.Execute(ctx, provider_source.RefreshNamespaceRequest{
		ProviderSource: providerSource,
		Namespace:      req.Namespace,
	}); err != nil {
		// Check if provider source or namespace doesn't exist
		if errors.Is(err, shared.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "Not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return empty array on success (matching Python behavior)
	RespondJSON(w, http.StatusOK, []interface{}{})
}

// HandlePublishProvider handles POST /{provider_source}/repositories/{repo_id}/publish-provider
// Python reference: github_repository_publish_provider.py::GithubRepositoryPublishProvider
func (h *ProviderSourceAPIHandler) HandlePublishProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider source and repo_id from path
	providerSource := chi.URLParam(r, "provider_source")
	repoIDStr := chi.URLParam(r, "repo_id")

	// Parse repo_id
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	// Get auth context FIRST (before repository lookup)
	// Python reference: decorator checks permission BEFORE handler (lines 33-41)
	// If repository doesn't exist, namespace is None, which fails permission check for non-admins
	authCtx := terrareg_middleware.GetAuthContext(ctx)

	// Python reference: lines 72-74 - ensure user is GitHub authenticated OR admin
	isAdmin := authCtx.IsAdmin()
	hasGithubAuth := authCtx.IsAuthenticated() && authCtx.GetProviderType() == "github"

	// If not admin and not GitHub auth, return 403 immediately
	// This matches Python behavior where permission check fails before repository lookup
	if !isAdmin && !hasGithubAuth {
		RespondError(w, http.StatusForbidden, "GitHub authentication or admin access required")
		return
	}

	// Now look up repository (after admin/GitHub auth check passed)
	// Python reference: get_namespace_from_request_repository_id (lines 20-28)
	repository, err := h.repositoryRepo.FindByProviderSourceAndProviderID(ctx, providerSource, strconv.Itoa(repoID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to lookup repository")
		return
	}
	if repository == nil {
		// Python returns 404 only if admin/GitHub auth check passed (line 84-86)
		RespondError(w, http.StatusNotFound, "Repository not found")
		return
	}

	// Check namespace permission (additional check for non-admins)
	// Python reference: decorator with check_namespace_access (FULL permission)
	if !isAdmin {
		hasFullPermission := authCtx.CheckNamespaceAccess("FULL", repository.Owner)
		if !hasFullPermission {
			RespondError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	// Get category_id from form
	categoryIDStr := r.FormValue("category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid category_id")
		return
	}

	// Execute command with the derived namespace from the repository
	result, err := h.publishProviderCommand.Execute(ctx, provider_source.PublishProviderRequest{
		ProviderSource: providerSource,
		RepoID:         repoID,
		CategoryID:     categoryID,
		Namespace:      repository.Owner,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusOK, dto.PublishProviderResponse{
		Name:      result.Name,
		Namespace: result.Namespace,
	})
}

// Helper functions for getting data from context

func getSessionIDFromContext(ctx context.Context) string {
	authCtx := terrareg_middleware.GetAuthContext(ctx)
	if authCtx != nil && authCtx.IsAuthenticated() {
		if data := authCtx.GetProviderData(); data != nil {
			if sessionID, ok := data["session_id"].(string); ok {
				return sessionID
			}
		}
	}
	return ""
}

func getIsAdminFromContext(ctx context.Context) bool {
	authCtx := terrareg_middleware.GetAuthContext(ctx)
	if authCtx != nil {
		return authCtx.IsAdmin()
	}
	return false
}

func getNamespaceFromContext(ctx context.Context) string {
	// The middleware should have put the namespace in the context
	if ns, ok := ctx.Value("namespace").(string); ok {
		return ns
	}
	return ""
}
