package user_group

import (
	"context"
	"fmt"
	"regexp"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// CreateUserGroupCommand handles creating a new user group
// Matches Python: UserGroup.create(name, site_admin)
type CreateUserGroupCommand struct {
	// userGroupRepo handles user group persistence (required)
	userGroupRepo         repository.UserGroupRepository
	userGroupAuditService *auditservice.UserGroupAuditService
}

// NewCreateUserGroupCommand creates a new create user group command
// Returns an error if userGroupRepo is nil
func NewCreateUserGroupCommand(userGroupRepo repository.UserGroupRepository, userGroupAuditService *auditservice.UserGroupAuditService) (*CreateUserGroupCommand, error) {
	if userGroupRepo == nil {
		return nil, fmt.Errorf("userGroupRepo cannot be nil")
	}
	return &CreateUserGroupCommand{
		userGroupRepo:         userGroupRepo,
		userGroupAuditService: userGroupAuditService,
	}, nil
}

// CreateUserGroupRequest represents the request to create a user group
// Matches Python JSON input: {name, site_admin}
type CreateUserGroupRequest struct {
	Name      string `json:"name"`
	SiteAdmin *bool  `json:"site_admin"`
}

// CreateUserGroupResponse represents the response after creating a user group
// Matches Python JSON response: {name, site_admin}
type CreateUserGroupResponse struct {
	Name      string `json:"name"`
	SiteAdmin bool   `json:"site_admin"`
}

// userGroupNameRegex matches the Python regex for user group names
// Python: ^[\s0-9a-zA-Z-_]+$
var userGroupNameRegex = regexp.MustCompile(`^[\s0-9a-zA-Z-_]+$`)

// Execute creates a new user group
// Matches Python: UserGroup.create(name, site_admin)
// Returns CreateUserGroupResponse on success, error on failure
func (c *CreateUserGroupCommand) Execute(ctx context.Context, req CreateUserGroupRequest) (*CreateUserGroupResponse, error) {
	// Validate user group name matches regex
	// Python: if not re.match(r'^[\s0-9a-zA-Z-_]+$', name):
	if !userGroupNameRegex.MatchString(req.Name) {
		return nil, ErrInvalidUserGroupName
	}

	// Validate site_admin is explicitly True or False (not None/undefined)
	// Python: if site_admin is None or not isinstance(site_admin, bool):
	if req.SiteAdmin == nil {
		return nil, ErrInvalidSiteAdminValue
	}

	// Check if user group already exists
	// Python: if UserGroup.get_by_group_name(name) is not None:
	existing, err := c.userGroupRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check user group existence: %w", err)
	}
	if existing != nil {
		return nil, ErrUserGroupAlreadyExists
	}

	// Create user group domain model
	userGroup := &auth.UserGroup{
		Name:      req.Name,
		SiteAdmin: *req.SiteAdmin,
	}

	// Persist to repository
	// Python: db.session.add(user_group); db.session.commit()
	if err := c.userGroupRepo.Save(ctx, userGroup); err != nil {
		return nil, fmt.Errorf("failed to save user group: %w", err)
	}

	// Log audit event (async, non-blocking)
	// Python reference: /app/terrareg/models.py:183 - AuditAction.USER_GROUP_CREATE
	go c.userGroupAuditService.LogUserGroupCreate(ctx, req.Name)

	// Return response matching Python format
	return &CreateUserGroupResponse{
		Name:      userGroup.Name,
		SiteAdmin: userGroup.SiteAdmin,
	}, nil
}
