package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
)

// ModuleHandler handles module-related requests
type ModuleHandler struct {
	listModulesQuery                *moduleQuery.ListModulesQuery
	searchModulesQuery              *moduleQuery.SearchModulesQuery
	getModuleProviderQuery          *moduleQuery.GetModuleProviderQuery
	listModuleProvidersQuery        *moduleQuery.ListModuleProvidersQuery
	getModuleVersionQuery           *moduleQuery.GetModuleVersionQuery
	getModuleDownloadQuery          *moduleQuery.GetModuleDownloadQuery
	getModuleProviderSettingsQuery  *moduleQuery.GetModuleProviderSettingsQuery
	getSubmodulesQuery              *moduleQuery.GetSubmodulesQuery
	getExamplesQuery                *moduleQuery.GetExamplesQuery
	getIntegrationsQuery             *moduleQuery.GetIntegrationsQuery
	createModuleProviderCmd         *moduleCmd.CreateModuleProviderCommand
	publishModuleVersionCmd         *moduleCmd.PublishModuleVersionCommand
	updateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand
	deleteModuleProviderCmd         *moduleCmd.DeleteModuleProviderCommand
	uploadModuleVersionCmd          *moduleCmd.UploadModuleVersionCommand
	importModuleVersionCmd          *moduleCmd.ImportModuleVersionCommand
	getModuleVersionFileCmd         *moduleCmd.GetModuleVersionFileQuery
	deleteModuleVersionCmd          *moduleCmd.DeleteModuleVersionCommand
	generateModuleSourceCmd         *moduleCmd.GenerateModuleSourceCommand
	getVariableTemplateQuery         *moduleCmd.GetVariableTemplateQuery
	presenter                       *presenter.ModulePresenter
	versionPresenter                *presenter.ModuleVersionPresenter
	domainConfig                    *model.DomainConfig
	analyticsRepo                   analyticsCmd.AnalyticsRepository
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(
	listModulesQuery *moduleQuery.ListModulesQuery,
	searchModulesQuery *moduleQuery.SearchModulesQuery,
	getModuleProviderQuery *moduleQuery.GetModuleProviderQuery,
	listModuleProvidersQuery *moduleQuery.ListModuleProvidersQuery,
	getModuleVersionQuery *moduleQuery.GetModuleVersionQuery,
	getModuleDownloadQuery *moduleQuery.GetModuleDownloadQuery,
	getModuleProviderSettingsQuery *moduleQuery.GetModuleProviderSettingsQuery,
	getSubmodulesQuery *moduleQuery.GetSubmodulesQuery,
	getExamplesQuery *moduleQuery.GetExamplesQuery,
	getIntegrationsQuery *moduleQuery.GetIntegrationsQuery,
	createModuleProviderCmd *moduleCmd.CreateModuleProviderCommand,
	publishModuleVersionCmd *moduleCmd.PublishModuleVersionCommand,
	updateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand,
	deleteModuleProviderCmd *moduleCmd.DeleteModuleProviderCommand,
	uploadModuleVersionCmd *moduleCmd.UploadModuleVersionCommand,
	importModuleVersionCmd *moduleCmd.ImportModuleVersionCommand,
	getModuleVersionFileCmd *moduleCmd.GetModuleVersionFileQuery,
	deleteModuleVersionCmd *moduleCmd.DeleteModuleVersionCommand,
	generateModuleSourceCmd *moduleCmd.GenerateModuleSourceCommand,
	getVariableTemplateQuery *moduleCmd.GetVariableTemplateQuery,
	domainConfig *model.DomainConfig,
	namespaceService *moduleService.NamespaceService,
	analyticsRepo analyticsCmd.AnalyticsRepository,
) *ModuleHandler {
	return &ModuleHandler{
		listModulesQuery:                listModulesQuery,
		searchModulesQuery:              searchModulesQuery,
		getModuleProviderQuery:          getModuleProviderQuery,
		listModuleProvidersQuery:        listModuleProvidersQuery,
		getModuleVersionQuery:           getModuleVersionQuery,
		getModuleDownloadQuery:          getModuleDownloadQuery,
		getModuleProviderSettingsQuery:  getModuleProviderSettingsQuery,
		getSubmodulesQuery:              getSubmodulesQuery,
		getExamplesQuery:                getExamplesQuery,
		getIntegrationsQuery:            getIntegrationsQuery,
		createModuleProviderCmd:         createModuleProviderCmd,
		publishModuleVersionCmd:         publishModuleVersionCmd,
		updateModuleProviderSettingsCmd: updateModuleProviderSettingsCmd,
		deleteModuleProviderCmd:         deleteModuleProviderCmd,
		uploadModuleVersionCmd:          uploadModuleVersionCmd,
		importModuleVersionCmd:          importModuleVersionCmd,
		getModuleVersionFileCmd:         getModuleVersionFileCmd,
		deleteModuleVersionCmd:          deleteModuleVersionCmd,
		generateModuleSourceCmd:        generateModuleSourceCmd,
		getVariableTemplateQuery:        getVariableTemplateQuery,
		presenter:                       presenter.NewModulePresenter(analyticsRepo),
		versionPresenter:                presenter.NewModuleVersionPresenter(namespaceService, analyticsRepo),
		domainConfig:                    domainConfig,
		analyticsRepo:                   analyticsRepo,
	}
}

// HandleModuleList handles GET /v1/modules
func (h *ModuleHandler) HandleModuleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute query
	modules, err := h.listModulesQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToListDTO(ctx, modules)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleSearch handles GET /v1/modules/search
func (h *ModuleHandler) HandleModuleSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query().Get("q")
	namespace := r.URL.Query().Get("namespace")
	provider := r.URL.Query().Get("provider")

	// Note: namespace is handled directly in Namespaces array below

	var providerPtr *string
	if provider != "" {
		providerPtr = &provider
	}

	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Execute search
	params := moduleQuery.SearchParams{
		Query:      query,
		Namespaces: []string{namespace},
		Provider:   providerPtr,
		Limit:      limit,
		Offset:     offset,
	}

	result, err := h.searchModulesQuery.Execute(ctx, params)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToSearchDTO(ctx, result.Modules, result.TotalCount, limit, offset)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleNamespaceModules handles GET /v1/modules/{namespace}
func (h *ModuleHandler) HandleNamespaceModules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = chi.URLParam(r, "namespace") // TODO: Filter by namespace when query supports it

	// Execute query
	modules, err := h.listModulesQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToListDTO(ctx, modules)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderCreate handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/create
func (h *ModuleHandler) HandleModuleProviderCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Create command request
	cmdReq := moduleCmd.CreateModuleProviderRequest{
		Namespace: namespace,
		Module:    name,
		Provider:  provider,
	}

	// Execute command
	moduleProvider, err := h.createModuleProviderCmd.Execute(ctx, cmdReq)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToDTO(ctx, moduleProvider)

	// Send response
	RespondJSON(w, http.StatusCreated, response)
}

