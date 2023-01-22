"""Add module version column for extraction version

Revision ID: 71d311c23025
Revises: 8b594ed19f9d
Create Date: 2023-01-22 09:18:23.880577

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '71d311c23025'
down_revision = '8b594ed19f9d'
branch_labels = None
depends_on = None


def upgrade():
    with op.batch_alter_table("module_version") as module_version_batch:
       module_version_batch.add_column(sa.Column('extraction_version', sa.Integer(), nullable=True))

    # Update column with default value
    c = op.get_bind()
    c.execute(f"""UPDATE module_version SET extraction_version=1""")


def downgrade():
    with op.batch_alter_table("module_version") as module_version_batch:
        module_version_batch.drop_column('extraction_version')
