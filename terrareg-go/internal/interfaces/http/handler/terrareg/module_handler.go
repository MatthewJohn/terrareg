package terrareg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
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
	getReadmeHTMLQuery              *moduleQuery.GetReadmeHTMLQuery
	getSubmodulesQuery              *moduleQuery.GetSubmodulesQuery
	getExamplesQuery                *moduleQuery.GetExamplesQuery
	getIntegrationsQuery            *moduleQuery.GetIntegrationsQuery
	createModuleProviderCmd         *moduleCmd.CreateModuleProviderCommand
	publishModuleVersionCmd         *moduleCmd.PublishModuleVersionCommand
	updateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand
	deleteModuleProviderCmd         *moduleCmd.DeleteModuleProviderCommand
	processModuleCmd                *moduleCmd.ProcessModuleCommand
	getModuleVersionFileCmd         *moduleCmd.GetModuleVersionFileQuery
	deleteModuleVersionCmd          *moduleCmd.DeleteModuleVersionCommand
	generateModuleSourceCmd         *moduleCmd.GenerateModuleSourceCommand
	getVariableTemplateQuery        *moduleCmd.GetVariableTemplateQuery
	createModuleProviderRedirectCmd *moduleCmd.CreateModuleProviderRedirectCommand
	deleteModuleProviderRedirectCmd *moduleCmd.DeleteModuleProviderRedirectCommand
	getModuleProviderRedirectsQuery *moduleQuery.GetModuleProviderRedirectsQuery
	recordModuleDownloadCmd         *analyticsCmd.RecordModuleDownloadCommand
	presenter                       *presenter.ModulePresenter
	versionPresenter                *presenter.ModuleVersionPresenter
	domainConfig                    *model.DomainConfig
	analyticsRepo                   analyticsCmd.AnalyticsRepository
}

