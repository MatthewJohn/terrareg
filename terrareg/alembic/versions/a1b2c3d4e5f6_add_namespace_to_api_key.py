"""add_namespace_to_api_key

Revision ID: a1b2c3d4e5f6
Revises: 5aa1e8d0d9fb
Create Date: 2026-05-12 00:00:00.000000

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = 'a1b2c3d4e5f6'
down_revision = '5aa1e8d0d9fb'
branch_labels = None
depends_on = None


def upgrade():
    op.add_column('api_key', sa.Column('namespace', sa.String(length=128), nullable=True))


def downgrade():
    op.drop_column('api_key', 'namespace')
