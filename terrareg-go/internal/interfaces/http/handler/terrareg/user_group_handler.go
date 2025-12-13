package terrareg

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	userGroupCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/user_group"
	userGroupQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/user_group"
)

// UserGroupHandler handles user group-related requests
// Matches Python APIs: ApiTerraregAuthUserGroups and ApiTerraregAuthUserGroup
type UserGroupHandler struct {
	listUserGroupsQuery *userGroupQuery.ListUserGroupsQuery
	createUserGroupCmd  *userGroupCmd.CreateUserGroupCommand
	deleteUserGroupCmd  *userGroupCmd.DeleteUserGroupCommand
}

// NewUserGroupHandler creates a new user group handler
func NewUserGroupHandler(
	listUserGroupsQuery *userGroupQuery.ListUserGroupsQuery,
	createUserGroupCmd *userGroupCmd.CreateUserGroupCommand,
	deleteUserGroupCmd *userGroupCmd.DeleteUserGroupCommand,
) *UserGroupHandler {
	return &UserGroupHandler{
		listUserGroupsQuery: listUserGroupsQuery,
		createUserGroupCmd:  createUserGroupCmd,
		deleteUserGroupCmd:  deleteUserGroupCmd,
	}
}

// HandleListUserGroups handles GET /v1/terrareg/auth/user-groups
// Matches Python ApiTerraregAuthUserGroups._get()
func (h *UserGroupHandler) HandleListUserGroups(w http.ResponseWriter, r *http.Request) {
	// Execute query
	groups, err := h.listUserGroupsQuery.Execute(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response matching Python
	RespondJSON(w, http.StatusOK, groups)
}

// HandleCreateUserGroup handles POST /v1/terrareg/auth/user-groups
// Matches Python ApiTerraregAuthUserGroups._post()
func (h *UserGroupHandler) HandleCreateUserGroup(w http.ResponseWriter, r *http.Request) {
	var req userGroupCmd.CreateUserGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate site_admin is boolean
	if req.SiteAdmin != true && req.SiteAdmin != false {
		RespondError(w, http.StatusBadRequest, "Invalid site_admin value")
		return
	}

	// Execute command
	response, err := h.createUserGroupCmd.Execute(r.Context(), &req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if response == nil {
		// Group already exists
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{})
		return
	}

	// Return response matching Python: 201 with name and site_admin
	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"name":       response.Name,
		"site_admin": response.SiteAdmin,
	})
}

// HandleDeleteUserGroup handles DELETE /v1/terrareg/auth/user-group/{group}
// Matches Python ApiTerraregAuthUserGroup._delete()
func (h *UserGroupHandler) HandleDeleteUserGroup(w http.ResponseWriter, r *http.Request) {
	groupName := chi.URLParam(r, "group")
	if groupName == "" {
		RespondError(w, http.StatusBadRequest, "Group name is required")
		return
	}

	// Execute command
	err := h.deleteUserGroupCmd.Execute(r.Context(), groupName)
	if err != nil {
		if err.Error() == "User group does not exist" {
			RespondJSON(w, http.StatusBadRequest, map[string]string{
				"message": "User group does not exist",
			})
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 200 like Python
	RespondJSON(w, http.StatusOK, map[string]interface{}{})
}