// NewModuleHandler creates a new module handler
// Returns an error if any required dependency is nil
func NewModuleHandler(
	listModulesQuery *moduleQuery.ListModulesQuery,
	searchModulesQuery *moduleQuery.SearchModulesQuery,
	getModuleProviderQuery *moduleQuery.GetModuleProviderQuery,
	listModuleProvidersQuery *moduleQuery.ListModuleProvidersQuery,
	getModuleVersionQuery *moduleQuery.GetModuleVersionQuery,
	getModuleDownloadQuery *moduleQuery.GetModuleDownloadQuery,
	getModuleProviderSettingsQuery *moduleQuery.GetModuleProviderSettingsQuery,
	getReadmeHTMLQuery *moduleQuery.GetReadmeHTMLQuery,
	getSubmodulesQuery *moduleQuery.GetSubmodulesQuery,
	getExamplesQuery *moduleQuery.GetExamplesQuery,
	getIntegrationsQuery *moduleQuery.GetIntegrationsQuery,
	createModuleProviderCmd *moduleCmd.CreateModuleProviderCommand,
	publishModuleVersionCmd *moduleCmd.PublishModuleVersionCommand,
	updateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand,
	deleteModuleProviderCmd *moduleCmd.DeleteModuleProviderCommand,
	processModuleCmd *moduleCmd.ProcessModuleCommand,
	getModuleVersionFileCmd *moduleCmd.GetModuleVersionFileQuery,
	deleteModuleVersionCmd *moduleCmd.DeleteModuleVersionCommand,
	generateModuleSourceCmd *moduleCmd.GenerateModuleSourceCommand,
	getVariableTemplateQuery *moduleCmd.GetVariableTemplateQuery,
	createModuleProviderRedirectCmd *moduleCmd.CreateModuleProviderRedirectCommand,
	deleteModuleProviderRedirectCmd *moduleCmd.DeleteModuleProviderRedirectCommand,
	getModuleProviderRedirectsQuery *moduleQuery.GetModuleProviderRedirectsQuery,
	recordModuleDownloadCmd *analyticsCmd.RecordModuleDownloadCommand,
	domainConfig *model.DomainConfig,
	namespaceService *moduleService.NamespaceService,
	analyticsRepo analyticsCmd.AnalyticsRepository,
	urlService *service.URLService,
) (*ModuleHandler, error) {
	// Validate all pointer dependencies
	if listModulesQuery == nil {
		return nil, fmt.Errorf("listModulesQuery cannot be nil")
	}
	if searchModulesQuery == nil {
		return nil, fmt.Errorf("searchModulesQuery cannot be nil")
	}
	if getModuleProviderQuery == nil {
		return nil, fmt.Errorf("getModuleProviderQuery cannot be nil")
	}
	if listModuleProvidersQuery == nil {
		return nil, fmt.Errorf("listModuleProvidersQuery cannot be nil")
	}
	if getModuleVersionQuery == nil {
		return nil, fmt.Errorf("getModuleVersionQuery cannot be nil")
	}
	if getModuleDownloadQuery == nil {
		return nil, fmt.Errorf("getModuleDownloadQuery cannot be nil")
	}
	if getModuleProviderSettingsQuery == nil {
		return nil, fmt.Errorf("getModuleProviderSettingsQuery cannot be nil")
	}
	if getReadmeHTMLQuery == nil {
		return nil, fmt.Errorf("getReadmeHTMLQuery cannot be nil")
	}
	if getSubmodulesQuery == nil {
		return nil, fmt.Errorf("getSubmodulesQuery cannot be nil")
	}
	if getExamplesQuery == nil {
		return nil, fmt.Errorf("getExamplesQuery cannot be nil")
	}
	if getIntegrationsQuery == nil {
		return nil, fmt.Errorf("getIntegrationsQuery cannot be nil")
	}
	if createModuleProviderCmd == nil {
		return nil, fmt.Errorf("createModuleProviderCmd cannot be nil")
	}
	if publishModuleVersionCmd == nil {
		return nil, fmt.Errorf("publishModuleVersionCmd cannot be nil")
	}
	if updateModuleProviderSettingsCmd == nil {
		return nil, fmt.Errorf("updateModuleProviderSettingsCmd cannot be nil")
	}
	if deleteModuleProviderCmd == nil {
		return nil, fmt.Errorf("deleteModuleProviderCmd cannot be nil")
	}
	if processModuleCmd == nil {
		return nil, fmt.Errorf("processModuleCmd cannot be nil")
	}
	if getModuleVersionFileCmd == nil {
		return nil, fmt.Errorf("getModuleVersionFileCmd cannot be nil")
	}
	if deleteModuleVersionCmd == nil {
		return nil, fmt.Errorf("deleteModuleVersionCmd cannot be nil")
	}
	if generateModuleSourceCmd == nil {
		return nil, fmt.Errorf("generateModuleSourceCmd cannot be nil")
	}
	if getVariableTemplateQuery == nil {
		return nil, fmt.Errorf("getVariableTemplateQuery cannot be nil")
	}
	if createModuleProviderRedirectCmd == nil {
		return nil, fmt.Errorf("createModuleProviderRedirectCmd cannot be nil")
	}
	if deleteModuleProviderRedirectCmd == nil {
		return nil, fmt.Errorf("deleteModuleProviderRedirectCmd cannot be nil")
	}
	if getModuleProviderRedirectsQuery == nil {
		return nil, fmt.Errorf("getModuleProviderRedirectsQuery cannot be nil")
	}
	if recordModuleDownloadCmd == nil {
		return nil, fmt.Errorf("recordModuleDownloadCmd cannot be nil")
	}
	if domainConfig == nil {
		return nil, fmt.Errorf("domainConfig cannot be nil")
	}
	if namespaceService == nil {
		return nil, fmt.Errorf("namespaceService cannot be nil")
	}

	return &ModuleHandler{
		listModulesQuery:                listModulesQuery,
		searchModulesQuery:              searchModulesQuery,
		getModuleProviderQuery:          getModuleProviderQuery,
		listModuleProvidersQuery:        listModuleProvidersQuery,
		getModuleVersionQuery:           getModuleVersionQuery,
		getModuleDownloadQuery:          getModuleDownloadQuery,
		getModuleProviderSettingsQuery:  getModuleProviderSettingsQuery,
		getReadmeHTMLQuery:              getReadmeHTMLQuery,
		getSubmodulesQuery:              getSubmodulesQuery,
		getExamplesQuery:                getExamplesQuery,
		getIntegrationsQuery:            getIntegrationsQuery,
		createModuleProviderCmd:         createModuleProviderCmd,
		publishModuleVersionCmd:         publishModuleVersionCmd,
		updateModuleProviderSettingsCmd: updateModuleProviderSettingsCmd,
		deleteModuleProviderCmd:         deleteModuleProviderCmd,
		processModuleCmd:                processModuleCmd,
		getModuleVersionFileCmd:         getModuleVersionFileCmd,
		deleteModuleVersionCmd:          deleteModuleVersionCmd,
		generateModuleSourceCmd:         generateModuleSourceCmd,
		getVariableTemplateQuery:        getVariableTemplateQuery,
		createModuleProviderRedirectCmd: createModuleProviderRedirectCmd,
		deleteModuleProviderRedirectCmd: deleteModuleProviderRedirectCmd,
		getModuleProviderRedirectsQuery: getModuleProviderRedirectsQuery,
		recordModuleDownloadCmd:         recordModuleDownloadCmd,
		presenter:                       presenter.NewModulePresenter(analyticsRepo),
		versionPresenter:                presenter.NewModuleVersionPresenter(namespaceService, analyticsRepo, urlService),
		domainConfig:                    domainConfig,
		analyticsRepo:                   analyticsRepo,
	}, nil
}

