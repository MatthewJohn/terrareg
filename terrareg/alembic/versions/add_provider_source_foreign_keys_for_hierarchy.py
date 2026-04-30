"""Add provider source foreign keys for hierarchy

Revision ID: 4a6b242082db
Revises: 71d311c23025
Create Date: 2026-04-30 00:00:00.000000

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '4a6b242082db'
down_revision = '71d311c23025'
branch_labels = None
depends_on = None


def upgrade():
    # Add provider_source_name column to module_provider table
    with op.batch_alter_table('module_provider', schema=None) as batch_op:
        batch_op.add_column(sa.Column('provider_source_name', sa.String(length=255), nullable=True))

    # Add default_provider_source_name column to namespace table
    with op.batch_alter_table('namespace', schema=None) as batch_op:
        batch_op.add_column(sa.Column('default_provider_source_name', sa.String(length=255), nullable=True))

    # Create foreign keys for PostgreSQL (SQLite handles them in batch mode)
    if op.get_bind().engine.name == "postgresql":
        op.create_foreign_key(
            'fk_module_provider_provider_source_name_provider_source_name',
            'module_provider', 'provider_source',
            ['provider_source_name'], ['name'],
            onupdate='CASCADE', ondelete='SET NULL'
        )
        op.create_foreign_key(
            'fk_namespace_default_provider_source_name_provider_source_name',
            'namespace', 'provider_source',
            ['default_provider_source_name'], ['name'],
            onupdate='CASCADE', ondelete='SET NULL'
        )


def downgrade():
    # Drop foreign keys for PostgreSQL
    if op.get_bind().engine.name == "postgresql":
        op.drop_constraint('fk_namespace_default_provider_source_name_provider_source_name', 'namespace', type_='foreignkey')
        op.drop_constraint('fk_module_provider_provider_source_name_provider_source_name', 'module_provider', type_='foreignkey')

    # Remove columns
    with op.batch_alter_table('namespace', schema=None) as batch_op:
        batch_op.drop_column('default_provider_source_name')

    with op.batch_alter_table('module_provider', schema=None) as batch_op:
        batch_op.drop_column('provider_source_name')
