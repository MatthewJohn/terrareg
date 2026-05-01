"""add_provider_source_hierarchy_columns

Revision ID: c72f7c6ef6a7
Revises: 6dd8adb5e1e3
Create Date: 2026-05-01 14:11:01.319937

"""
from alembic import op
import sqlalchemy as sa
from terrareg.alembic.versions import update_enum


# revision identifiers, used by Alembic.
revision = 'c72f7c6ef6a7'
down_revision = '6dd8adb5e1e3'
branch_labels = None
depends_on = None


# Current audit values before adding new ones
old_audit_values = [
    'NAMESPACE_CREATE', 'NAMESPACE_MODIFY_NAME', 'NAMESPACE_MODIFY_DISPLAY_NAME', 'NAMESPACE_DELETE',
    'MODULE_PROVIDER_CREATE', 'MODULE_PROVIDER_DELETE',
    'MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT', 'MODULE_PROVIDER_UPDATE_GIT_PROVIDER',
    'MODULE_PROVIDER_UPDATE_GIT_PATH', 'MODULE_PROVIDER_UPDATE_ARCHIVE_GIT_PATH',
    'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL',
    'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL', 'MODULE_PROVIDER_UPDATE_VERIFIED',
    'MODULE_PROVIDER_UPDATE_NAMESPACE', 'MODULE_PROVIDER_UPDATE_MODULE_NAME', 'MODULE_PROVIDER_UPDATE_PROVIDER_NAME',
    'MODULE_PROVIDER_REDIRECT_DELETE',
    'MODULE_VERSION_INDEX', 'MODULE_VERSION_PUBLISH', 'MODULE_VERSION_DELETE',
    'USER_GROUP_CREATE', 'USER_GROUP_DELETE', 'USER_GROUP_NAMESPACE_PERMISSION_ADD',
    'USER_GROUP_NAMESPACE_PERMISSION_MODIFY', 'USER_GROUP_NAMESPACE_PERMISSION_DELETE',
    'USER_LOGIN',
    'GPG_KEY_CREATE', 'GPG_KEY_DELETE',
    'PROVIDER_CREATE', 'PROVIDER_DELETE', 'PROVIDER_VERSION_INDEX', 'PROVIDER_VERSION_DELETE',
    'REPOSITORY_CREATE', 'REPOSITORY_UPDATE', 'REPOSITORY_DELETE'
]

new_audit_values = [
    'NAMESPACE_MODIFY_DEFAULT_PROVIDER_SOURCE',
    'MODULE_PROVIDER_UPDATE_PROVIDER_SOURCE',
    'MODULE_PROVIDER_UPDATE_PROVIDER_SOURCE_INHERITANCE_DISABLED'
]


def upgrade():
    # Add provider_source_name column to module_provider
    op.add_column('module_provider', sa.Column('provider_source_name', sa.String(length=128), nullable=True))

    # Add provider_source_inheritance_disabled column to module_provider
    op.add_column('module_provider', sa.Column('provider_source_inheritance_disabled', sa.Boolean(), nullable=False, server_default='false'))

    # Add foreign key for provider_source_name (PostgreSQL only, SQLite handles in batch)
    if op.get_bind().engine.name == "postgresql":
        op.create_foreign_key(
            'fk_module_provider_provider_source_name_provider_source_name',
            'module_provider', 'provider_source',
            ['provider_source_name'], ['name'],
            onupdate='CASCADE', ondelete='SET NULL'
        )

    # Add default_provider_source_name column to namespace
    op.add_column('namespace', sa.Column('default_provider_source_name', sa.String(length=255), nullable=True))

    # Add foreign key for default_provider_source_name (PostgreSQL only, SQLite handles in batch)
    if op.get_bind().engine.name == "postgresql":
        op.create_foreign_key(
            'fk_namespace_default_provider_source_name_provider_source_name',
            'namespace', 'provider_source',
            ['default_provider_source_name'], ['name'],
            onupdate='CASCADE', ondelete='SET NULL'
        )

    # Add new audit action enum values
    update_enum(
        'audit_history', 'action', 'auditaction',
        old_audit_values,
        old_audit_values + new_audit_values
    )


def downgrade():
    # Remove new audit action enum values
    update_enum(
        'audit_history', 'action', 'auditaction',
        old_audit_values + new_audit_values,
        old_audit_values
    )

    # Drop foreign keys (PostgreSQL only)
    if op.get_bind().engine.name == "postgresql":
        op.drop_constraint('fk_namespace_default_provider_source_name_provider_source_name', 'namespace', type_='foreignkey')
        op.drop_constraint('fk_module_provider_provider_source_name_provider_source_name', 'module_provider', type_='foreignkey')

    # Drop columns
    op.drop_column('namespace', 'default_provider_source_name')
    op.drop_column('module_provider', 'provider_source_inheritance_disabled')
    op.drop_column('module_provider', 'provider_source_name')