// NewModuleReadHandlerForTesting creates a ModuleHandler for testing read operations.
// Only the dependencies needed for read operations are required.
// All command and write-operation dependencies are left nil.
//
// The versionPresenter parameter is optional - pass nil for basic tests that
// don't need Terrareg-specific endpoints (like HandleTerraregModuleProviderDetails).
//
// This is intended for testing purposes where the full 29-parameter constructor
// would be impractical. Read operations (HandleModuleList, HandleNamespaceModules,
// HandleModuleDetails, HandleModuleProviderDetails, HandleModuleSearch) only use
// the queries and presenter specified here.
func NewModuleReadHandlerForTesting(
	listModulesQuery *moduleQuery.ListModulesQuery,
	searchModulesQuery *moduleQuery.SearchModulesQuery,
	getModuleProviderQuery *moduleQuery.GetModuleProviderQuery,
	listModuleProvidersQuery *moduleQuery.ListModuleProvidersQuery,
	analyticsRepo analyticsCmd.AnalyticsRepository,
	versionPresenter *presenter.ModuleVersionPresenter,
) *ModuleHandler {
	return &ModuleHandler{
		listModulesQuery:         listModulesQuery,
		searchModulesQuery:       searchModulesQuery,
		getModuleProviderQuery:   getModuleProviderQuery,
		listModuleProvidersQuery: listModuleProvidersQuery,
		presenter:                presenter.NewModulePresenter(analyticsRepo),
		versionPresenter:         versionPresenter,
		analyticsRepo:            analyticsRepo,
		// All other fields remain nil - not used by read operations
	}
}

// NewModuleVersionDetailsHandlerForTesting creates a handler with getModuleVersionQuery for testing version details endpoints
func NewModuleVersionDetailsHandlerForTesting(
	getModuleVersionQuery *moduleQuery.GetModuleVersionQuery,
	getModuleProviderQuery *moduleQuery.GetModuleProviderQuery,
	analyticsRepo analyticsCmd.AnalyticsRepository,
	versionPresenter *presenter.ModuleVersionPresenter,
) *ModuleHandler {
	return &ModuleHandler{
		getModuleVersionQuery:  getModuleVersionQuery,
		getModuleProviderQuery: getModuleProviderQuery,
		presenter:              presenter.NewModulePresenter(analyticsRepo),
		versionPresenter:       versionPresenter,
		analyticsRepo:          analyticsRepo,
		// All other fields remain nil - not used by version details operations
	}
}

