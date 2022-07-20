"""Add session table to hold active sessions

Revision ID: 437ccafb9882
Revises: d499042fad3b
Create Date: 2022-07-18 07:23:46.651210

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = '437ccafb9882'
down_revision = 'd499042fad3b'
branch_labels = None
depends_on = None


def upgrade():
    # Create session table
    op.create_table('session',
        sa.Column('id', sa.String(length=128), nullable=False),
        sa.Column('expiry', sa.DateTime(), nullable=False),
        sa.PrimaryKeyConstraint('id')
    )


def downgrade():
    # Remove session table
    op.drop_table('session')

