"""Add tables for Terraform IDP sessions

Revision ID: fb6a94791a14
Revises: 0832273b5d04
Create Date: 2023-09-15 07:27:59.107986

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import mysql

# revision identifiers, used by Alembic.
revision = 'fb6a94791a14'
down_revision = '0832273b5d04'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.create_table('terraform_idp_access_token',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('key', sa.String(length=128), nullable=False),
    sa.Column('data', sa.LargeBinary(length=16777215).with_variant(mysql.MEDIUMBLOB(), 'mysql'), nullable=True),
    sa.Column('expiry', sa.DateTime(), nullable=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('key')
    )
    op.create_table('terraform_idp_authorization_code',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('key', sa.String(length=128), nullable=False),
    sa.Column('data', sa.LargeBinary(length=16777215).with_variant(mysql.MEDIUMBLOB(), 'mysql'), nullable=True),
    sa.Column('expiry', sa.DateTime(), nullable=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('key')
    )
    op.create_table('terraform_idp_subject_identifier',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('key', sa.String(length=128), nullable=False),
    sa.Column('data', sa.LargeBinary(length=16777215).with_variant(mysql.MEDIUMBLOB(), 'mysql'), nullable=True),
    sa.Column('expiry', sa.DateTime(), nullable=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('key')
    )
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_table('terraform_idp_subject_identifier')
    op.drop_table('terraform_idp_authorization_code')
    op.drop_table('terraform_idp_access_token')
    # ### end Alembic commands ###
