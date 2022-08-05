"""Add git_path to module provider table

Revision ID: 51d3707d6334
Revises: 437ccafb9882
Create Date: 2022-07-22 07:57:02.500034

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '51d3707d6334'
down_revision = '437ccafb9882'
branch_labels = None
depends_on = None


def upgrade():
    # Add new column to module provider table for git path
    with op.batch_alter_table('module_provider', schema=None) as batch_op:
        batch_op.add_column(sa.Column('git_path', sa.String(length=1024), nullable=True))


def downgrade():
    # Remove newly added column
    with op.batch_alter_table('module_provider', schema=None) as batch_op:
       batch_op.drop_column('git_path')
