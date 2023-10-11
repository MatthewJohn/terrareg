"""Add names to all foreign key constaints

Revision ID: eea5e27cac2b
Revises: 0776813feaeb
Create Date: 2022-06-25 11:55:38.389322

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = 'eea5e27cac2b'
down_revision = '0776813feaeb'
branch_labels = None
depends_on = None

foreign_keys = [
    # op.drop_constraint('example_file_ibfk_1', 'example_file', type_='foreignkey')
    # op.create_foreign_key('fk_example_file_submodule_id_submodule_id', 'example_file', 'submodule', ['submodule_id'], ['id'], onupdate='CASCADE', ondelete='CASCADE')
    {
        'old_name': 'example_file_ibfk_1', 'new_name': 'fk_example_file_submodule_id_submodule_id',
        'table': 'example_file', 'columns': ['submodule_id'],
        'other_table': 'submodule', 'other_columns': ['id'],
        'onupdate': 'CASCADE', 'ondelete': 'CASCADE'
    },
    # op.drop_constraint('latest_version_id', 'module_provider', type_='foreignkey')
    # op.create_foreign_key('fk_module_provider_latest_version_id_module_version_id', 'module_provider', 'module_version', ['latest_version_id'], ['id'], onupdate='CASCADE', ondelete='SET NULL', use_alter=True)
    {
        'old_name': 'latest_version_id', 'new_name': 'fk_module_provider_latest_version_id_module_version_id',
        'table': 'module_provider', 'columns': ['latest_version_id'],
        'other_table': 'module_version', 'other_columns': ['id'],
        'onupdate': 'CASCADE', 'ondelete': 'SET NULL'
    },
    # op.drop_constraint('module_provider_ibfk_1', 'module_provider', type_='foreignkey')
    # op.create_foreign_key('fk_module_provider_git_provider_id_git_provider_id', 'module_provider', 'git_provider', ['git_provider_id'], ['id'], onupdate='CASCADE', ondelete='SET NULL')
    {
        'old_name': 'module_provider_ibfk_1', 'new_name': 'fk_module_provider_git_provider_id_git_provider_id',
        'table': 'module_provider', 'columns': ['git_provider_id'],
        'other_table': 'git_provider', 'other_columns': ['id'],
        'onupdate': 'CASCADE', 'ondelete': 'SET NULL'
    },
    # op.drop_constraint('module_version_ibfk_1', 'module_version', type_='foreignkey')
    # op.create_foreign_key('fk_module_version_module_provider_id_module_provider_id', 'module_version', 'module_provider', ['module_provider_id'], ['id'], onupdate='CASCADE', ondelete='CASCADE')
    {
        'old_name': 'module_version_ibfk_1', 'new_name': 'fk_module_version_module_provider_id_module_provider_id',
        'table': 'module_version', 'columns': ['module_provider_id'],
        'other_table': 'module_provider', 'other_columns': ['id'],
        'onupdate': 'CASCADE', 'ondelete': 'CASCADE'
    },
    # op.drop_constraint('submodule_ibfk_1', 'submodule', type_='foreignkey')
    # op.create_foreign_key('fk_submodule_parent_module_version_module_version_id', 'submodule', 'module_version', ['parent_module_version'], ['id'], onupdate='CASCADE', ondelete='CASCADE')
    {
        'old_name': 'submodule_ibfk_1', 'new_name': 'fk_submodule_parent_module_version_module_version_id',
        'table': 'submodule', 'columns': ['parent_module_version'],
        'other_table': 'module_version', 'other_columns': ['id'],
        'onupdate': 'CASCADE', 'ondelete': 'CASCADE'
    }
]


def upgrade():
    for foreign_key in foreign_keys:
        try:
            with op.batch_alter_table(foreign_key['table'], schema=None) as batch_op:
                # Use the default mysql name
                batch_op.drop_constraint(foreign_key['old_name'], type_='foreignkey')
                batch_op.create_foreign_key(
                    foreign_key['new_name'], foreign_key['other_table'],
                    foreign_key['columns'], foreign_key['other_columns'],
                    onupdate=foreign_key['onupdate'], ondelete=foreign_key['ondelete']
                )
        except ValueError:
            # Catch "No such constraint: 'new_name'" error from SQLite
            with op.batch_alter_table('analytics', schema=None) as batch_op:
                batch_op.create_foreign_key(
                    foreign_key['new_name'], foreign_key['other_table'],
                    foreign_key['columns'], foreign_key['other_columns'],
                    onupdate=foreign_key['onupdate'], ondelete=foreign_key['ondelete']
                )


def downgrade():
    for foreign_key in foreign_keys:
        with op.batch_alter_table('analytics', schema=None) as batch_op:
            batch_op.drop_constraint(foreign_key['new_name'], type_='foreignkey')
            # Use default mysql name when downgrading - this will still allow the SQLite upgrade
            # to happen, but without the catch
            batch_op.create_foreign_key(
                foreign_key['old_name'], foreign_key['other_table'],
                foreign_key['columns'], foreign_key['other_columns'],
                onupdate=foreign_key['onupdate'], ondelete=foreign_key['ondelete']
            )
