"""Update analytics to perform no action when parent module version is deleted

Revision ID: 0776813feaeb
Revises: ef71db86c2a1
Create Date: 2022-06-25 09:55:37.820127

"""
from alembic import op
import sqlalchemy as sa

from terrareg.alembic.versions import constraint_exists


# revision identifiers, used by Alembic.
revision = '0776813feaeb'
down_revision = 'ef71db86c2a1'
branch_labels = None
depends_on = None

naming_convention = {
    "ix": 'ix_%(column_0_label)s',
    "uq": "uq_%(table_name)s_%(column_0_name)s",
    "ck": "ck_%(table_name)s_%(column_0_name)s",
    "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
    "pk": "pk_%(table_name)s"
}

def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    with op.batch_alter_table('analytics', schema=None, naming_convention=naming_convention) as batch_op:
        if constraint_exists(op.get_bind(), 'analytics', 'analytics_ibfk_1'):
            # Use the default mysql name
            batch_op.drop_constraint('analytics_ibfk_1', type_='foreignkey')

    with op.batch_alter_table('analytics', schema=None) as batch_op:
        batch_op.create_foreign_key('fk_analytics_parent_module_version_module_version_id', 'module_version', ['parent_module_version'], ['id'], onupdate='CASCADE', ondelete='NO ACTION')
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    with op.batch_alter_table('analytics', schema=None, naming_convention=naming_convention) as batch_op:
        batch_op.drop_constraint('fk_analytics_parent_module_version_module_version_id', type_='foreignkey')
        # Use default mysql name when downgrading - this will still allow the SQLite upgrade
        # to happen, but without the catch
        batch_op.create_foreign_key('analytics_ibfk_1', 'module_version', ['parent_module_version'], ['id'], onupdate='CASCADE', ondelete='CASCADE')
    # ### end Alembic commands ###