// HandleModuleList handles GET /v1/modules
// Python reference: /app/test/unit/terrareg/server/test_api_module_list.py
func (h *ModuleHandler) HandleModuleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters matching Python's test_api_module_list.py
	queryParams := r.URL.Query()

	// Parse providers (support multiple values, like Python)
	var providers []types.ModuleProviderName
	if p := queryParams["provider"]; len(p) > 0 {
		for _, prov := range p {
			if prov != "" {
				providers = append(providers, types.ModuleProviderName(prov))
			}
		}
	}

	// Parse verified parameter
	verified := parseBoolPtr(queryParams.Get("verified"))

	// Parse pagination parameters
	offset := parseInt(queryParams.Get("offset"), 0)
	limit := parseInt(queryParams.Get("limit"), 10)

	// Build input for query
	input := moduleQuery.ListModulesInput{
		Offset:       offset,
		Limit:        limit,
		Providers:    providers,
		Verified:     verified,
		IncludeCount: true, // Always include count to determine if more results exist
	}

	// Execute query
	modules, totalCount, err := h.listModulesQuery.Execute(ctx, input)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToListDTOWithMeta(ctx, modules, offset, limit, totalCount)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleSearch handles GET /v1/modules/search
func (h *ModuleHandler) HandleModuleSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query().Get("q")

	// Parse namespaces (support multiple values, like Python)
	var namespaces []types.NamespaceName
	if ns := r.URL.Query()["namespace"]; len(ns) > 0 {
		// Only include non-empty namespace values (matching Python behavior)
		for _, n := range ns {
			if n != "" {
				namespaces = append(namespaces, types.NamespaceName(n))
			}
		}
	}

	// Parse providers (support multiple values, like Python)
	var providers []types.ModuleProviderName
	if p := r.URL.Query()["provider"]; len(p) > 0 {
		// Only include non-empty provider values
		for _, prov := range p {
			if prov != "" {
				providers = append(providers, types.ModuleProviderName(prov))
			}
		}
	}

	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Parse verified filter
	var verified *bool
	if v := r.URL.Query().Get("verified"); v != "" {
		if val, err := strconv.ParseBool(v); err == nil {
			verified = &val
		}
	}

	// Parse trusted_namespaces filter
	var trustedNamespaces *bool
	if tn := r.URL.Query().Get("trusted_namespaces"); tn != "" {
		if val, err := strconv.ParseBool(tn); err == nil {
			trustedNamespaces = &val
		}
	}

	// Parse contributed filter
	var contributed *bool
	if cb := r.URL.Query().Get("contributed"); cb != "" {
		if val, err := strconv.ParseBool(cb); err == nil {
			contributed = &val
		}
	}

	// Parse target_terraform_version
	var targetTerraformVersion *string
	if ttv := r.URL.Query().Get("target_terraform_version"); ttv != "" {
		targetTerraformVersion = &ttv
	}

	// Execute search
	params := moduleQuery.SearchParams{
		Query:                  query,
		Namespaces:             namespaces,
		Providers:              providers,
		Verified:               verified,
		TrustedNamespaces:      trustedNamespaces,
		Contributed:            contributed,
		TargetTerraformVersion: targetTerraformVersion,
		Limit:                  limit,
		Offset:                 offset,
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
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))

	// Build input for query with namespace filter
	input := moduleQuery.ListModulesInput{
		Offset:       0,
		Limit:        0, // No limit for namespace modules
		Verified:     nil,
		IncludeCount: false,
	}

	// Execute query with namespace filter
	// Note: The current implementation doesn't support namespace filtering in ListModulesQuery
	// We need to use the repository search with namespace filter
	modules, _, err := h.listModulesQuery.Execute(ctx, input)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter by namespace in memory (temporary until proper query support)
	filteredModules := make([]*moduleModel.ModuleProvider, 0)
	for _, mp := range modules {
		if mp.Namespace().Name() == namespace {
			filteredModules = append(filteredModules, mp)
		}
	}

	// Convert to DTO
	response := h.presenter.ToListDTO(ctx, filteredModules)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderCreate handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/create
