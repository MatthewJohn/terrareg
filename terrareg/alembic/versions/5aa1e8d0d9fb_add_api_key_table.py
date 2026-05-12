"""add_api_key_table

Revision ID: 5aa1e8d0d9fb
Revises: c72f7c6ef6a7
Create Date: 2026-05-12 00:00:00.000000

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '5aa1e8d0d9fb'
down_revision = 'c72f7c6ef6a7'
branch_labels = None
depends_on = None


def upgrade():
    op.create_table(
        'api_key',
        sa.Column('id', sa.Integer(), nullable=False, autoincrement=True),
        sa.Column('name', sa.String(length=128), nullable=False),
        sa.Column('key_type', sa.String(length=32), nullable=False),
        sa.Column('key_prefix', sa.String(length=16), nullable=False),
        sa.Column('key_hash', sa.String(length=128), nullable=False),
        sa.Column('key_salt', sa.String(length=64), nullable=False),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('created_by', sa.String(length=128), nullable=True),
        sa.Column('last_used_at', sa.DateTime(), nullable=True),
        sa.Column('expires_at', sa.DateTime(), nullable=True),
        sa.Column('is_active', sa.Boolean(), nullable=False, server_default=sa.true()),
        sa.PrimaryKeyConstraint('id')
    )


def downgrade():
    op.drop_table('api_key')