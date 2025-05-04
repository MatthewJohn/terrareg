"""Convert blobs to medium blobs

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

def MediumBlob():
    return sa.BLOB().with_variant(sqlalchemy.dialects.mysql.MEDIUMBLOB(), "mysql").with_variant(sa.LargeBinary(), "postgresql")

def upgrade():
    with op.batch_alter_table('module_version') as batch_op:
        # Convert blob columns to MEDIUMBLOB
        batch_op.alter_column('readme_content', existing_type=sa.BLOB, type_=MediumBlob())
        batch_op.alter_column('module_details', existing_type=sa.BLOB, type_=MediumBlob())
        batch_op.alter_column('variable_template', existing_type=sa.BLOB, type_=MediumBlob())
    with op.batch_alter_table('submodule') as batch_op:
        batch_op.alter_column('readme_content', existing_type=sa.BLOB, type_=MediumBlob())
        batch_op.alter_column('module_details', existing_type=sa.BLOB, type_=MediumBlob())

    # ### end Alembic commands ###


def downgrade():
    with op.batch_alter_table('module_version') as batch_op:
        batch_op.alter_column('readme_content', existing_type=MediumBlob(), type_=sa.BLOB)
        batch_op.alter_column('module_details', existing_type=MediumBlob(), type_=sa.BLOB)
        batch_op.alter_column('variable_template', existing_type=MediumBlob(), type_=sa.BLOB)

    with op.batch_alter_table('submodule') as batch_op:
        batch_op.alter_column('readme_content', existing_type=MediumBlob(), type_=sa.BLOB)
        batch_op.alter_column('module_details', existing_type=MediumBlob(), type_=sa.BLOB)
