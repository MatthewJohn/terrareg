"""Add table for module version files

Revision ID: acd5e83c690f
Revises: 6416ffbf606d
Create Date: 2022-09-01 08:02:54.337726

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import mysql

# revision identifiers, used by Alembic.
revision = 'acd5e83c690f'
down_revision = '6416ffbf606d'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.create_table('module_version_file',
    sa.Column('id', sa.Integer(), nullable=False, autoincrement=True),
    sa.Column('module_version_id', sa.Integer(), nullable=False),
    sa.Column('path', sa.String(length=128), nullable=False),
    sa.Column('content', sa.LargeBinary(length=16777215).with_variant(mysql.MEDIUMBLOB(), 'mysql'), nullable=True),
    sa.ForeignKeyConstraint(['module_version_id'], ['module_version.id'], name='fk_module_version_file_module_version_id_module_version_id', onupdate='CASCADE', ondelete='CASCADE'),
    sa.PrimaryKeyConstraint('id')
    )
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_table('module_version_file')
    # ### end Alembic commands ###