// HandleModuleDetails handles GET /v1/modules/{namespace}/{name}
func (h *ModuleHandler) HandleModuleDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	// Execute query to list all providers for this module
	providers, err := h.listModuleProvidersQuery.Execute(ctx, namespace, name)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToListDTO(ctx, providers)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderDetails handles GET /v1/modules/{namespace}/{name}/{provider}
func (h *ModuleHandler) HandleModuleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Execute query
	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToDTO(ctx, moduleProvider)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleVersions handles GET /v1/modules/{namespace}/{name}/{provider}/versions
func (h *ModuleHandler) HandleModuleVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Get the module provider first
	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Get versions from the module provider
	versions := moduleProvider.GetAllVersions()

	// Convert to version DTOs
	versionDTOs := make([]map[string]interface{}, len(versions))
	for i, version := range versions {
		versionDTOs[i] = map[string]interface{}{
			"version": version.Version().String(),
		}
	}

	// Build response matching Terraform Registry API format
	response := map[string]interface{}{
		"modules": []map[string]interface{}{
			{
				"versions": versionDTOs,
			},
		},
	}

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleVersionPublish handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/publish
func (h *ModuleHandler) HandleModuleVersionPublish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Parse request body for optional fields
	var req moduledto.ModuleVersionPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body or invalid JSON, use defaults
		req = moduledto.ModuleVersionPublishRequest{
			Version: version, // Use version from URL
			Beta:    false,
		}
	} else {
		// Override version with URL parameter
		req.Version = version
	}

	// Create command request
	cmdReq := moduleCmd.PublishModuleVersionRequest{
		Namespace:   namespace,
		Module:      name,
		Provider:    provider,
		Version:     req.Version,
		Beta:        req.Beta,
		Description: req.Description,
		Owner:       req.Owner,
	}

	// Execute command
	moduleVersion, err := h.publishModuleVersionCmd.Execute(ctx, cmdReq)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.versionPresenter.ToDTO(ctx, moduleVersion, namespace, name, provider)

	// Send response
	RespondJSON(w, http.StatusCreated, response)
}

