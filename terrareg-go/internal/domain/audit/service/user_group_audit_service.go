package service

import (
	"context"

	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UserGroupAuditServiceInterface defines the interface for user group audit operations
// This allows for proper mocking in tests while keeping the implementation in UserGroupAuditService
type UserGroupAuditServiceInterface interface {
	LogUserGroupCreate(ctx context.Context, groupName string) error
	LogUserGroupDelete(ctx context.Context, groupName string) error
	LogUserGroupNamespacePermissionAdd(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error
	LogUserGroupNamespacePermissionModify(ctx context.Context, groupName string, namespace types.NamespaceName, oldPermission, newPermission string) error
	LogUserGroupNamespacePermissionDelete(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error
}

// UserGroupAuditService handles audit logging for user group operations
// It implements UserGroupAuditServiceInterface
// Python reference: /app/terrareg/models.py - user group audit event creation
type UserGroupAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// Ensure UserGroupAuditService implements the interface at compile time
var _ UserGroupAuditServiceInterface = (*UserGroupAuditService)(nil)

// NewUserGroupAuditService creates a new UserGroupAuditService
func NewUserGroupAuditService(auditRepo auditRepo.AuditHistoryRepository) *UserGroupAuditService {
	return &UserGroupAuditService{
		auditRepo: auditRepo,
	}
}

// LogUserGroupCreate logs user group creation audit event
// Python reference: /app/terrareg/models.py:183 - AuditAction.USER_GROUP_CREATE
func (s *UserGroupAuditService) LogUserGroupCreate(ctx context.Context, groupName string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserGroupCreate,
		"UserGroup",
		groupName,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogUserGroupDelete logs user group deletion audit event
// Python reference: /app/terrareg/models.py:256 - AuditAction.USER_GROUP_DELETE
func (s *UserGroupAuditService) LogUserGroupDelete(ctx context.Context, groupName string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserGroupDelete,
		"UserGroup",
		groupName,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogUserGroupNamespacePermissionAdd logs user group namespace permission addition audit event
// Python reference: /app/terrareg/models.py:385 - AuditAction.USER_GROUP_NAMESPACE_PERMISSION_ADD
func (s *UserGroupAuditService) LogUserGroupNamespacePermissionAdd(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error {
	username := getUsernameFromContext(ctx)
	objectID := string(namespace) + "/" + permissionType
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserGroupNamespacePermissionAdd,
		"UserGroupNamespacePermission",
		objectID,
		nil,
		&permissionType,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogUserGroupNamespacePermissionModify logs user group namespace permission modification audit event
// Python reference: /app/terrareg/models.py:427 - AuditAction.USER_GROUP_NAMESPACE_PERMISSION_MODIFY
func (s *UserGroupAuditService) LogUserGroupNamespacePermissionModify(ctx context.Context, groupName string, namespace types.NamespaceName, oldPermission, newPermission string) error {
	username := getUsernameFromContext(ctx)
	objectID := string(namespace) + "/" + newPermission
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserGroupNamespacePermissionModify,
		"UserGroupNamespacePermission",
		objectID,
		&oldPermission,
		&newPermission,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogUserGroupNamespacePermissionDelete logs user group namespace permission deletion audit event
// Python reference: /app/terrareg/models.py:463 - AuditAction.USER_GROUP_NAMESPACE_PERMISSION_DELETE
func (s *UserGroupAuditService) LogUserGroupNamespacePermissionDelete(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error {
	username := getUsernameFromContext(ctx)
	objectID := string(namespace) + "/" + permissionType
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserGroupNamespacePermissionDelete,
		"UserGroupNamespacePermission",
		objectID,
		&permissionType,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
