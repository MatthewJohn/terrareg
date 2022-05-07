"""empty message

Revision ID: 0e3de802e5e0
Revises: aef5947a7e1d
Create Date: 2022-05-07 09:08:33.674549

"""
from alembic import op
import sqlalchemy as sa
import sqlalchemy.dialects.mysql


# revision identifiers, used by Alembic.
revision = '0e3de802e5e0'
down_revision = 'aef5947a7e1d'
branch_labels = None
depends_on = None


def upgrade():
    print('HADGHADGIAD')
    def MediumBlob():
        return sa.LargeBinary(length=((2 ** 24) - 1)).with_variant(sqlalchemy.dialects.mysql.MEDIUMBLOB(), "mysql")

    # Convert blob columns to MEDIUMBLOB
    op.alter_column('module_version', 'readme_content', type=MediumBlob())
    op.alter_column('module_version', 'module_details', type=MediumBlob())
    op.alter_column('module_version', 'variable_template', type=MediumBlob())

    op.alter_column('submodule', 'readme_content', type=MediumBlob())
    op.alter_column('submodule', 'module_details', type=MediumBlob())

    # ### end Alembic commands ###


def downgrade():
    # Convert columns back to BLOB
    op.alter_column('module_version', 'readme_content', type=sa.BLOB)
    op.alter_column('module_version', 'module_details', type=sa.BLOB)
    op.alter_column('module_version', 'variable_template', type=sa.BLOB)

    op.alter_column('submodule', 'readme_content', type=sa.BLOB)
    op.alter_column('submodule', 'module_details', type=sa.BLOB)