// HandleModuleVersionDetails handles GET /v1/modules/{namespace}/{name}/{provider}/{version}
func (h *ModuleHandler) HandleModuleVersionDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Execute query
	moduleVersion, err := h.getModuleVersionQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Convert to DTO - use the same detailed format as module provider endpoint
	response := h.versionPresenter.ToTerraregProviderDetailsDTO(ctx, moduleVersion, namespace, name, provider, r.Host)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleDownload handles GET /v1/modules/{namespace}/{name}/{provider}/download
// and GET /v1/modules/{namespace}/{name}/{provider}/{version}/download
// Returns download location in Terraform Registry API format
func (h *ModuleHandler) HandleModuleDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version") // May be empty for latest

	// Check if module hosting is disallowed
	// TODO: Implement full logic for ALLOW/ENFORCE modes with git URL handling
	if h.domainConfig.AllowModuleHosting == model.ModuleHostingModeDisallow {
		RespondError(w, http.StatusInternalServerError, "Module hosting is disabled")
		return
	}

	// Execute query to get download info
	downloadInfo, err := h.getModuleDownloadQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build download URL
	// In Terraform Registry API, the X-Terraform-Get header contains the download location
	// The response body is empty, but we return JSON with version info
	downloadURL := fmt.Sprintf("/v1/modules/%s/%s/%s/%s/download",
		namespace, name, provider, downloadInfo.Version.Version().String())

	// Set the X-Terraform-Get header (Terraform will use this to download)
	w.Header().Set("X-Terraform-Get", downloadURL)

	// Return version information in response body
	response := map[string]interface{}{
		"version": downloadInfo.Version.Version().String(),
	}

	RespondJSON(w, http.StatusNoContent, response)
}

