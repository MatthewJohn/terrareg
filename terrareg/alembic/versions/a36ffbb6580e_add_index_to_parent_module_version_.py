"""Add index to parent_module_version column of analytics table

Revision ID: a36ffbb6580e
Revises: ee82678fcda1
Create Date: 2022-06-25 13:03:41.700763

"""
from alembic import op
import sqlalchemy as sa

from terrareg.alembic.versions import constraint_exists


# revision identifiers, used by Alembic.
revision = 'a36ffbb6580e'
down_revision = 'ee82678fcda1'
branch_labels = None
depends_on = None


def upgrade():
    with op.batch_alter_table('analytics', schema=None) as batch_op:
        if constraint_exists(op.get_bind(), 'analytics', 'fk_analytics_parent_module_version_module_version_id'):
            batch_op.drop_index('fk_analytics_parent_module_version_module_version_id')
        batch_op.create_index(op.f('ix_analytics_parent_module_version'), ['parent_module_version'], unique=False)

def downgrade():
    with op.batch_alter_table('analytics', schema=None) as batch_op:
        if constraint_exists(op.get_bind(), 'analytics', 'ix_analytics_parent_module_version'):
            batch_op.drop_index(op.f('ix_analytics_parent_module_version'))
        batch_op.create_index('fk_analytics_parent_module_version_module_version_id', ['parent_module_version'], unique=False)
