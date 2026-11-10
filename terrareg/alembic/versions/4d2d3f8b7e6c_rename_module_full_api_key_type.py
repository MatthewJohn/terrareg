"""Rename module_full API key type to upload_and_publish

Revision ID: 4d2d3f8b7e6c
Revises: 5aa1e8d0d9fb
Create Date: 2026-05-24 00:00:00.000000

"""

from alembic import op


# revision identifiers, used by Alembic.
revision = '4d2d3f8b7e6c'
down_revision = '5aa1e8d0d9fb'
branch_labels = None
depends_on = None


def upgrade():
    """Rename stored API key type values."""
    op.execute(
        "UPDATE api_key SET key_type = 'upload_and_publish' WHERE key_type = 'module_full'"
    )


def downgrade():
    """Restore the previous API key type value."""
    op.execute(
        "UPDATE api_key SET key_type = 'module_full' WHERE key_type = 'upload_and_publish'"
    )