// HandleModuleProviderSettingsGet handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/settings
func (h *ModuleHandler) HandleModuleProviderSettingsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Execute query
	settings, err := h.getModuleProviderSettingsQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build response
	response := moduledto.ModuleProviderSettingsResponse{
		Namespace:             namespace,
		Module:                name,
		Provider:              provider,
		GitProviderID:         settings.GitProviderID,
		RepoBaseURLTemplate:   settings.RepoBaseURLTemplate,
		RepoCloneURLTemplate:  settings.RepoCloneURLTemplate,
		RepoBrowseURLTemplate: settings.RepoBrowseURLTemplate,
		GitTagFormat:          settings.GitTagFormat,
		GitPath:               settings.GitPath,
		ArchiveGitPath:        settings.ArchiveGitPath,
		Verified:              settings.Verified,
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderSettingsUpdate handles PUT /v1/terrareg/modules/{namespace}/{name}/{provider}/settings
func (h *ModuleHandler) HandleModuleProviderSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Parse request body
	var req moduledto.ModuleProviderSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	// Execute command
	cmdReq := moduleCmd.UpdateModuleProviderSettingsRequest{
		Namespace:             namespace,
		Module:                name,
		Provider:              provider,
		GitProviderID:         req.GitProviderID,
		RepoBaseURLTemplate:   req.RepoBaseURLTemplate,
		RepoCloneURLTemplate:  req.RepoCloneURLTemplate,
		RepoBrowseURLTemplate: req.RepoBrowseURLTemplate,
		GitTagFormat:          req.GitTagFormat,
		GitPath:               req.GitPath,
		ArchiveGitPath:        req.ArchiveGitPath,
		Verified:              req.Verified,
	}

	if err := h.updateModuleProviderSettingsCmd.Execute(ctx, cmdReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return updated settings
	settings, err := h.getModuleProviderSettingsQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := moduledto.ModuleProviderSettingsResponse{
		Namespace:             namespace,
		Module:                name,
		Provider:              provider,
		GitProviderID:         settings.GitProviderID,
		RepoBaseURLTemplate:   settings.RepoBaseURLTemplate,
		RepoCloneURLTemplate:  settings.RepoCloneURLTemplate,
		RepoBrowseURLTemplate: settings.RepoBrowseURLTemplate,
		GitTagFormat:          settings.GitTagFormat,
		GitPath:               settings.GitPath,
		ArchiveGitPath:        settings.ArchiveGitPath,
		Verified:              settings.Verified,
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderDelete handles DELETE /v1/terrareg/modules/{namespace}/{name}/{provider}
func (h *ModuleHandler) HandleModuleProviderDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Execute command
	cmdReq := moduleCmd.DeleteModuleProviderRequest{
		Namespace: namespace,
		Module:    name,
		Provider:  provider,
	}

	if err := h.deleteModuleProviderCmd.Execute(ctx, cmdReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 204 No Content on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// HandleModuleVersionUpload handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/upload
func (h *ModuleHandler) HandleModuleVersionUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	if namespace == "" || name == "" || provider == "" || version == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Check if module hosting is disallowed
	if h.domainConfig.AllowModuleHosting == model.ModuleHostingModeDisallow {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Module upload is disabled."))
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Failed to parse multipart form"))
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("No file provided in 'file' field"))
		return
	}
	defer file.Close()

	// Execute upload command
	uploadReq := moduleCmd.UploadModuleVersionRequest{
		Namespace:  namespace,
		Module:     name,
		Provider:   provider,
		Version:    version,
		Source:     file,
		SourceSize: header.Size,
	}

	if err := h.uploadModuleVersionCmd.Execute(ctx, uploadReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Module version uploaded successfully",
	})
}

// HandleModuleVersionImport handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/import
func (h *ModuleHandler) HandleModuleVersionImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	if namespace == "" || name == "" || provider == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Parse JSON request body
	var reqBody struct {
		Version *string `json:"version"`
		GitTag  *string `json:"git_tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Invalid request body"))
		return
	}

	// Execute import command
	importReq := moduleCmd.ImportModuleVersionRequest{
		Namespace: namespace,
		Module:    name,
		Provider:  provider,
		Version:   reqBody.Version,
		GitTag:    reqBody.GitTag,
	}

	if err := h.importModuleVersionCmd.Execute(ctx, importReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "Success",
	})
}

// HandleGetSubmodules handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/submodules
func (h *ModuleHandler) HandleGetSubmodules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	if namespace == "" || name == "" || provider == "" || version == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Execute query
	submodules, err := h.getSubmodulesQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return submodules
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"submodules": submodules,
	})
}

// HandleGetExamples handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/examples
func (h *ModuleHandler) HandleGetExamples(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	if namespace == "" || name == "" || provider == "" || version == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Execute query
	examples, err := h.getExamplesQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return examples
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"examples": examples,
	})
}

// HandleTerraregModuleProviderDetails handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}
// Returns full terrareg API details for the latest version (published or unpublished)
func (h *ModuleHandler) HandleTerraregModuleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Extract request domain for usage example generation
	requestDomain := r.Host
	if requestDomain == "" {
		requestDomain = "localhost" // fallback for testing
	}

	// Get the module provider
	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Get all versions to find the most recent one
	allVersions := moduleProvider.GetAllVersions()
	if len(allVersions) == 0 {
		RespondError(w, http.StatusNotFound, "No versions found")
		return
	}

	// Find the most recent version (regardless of publish status)
	var latestVersion *moduleModel.ModuleVersion
	for _, version := range allVersions {
		if latestVersion == nil || version.Version().GreaterThan(latestVersion.Version()) {
			latestVersion = version
		}
	}

	// Convert to Terrareg Provider Details DTO with request domain and analytics token
	response := h.versionPresenter.ToTerraregProviderDetailsDTO(ctx, latestVersion, namespace, name, provider, requestDomain)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleTerraregModuleProviderVersions handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/versions
func (h *ModuleHandler) HandleTerraregModuleProviderVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Parse query parameters
	includeBeta := r.URL.Query().Get("include-beta") == "true"
	includeUnpublished := r.URL.Query().Get("include-unpublished") == "true"

	// Get the module provider
	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Get versions from the module provider
	allVersions := moduleProvider.GetAllVersions()
	if len(allVersions) == 0 {
		// Return empty array instead of error
		RespondJSON(w, http.StatusOK, []interface{}{})
		return
	}

	// Filter and build version response
	var versions []map[string]interface{}
	for _, version := range allVersions {
		// Skip versions based on query parameters
		if !includeBeta && version.IsBeta() {
			continue
		}
		if !includeUnpublished && !version.IsPublished() {
			continue
		}

		versions = append(versions, map[string]interface{}{
			"version":   version.Version().String(),
			"published": version.IsPublished(),
			"beta":      version.IsBeta(),
		})
	}

	// Send response
	RespondJSON(w, http.StatusOK, versions)
}

// HandleModuleFile handles GET /modules/{namespace}/{name}/{provider}/{version}/files/{path}
func (h *ModuleHandler) HandleModuleFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	path := chi.URLParam(r, "path")

	// Create request
	req := &moduleCmd.GetModuleVersionFileRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
		Path:      path,
	}

	// Execute query
	resp, err := h.getModuleVersionFileCmd.Execute(ctx, req)
	if err != nil {
		if err.Error() == "module version file not found" {
			RespondError(w, http.StatusNotFound, "Module version file not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", resp.ContentType)

	// Write the file content
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.Content))
}

// HandleModuleVersionCreate handles POST /modules/{namespace}/{name}/{provider}/{version}/import
// This is the deprecated endpoint that requires version in URL
func (h *ModuleHandler) HandleModuleVersionCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Parse request body
	var importReq struct {
		// The deprecated endpoint doesn't accept a body, but we'll parse it for consistency
		// The version is already known from the URL
	}

	if err := json.NewDecoder(r.Body).Decode(&importReq); err != nil {
		// For this deprecated endpoint, we can ignore body parsing errors
		// since the version comes from the URL
	}

	// Create import request - version comes from URL for this endpoint
	request := moduleCmd.ImportModuleVersionRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   &version, // Use version from URL
		GitTag:    nil,       // No git tag, derive from version
	}

	// Execute import command
	err := h.importModuleVersionCmd.Execute(ctx, request)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Module version import started successfully",
	})
}

// HandleModuleVersionDelete handles DELETE /modules/{namespace}/{name}/{provider}/{version}/delete
func (h *ModuleHandler) HandleModuleVersionDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Create delete request
	req := moduleCmd.DeleteModuleVersionRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
	}

	// Execute delete command
	if err := h.deleteModuleVersionCmd.Execute(ctx, req); err != nil {
		if err.Error() == "module provider not found" {
			RespondError(w, http.StatusNotFound, "Module provider not found")
			return
		}
		if err.Error() == "module version not found" {
			RespondError(w, http.StatusNotFound, "Module version not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Module version deleted successfully",
	})
}

// HandleModuleVersionReadmeHTML handles GET /modules/{namespace}/{name}/{provider}/{version}/readme_html
func (h *ModuleHandler) HandleModuleVersionReadmeHTML(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Create request to get README file (typically named README.md)
	req := &moduleCmd.GetModuleVersionFileRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
		Path:      "README.md", // Standard README filename
	}

	// Execute query to get README file
	resp, err := h.getModuleVersionFileCmd.Execute(ctx, req)
	if err != nil {
		// Try alternative README filenames
		altReadmeFiles := []string{"README", "readme.md", "Readme.md"}

		for _, readmeFile := range altReadmeFiles {
			req.Path = readmeFile
			resp, err = h.getModuleVersionFileCmd.Execute(ctx, req)
			if err == nil {
				break
			}
		}

		if err != nil {
			if err.Error() == "module version file not found" ||
			   err.Error() == "file not found: module version file not found" {
				// Return HTML error page for missing README
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<div class="alert alert-warning">No README found for this module version</div>`))
				return
			}
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the processed HTML content
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.ContentHTML))
}