func (h *ModuleHandler) HandleModuleProviderCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

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
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))

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
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

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
// Returns detailed version information matching Python's ApiModuleVersions format
// Python reference: /app/test/unit/terrareg/server/test_api_module_versions.py
func (h *ModuleHandler) HandleModuleVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

	// Handle analytics token conversion (strip token from namespace if present)
	// Python converts "test_token-name__testnamespace" to "testnamespace"
	cleanNamespace := namespace
	namespaceStr := string(namespace)
	if strings.Contains(namespaceStr, "__") {
		parts := strings.SplitN(namespaceStr, "__", 2)
		if len(parts) == 2 {
			cleanNamespace = types.NamespaceName(parts[1])
		}
	}

	// Get the module provider first (use cleaned namespace for lookup)
	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, cleanNamespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Not Found")
		return
	}

	// Get versions from the module provider
	versions := moduleProvider.GetAllVersions()

	// Convert to version DTOs matching Python format:
	// {'root': {'dependencies': [], 'providers': []}, 'submodules': [], 'version': '1.2.0'}
	versionDTOs := make([]map[string]interface{}, len(versions))
	for i, version := range versions {
		// Get module specs for this version
		rootSpecs := version.GetRootModuleSpecs()

		// Build root with dependencies and providers arrays
		root := map[string]interface{}{
			"dependencies": []interface{}{},
			"providers":    []interface{}{},
		}

		// Add dependencies if present
		if len(rootSpecs.Dependencies) > 0 {
			deps := make([]map[string]interface{}, len(rootSpecs.Dependencies))
			for j, dep := range rootSpecs.Dependencies {
				deps[j] = map[string]interface{}{
					"module":  dep.Module,
					"source":  dep.Source,
					"version": dep.Version,
				}
			}
			root["dependencies"] = deps
		}

		// Add providers if present
		if len(rootSpecs.ProviderDependencies) > 0 {
			provDeps := make([]map[string]interface{}, len(rootSpecs.ProviderDependencies))
			for j, provDep := range rootSpecs.ProviderDependencies {
				provDeps[j] = map[string]interface{}{
					"name":      provDep.Name,
					"namespace": provDep.Namespace,
					"source":    provDep.Source,
					"version":   provDep.Version,
				}
			}
			root["providers"] = provDeps
		}

		// Build submodules array (always present, may be empty)
		submodules := make([]map[string]interface{}, 0)
		for _, subSpec := range version.GetSubmodules() {
			submoduleRoot := map[string]interface{}{
				"dependencies": []interface{}{},
				"providers":    []interface{}{},
			}

			// Add submodule dependencies if present
			if len(subSpec.Dependencies) > 0 {
				deps := make([]map[string]interface{}, len(subSpec.Dependencies))
				for j, dep := range subSpec.Dependencies {
					deps[j] = map[string]interface{}{
						"module":  dep.Module,
						"source":  dep.Source,
						"version": dep.Version,
					}
				}
				submoduleRoot["dependencies"] = deps
			}

			// Add submodule providers if present
			if len(subSpec.ProviderDependencies) > 0 {
				provDeps := make([]map[string]interface{}, len(subSpec.ProviderDependencies))
				for j, provDep := range subSpec.ProviderDependencies {
					provDeps[j] = map[string]interface{}{
						"name":      provDep.Name,
						"namespace": provDep.Namespace,
						"source":    provDep.Source,
						"version":   provDep.Version,
					}
				}
				submoduleRoot["providers"] = provDeps
			}

			submodules = append(submodules, map[string]interface{}{
				"path":       subSpec.Path,
				"submodules": []interface{}{},
			})
		}

		versionDTOs[i] = map[string]interface{}{
			"root":       root,
			"submodules": submodules,
			"version":    version.Version().String(),
		}
	}

	// Build response matching Python format:
	// {'modules': [{'source': 'ns/module/provider', 'versions': [...]}]}
	response := map[string]interface{}{
		"modules": []map[string]interface{}{
			{
				"source":   fmt.Sprintf("%s/%s/%s", cleanNamespace, name, provider),
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

	// Parse path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

	// Execute query
	moduleVersion, err := h.getModuleVersionQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	moduleProvider := moduleVersion.ModuleProvider()
	if moduleProvider == nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Convert to DTO - use the same detailed format as module provider endpoint
	response := h.versionPresenter.ToTerraregProviderDetailsDTO(ctx, moduleProvider, moduleVersion, string(namespace), string(name), string(provider), r.Host)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleDownload handles GET /v1/modules/{namespace}/{name}/{provider}/download
// and GET /v1/modules/{namespace}/{name}/{provider}/{version}/download
// Returns download location in Terraform Registry API format
func (h *ModuleHandler) HandleModuleDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version")) // May be empty for latest

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
		string(namespace), string(name), string(provider), downloadInfo.Version.Version().String())

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

	// Parse path parameters and convert to typed values
	namespaceStr := chi.URLParam(r, "namespace")
	nameStr := chi.URLParam(r, "name")
	providerStr := chi.URLParam(r, "provider")
	namespace := types.NamespaceName(namespaceStr)
	name := types.ModuleName(nameStr)
	provider := types.ModuleProviderName(providerStr)

	// Execute query
	settings, err := h.getModuleProviderSettingsQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build response
	response := moduledto.ModuleProviderSettingsResponse{
		Namespace:             namespaceStr,
		Module:                nameStr,
		Provider:              providerStr,
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

	// Parse path parameters and convert to typed values
	namespaceStr := chi.URLParam(r, "namespace")
	nameStr := chi.URLParam(r, "name")
	providerStr := chi.URLParam(r, "provider")
	namespace := types.NamespaceName(namespaceStr)
	name := types.ModuleName(nameStr)
	provider := types.ModuleProviderName(providerStr)

	// Parse request body
	var req moduledto.ModuleProviderSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	// Execute command
	cmdReq := moduleCmd.UpdateModuleProviderSettingsRequest{
		Namespace:             namespaceStr,
		Module:                nameStr,
		Provider:              providerStr,
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
		Namespace:             namespaceStr,
		Module:                nameStr,
		Provider:              providerStr,
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

	// Parse path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

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

	// Return 204 No Content (matching Python behavior)
	w.WriteHeader(http.StatusNoContent)
}

// HandleModuleVersionUpload handles POST /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/upload
func (h *ModuleHandler) HandleModuleVersionUpload(w http.ResponseWriter, r *http.Request) {
	// Use background context to avoid HTTP request timeout issues during long-running operations
	ctx := context.Background()

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

	// Execute unified process module command
	processReq := moduleCmd.ProcessModuleRequest{
		Namespace:    namespace,
		Module:       name,
		Provider:     provider,
		Version:      version,
		UploadSource: file,
		UploadSize:   header.Size,
		Options: moduleService.ProcessingOptions{
			SkipArchiveExtraction:   false,
			SkipTerraformProcessing: false,
			SkipMetadataProcessing:  false,
			SkipSecurityScanning:    false,
			SkipFileContentStorage:  false,
			SkipArchiveGeneration:   false,
			SecurityScanEnabled:     true,
			FileProcessingEnabled:   true,
			GenerateArchives:        true,
			ArchiveFormats: []moduleService.ArchiveFormat{
				moduleService.ArchiveFormatZIP,
				moduleService.ArchiveFormatTarGz,
			},
		},
	}

	if err := h.processModuleCmd.Execute(ctx, processReq); err != nil {
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

	// Execute unified process module command
	processReq := moduleCmd.ProcessModuleRequest{
		Namespace: namespace,
		Module:    name,
		Provider:  provider,
		Version:   *reqBody.Version,
		GitTag:    reqBody.GitTag,
		Options: moduleService.ProcessingOptions{
			SkipArchiveExtraction:   false,
			SkipTerraformProcessing: false,
			SkipMetadataProcessing:  false,
			SkipSecurityScanning:    false,
			SkipFileContentStorage:  false,
			SkipArchiveGeneration:   false,
			SecurityScanEnabled:     true,
			FileProcessingEnabled:   true,
			GenerateArchives:        true,
			ArchiveFormats: []moduleService.ArchiveFormat{
				moduleService.ArchiveFormatZIP,
				moduleService.ArchiveFormatTarGz,
			},
		},
	}

	if err := h.processModuleCmd.Execute(ctx, processReq); err != nil {
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

	// Get path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

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

	// Return submodules (match Python terrareg format - direct array)
	RespondJSON(w, http.StatusOK, submodules)
}

// HandleGetExamples handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/examples
func (h *ModuleHandler) HandleGetExamples(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

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

	// Return examples (match Python terrareg format - direct array)
	RespondJSON(w, http.StatusOK, examples)
}

// HandleTerraregModuleProviderDetails handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}
// Returns full terrareg API details for the latest version (published or unpublished)
func (h *ModuleHandler) HandleTerraregModuleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters and convert to typed values
	namespaceStr := chi.URLParam(r, "namespace")
	nameStr := chi.URLParam(r, "name")
	providerStr := chi.URLParam(r, "provider")
	namespace := types.NamespaceName(namespaceStr)
	name := types.ModuleName(nameStr)
	provider := types.ModuleProviderName(providerStr)

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

	// Convert to Terrareg Provider Details DTO with request domain and analytics token
	response := h.versionPresenter.ToTerraregProviderDetailsDTO(ctx, moduleProvider, moduleProvider.GetLatestVersion(), namespaceStr, nameStr, providerStr, requestDomain)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleTerraregModuleProviderVersions handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/versions
func (h *ModuleHandler) HandleTerraregModuleProviderVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

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
	var reqBody struct {
		GitTag *string `json:"git_tag,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		// For this deprecated endpoint, body parsing is optional
		// If no git_tag is provided, we'll try to derive from version
	}

	// Determine git tag - use provided git_tag or derive from version
	gitTag := reqBody.GitTag
	if gitTag == nil {
		// Derive git tag from version (common pattern: use version as git tag)
		gitTag = &version
	}

	// Execute unified process module command
	processReq := moduleCmd.ProcessModuleRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
		GitTag:    gitTag,
		Options: moduleService.ProcessingOptions{
			SkipArchiveExtraction:   false,
			SkipTerraformProcessing: false,
			SkipMetadataProcessing:  false,
			SkipSecurityScanning:    false,
			SkipFileContentStorage:  false,
			SkipArchiveGeneration:   false,
			SecurityScanEnabled:     true,
			FileProcessingEnabled:   true,
			GenerateArchives:        true,
			ArchiveFormats: []moduleService.ArchiveFormat{
				moduleService.ArchiveFormatZIP,
				moduleService.ArchiveFormatTarGz,
			},
		},
	}

	if err := h.processModuleCmd.Execute(ctx, processReq); err != nil {
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

	// Extract path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	moduleName := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

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

	// Send success response - 200 OK with status:Success (matching Python behavior)
	RespondJSON(w, http.StatusOK, map[string]string{"status": "Success"})
}

// HandleModuleVersionReadmeHTML handles GET /modules/{namespace}/{name}/{provider}/{version}/readme_html
func (h *ModuleHandler) HandleModuleVersionReadmeHTML(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	moduleName := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

	// Create request to get README HTML from module details
	req := &moduleQuery.GetReadmeHTMLRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
	}

	// Execute query to get README HTML
	resp, err := h.getReadmeHTMLQuery.Execute(ctx, req)
	if err != nil {
		if err.Error() == "no README content found" {
			// Return HTML error page for missing README
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="alert alert-warning">No README found for this module version</div>`))
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the processed HTML content
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.HTML))
}

// HandleModuleVersionSourceDownload handles GET /modules/{namespace}/{name}/{provider}/{version}/source.zip
func (h *ModuleHandler) HandleModuleVersionSourceDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters and convert to typed values
	namespaceStr := chi.URLParam(r, "namespace")
	moduleName := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))

	// Record analytics for the download (async, don't block the download)
	go func() {
		// Extract analytics information
		terraformVersion := r.Header.Get("X-Terraform-Version")
		if terraformVersion == "" {
			// Try to extract from User-Agent for older Terraform versions
			userAgent := r.Header.Get("User-Agent")
			if strings.Contains(userAgent, "Terraform/") {
				parts := strings.Split(userAgent, "/")
				if len(parts) > 1 {
					terraformVersion = parts[1]
				}
			}
		}

		// Extract analytics token from namespace if present (format: {token}__{namespace})
		var analyticsToken *string
		var cleanNamespaceStr string
		if strings.Contains(namespaceStr, "__") {
			parts := strings.SplitN(namespaceStr, "__", 2)
			if len(parts) == 2 {
				analyticsToken = &parts[0]
				cleanNamespaceStr = parts[1]
			}
		} else {
			cleanNamespaceStr = namespaceStr
		}
		cleanNamespace := types.NamespaceName(cleanNamespaceStr)

		// Get auth token from context if authenticated
		authUsername := ""
		if authCtx := middleware.GetAuthContext(ctx); authCtx.IsAuthenticated() {
			authUsername = authCtx.GetUsername()
		}

		// Create analytics recording request
		anaylticsReq := analyticsCmd.RecordModuleDownloadRequest{
			Namespace:        cleanNamespace,
			Module:           moduleName,
			Provider:         provider,
			Version:          version,
			TerraformVersion: &terraformVersion,
			AnalyticsToken:   analyticsToken,
			AuthToken:        &authUsername,
			Environment:      nil, // TODO: Extract environment from auth token if needed
		}

		// Record analytics (silently fail on error to not affect downloads)
		h.recordModuleDownloadCmd.Execute(ctx, anaylticsReq)
	}()

	// Create request (use original namespace with potential analytics token)
	namespace := types.NamespaceName(namespaceStr)
	req := &moduleCmd.GenerateModuleSourceRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
	}

	// Execute command with streaming
	resp, err := h.generateModuleSourceCmd.ExecuteAndStore(ctx, req)
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
	if resp.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", resp.Size))
	}

	// Stream the file directly from storage to HTTP response
	w.WriteHeader(http.StatusOK)
	err = h.generateModuleSourceCmd.StreamFromStorage(ctx, resp.StoragePath, w)
	if err != nil {
		// Log error but can't send error response as headers already sent
		return
	}
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

	// Get path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

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

