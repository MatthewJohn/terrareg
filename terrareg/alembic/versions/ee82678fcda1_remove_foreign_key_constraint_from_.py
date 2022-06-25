"""Remove foreign key constraint from analytics parent module version ID column

Revision ID: ee82678fcda1
Revises: eea5e27cac2b
Create Date: 2022-06-25 12:54:28.649027

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = 'ee82678fcda1'
down_revision = 'eea5e27cac2b'
branch_labels = None
depends_on = None


def upgrade():
    """Remove foreign key constraint from analytics parent_module_version column."""
    with op.batch_alter_table('analytics', schema=None) as batch_op:
        batch_op.drop_constraint('fk_analytics_parent_module_version_module_version_id', type_='foreignkey')


def downgrade():
    """Re-add foreign key constraint from analytics parent_module_version column."""
    with op.batch_alter_table('analytics', schema=None) as batch_op:
        batch_op.create_foreign_key('fk_analytics_parent_module_version_module_version_id', 'module_version', ['parent_module_version'], ['id'], onupdate='CASCADE')