// HandleModuleVersionSourceDownload handles GET /modules/{namespace}/{name}/{provider}/{version}/source.zip
func (h *ModuleHandler) HandleModuleVersionSourceDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Create request
	req := &moduleCmd.GenerateModuleSourceRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
	}

	// Execute command
	resp, err := h.generateModuleSourceCmd.Execute(ctx, req)
	if err != nil {
		if err.Error() == "missing required parameters" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "Missing required path parameters"}`))
			return
		}
		if err.Error() == "no files found for module version" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "No source files found for module version"}`))
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", resp.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", resp.ContentSize))

	// Write the ZIP content
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Content)
}

// HandleModuleVersionVariableTemplate handles GET /modules/{namespace}/{name}/{provider}/{version}/variable_template
func (h *ModuleHandler) HandleModuleVersionVariableTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Extract query parameter for output format
	outputFormat := r.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "md" // Default to markdown
	}

	// Create request
	req := &moduleCmd.GetVariableTemplateRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
		Output:    outputFormat,
	}

	// Execute query
	resp, err := h.getVariableTemplateQuery.Execute(ctx, req)
	if err != nil {
		if err.Error() == "missing required parameters" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "Missing required path parameters"}`))
			return
		}
		if err.Error() == "module version not found" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Module version not found"}`))
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Send response
	RespondJSON(w, http.StatusOK, resp)
}

// HandleGetIntegrations handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/integrations
func (h *ModuleHandler) HandleGetIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	if namespace == "" || name == "" || provider == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Execute query
	integrations, err := h.getIntegrationsQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return integrations
	RespondJSON(w, http.StatusOK, integrations)
}
