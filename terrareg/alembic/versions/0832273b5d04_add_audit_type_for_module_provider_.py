"""Add audit type for module provider redirect deletion

Revision ID: 0832273b5d04
Revises: 7b6cac0d522f
Create Date: 2023-08-20 07:31:52.029767

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '0832273b5d04'
down_revision = '7b6cac0d522f'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    if op.get_bind().engine.name == 'mysql':
        op.alter_column('audit_history', 'action',
            existing_type=sa.Enum(
                'NAMESPACE_CREATE', 'NAMESPACE_MODIFY_NAME', 'NAMESPACE_MODIFY_DISPLAY_NAME', 'MODULE_PROVIDER_CREATE', 'MODULE_PROVIDER_DELETE', 'MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT',
                'MODULE_PROVIDER_UPDATE_GIT_PROVIDER', 'MODULE_PROVIDER_UPDATE_GIT_PATH', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL',
                'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL', 'MODULE_PROVIDER_UPDATE_VERIFIED',
                'MODULE_VERSION_INDEX', 'MODULE_VERSION_PUBLISH', 'MODULE_VERSION_DELETE', 'USER_GROUP_CREATE', 'USER_GROUP_DELETE', 'USER_GROUP_NAMESPACE_PERMISSION_ADD',
                'USER_GROUP_NAMESPACE_PERMISSION_MODIFY', 'USER_GROUP_NAMESPACE_PERMISSION_DELETE', 'USER_LOGIN',
                'MODULE_PROVIDER_UPDATE_NAMESPACE', 'MODULE_PROVIDER_UPDATE_MODULE_NAME', 'MODULE_PROVIDER_UPDATE_PROVIDER_NAME', 'NAMESPACE_DELETE', name='auditaction'),
            type_=sa.Enum(
                'NAMESPACE_CREATE', 'NAMESPACE_MODIFY_NAME', 'NAMESPACE_MODIFY_DISPLAY_NAME', 'MODULE_PROVIDER_CREATE', 'MODULE_PROVIDER_DELETE', 'MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT',
                'MODULE_PROVIDER_UPDATE_GIT_PROVIDER', 'MODULE_PROVIDER_UPDATE_GIT_PATH', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL',
                'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL', 'MODULE_PROVIDER_UPDATE_VERIFIED',
                'MODULE_VERSION_INDEX', 'MODULE_VERSION_PUBLISH', 'MODULE_VERSION_DELETE', 'USER_GROUP_CREATE', 'USER_GROUP_DELETE', 'USER_GROUP_NAMESPACE_PERMISSION_ADD',
                'USER_GROUP_NAMESPACE_PERMISSION_MODIFY', 'USER_GROUP_NAMESPACE_PERMISSION_DELETE', 'USER_LOGIN',
                'MODULE_PROVIDER_UPDATE_NAMESPACE', 'MODULE_PROVIDER_UPDATE_MODULE_NAME', 'MODULE_PROVIDER_UPDATE_PROVIDER_NAME', 'NAMESPACE_DELETE', 'MODULE_PROVIDER_REDIRECT_DELETE', name='auditaction'),
            nullable=False)
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    if op.get_bind().engine.name == 'mysql':
        op.alter_column('audit_history', 'action',
            existing_type=sa.Enum(
                'NAMESPACE_CREATE', 'NAMESPACE_MODIFY_NAME', 'NAMESPACE_MODIFY_DISPLAY_NAME', 'MODULE_PROVIDER_CREATE', 'MODULE_PROVIDER_DELETE', 'MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT',
                'MODULE_PROVIDER_UPDATE_GIT_PROVIDER', 'MODULE_PROVIDER_UPDATE_GIT_PATH', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL',
                'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL', 'MODULE_PROVIDER_UPDATE_VERIFIED',
                'MODULE_VERSION_INDEX', 'MODULE_VERSION_PUBLISH', 'MODULE_VERSION_DELETE', 'USER_GROUP_CREATE', 'USER_GROUP_DELETE', 'USER_GROUP_NAMESPACE_PERMISSION_ADD',
                'USER_GROUP_NAMESPACE_PERMISSION_MODIFY', 'USER_GROUP_NAMESPACE_PERMISSION_DELETE', 'USER_LOGIN',
                'MODULE_PROVIDER_UPDATE_NAMESPACE', 'MODULE_PROVIDER_UPDATE_MODULE_NAME', 'MODULE_PROVIDER_UPDATE_PROVIDER_NAME', 'NAMESPACE_DELETE', 'MODULE_PROVIDER_REDIRECT_DELETE', name='auditaction'),
            type_=sa.Enum(
                'NAMESPACE_CREATE', 'NAMESPACE_MODIFY_NAME', 'NAMESPACE_MODIFY_DISPLAY_NAME', 'MODULE_PROVIDER_CREATE', 'MODULE_PROVIDER_DELETE', 'MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT',
                'MODULE_PROVIDER_UPDATE_GIT_PROVIDER', 'MODULE_PROVIDER_UPDATE_GIT_PATH', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL',
                'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL', 'MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL', 'MODULE_PROVIDER_UPDATE_VERIFIED',
                'MODULE_VERSION_INDEX', 'MODULE_VERSION_PUBLISH', 'MODULE_VERSION_DELETE', 'USER_GROUP_CREATE', 'USER_GROUP_DELETE', 'USER_GROUP_NAMESPACE_PERMISSION_ADD',
                'USER_GROUP_NAMESPACE_PERMISSION_MODIFY', 'USER_GROUP_NAMESPACE_PERMISSION_DELETE', 'USER_LOGIN',
                'MODULE_PROVIDER_UPDATE_NAMESPACE', 'MODULE_PROVIDER_UPDATE_MODULE_NAME', 'MODULE_PROVIDER_UPDATE_PROVIDER_NAME', 'NAMESPACE_DELETE', name='auditaction'),
            nullable=False)
    # ### end Alembic commands ###