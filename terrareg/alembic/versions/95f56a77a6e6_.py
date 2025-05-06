"""Add beta flag to module version

Revision ID: 95f56a77a6e6
Revises: 0e3de802e5e0
Create Date: 2022-05-21 07:49:04.047682

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '95f56a77a6e6'
down_revision = '0e3de802e5e0'
branch_labels = None
depends_on = None


def upgrade():
    # Add beta column, allowing nullable value
    op.add_column('module_version', sa.Column('beta', sa.BOOLEAN(), nullable=True))

    # Set any pre-existing rows to beta 0
    false_value = "0"
    if op.get_bind().engine.name == 'postgresql':
        false_value = "false"
    op.execute(f"""UPDATE module_version set beta={false_value}""")

    # Disable nullable flag in column
    with op.batch_alter_table('module_version', schema=None) as batch_op:
        batch_op.alter_column('beta', existing_type=sa.BOOLEAN(), nullable=False)


def downgrade():
    op.drop_column('module_version', 'beta')
