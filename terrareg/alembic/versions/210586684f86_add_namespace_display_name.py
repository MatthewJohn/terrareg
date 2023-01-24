"""Add namespace display_name

Revision ID: 210586684f86
Revises: 71d311c23025
Create Date: 2023-01-24 07:07:45.170022

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '210586684f86'
down_revision = '71d311c23025'
branch_labels = None
depends_on = None


def upgrade():
    with op.batch_alter_table("namespace") as namespace_batch:
        namespace_batch.add_column(sa.Column("display_name", sa.String(length=128), nullable=True))


def downgrade():
    with op.batch_alter_table("namespace") as namespace_batch:
        namespace_batch.drop_column("display_name")