// HandleModuleProviderRedirectsGet handles GET /v1/terrareg/modules/redirects
func (h *ModuleHandler) HandleModuleProviderRedirectsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute query to get all redirects
	redirects, err := h.getModuleProviderRedirectsQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO format
	response := make([]map[string]interface{}, len(redirects))
	for i, redirect := range redirects {
		response[i] = map[string]interface{}{
			"id":                 redirect.ID,
			"module_provider_id": redirect.ModuleProviderID,
			"namespace_id":       redirect.NamespaceID,
			"module":             redirect.Module,
			"provider":           redirect.Provider,
		}
	}

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderRedirectCreate handles PUT /v1/terrareg/modules/{namespace}/{name}/{provider}/redirect
func (h *ModuleHandler) HandleModuleProviderRedirectCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters and convert to typed values
	namespaceStr := chi.URLParam(r, "namespace")
	nameStr := chi.URLParam(r, "name")
	providerStr := chi.URLParam(r, "provider")
	namespace := types.NamespaceName(namespaceStr)
	name := types.ModuleName(nameStr)
	provider := types.ModuleProviderName(providerStr)

	// Parse request body
	var reqBody struct {
		ToModuleProviderID int `json:"to_module_provider_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	// Create command request
	cmdReq := moduleCmd.CreateModuleProviderRedirectRequest{
		FromNamespace:      namespace,
		FromModule:         name,
		FromProvider:       provider,
		ToModuleProviderID: reqBody.ToModuleProviderID,
	}

	// Execute command
	if err := h.createModuleProviderRedirectCmd.Execute(ctx, cmdReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Module provider redirect created successfully",
		"redirect": map[string]interface{}{
			"from_namespace":        namespaceStr,
			"from_module":           nameStr,
			"from_provider":         providerStr,
			"to_module_provider_id": reqBody.ToModuleProviderID,
		},
	})
}

// HandleModuleProviderRedirectDelete handles DELETE /v1/terrareg/modules/{namespace}/{name}/{provider}/redirects/{redirect_id}
func (h *ModuleHandler) HandleModuleProviderRedirectDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	name := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))

	// Create command request
	cmdReq := moduleCmd.DeleteModuleProviderRedirectRequest{
		Namespace: namespace,
		Module:    name,
		Provider:  provider,
	}

	// Execute command
	if err := h.deleteModuleProviderRedirectCmd.Execute(ctx, cmdReq); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 204 No Content on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions for parsing query parameters
// These functions handle parsing of optional query parameters matching Python behavior

// parseBoolPtr parses a string to a bool pointer
// Returns nil if the string is empty or invalid
func parseBoolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}
	return &b
}

// parseInt parses a string to an int with a default value
// Returns defaultValue if the string is empty or invalid
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return i
}
