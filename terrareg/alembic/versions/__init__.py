
from typing import List

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import ENUM

def Enum(*args, **kwargs):
    return sa.Enum(*args, **kwargs).with_variant(ENUM(*args, **kwargs), "postgresql")


def update_enum(table: str, column: str, enum_name: str, old_enum_values: List[str], new_enum_values: List[str], nullable: bool=False):
    if op.get_bind().engine.name == 'sqlite':
        return
    elif op.get_bind().engine.name == 'mysql':
        op.alter_column(table, column,
            existing_type=Enum(
                *old_enum_values, name=enum_name),
            type_=Enum(
                *new_enum_values, name=enum_name),
            nullable=nullable)
    elif op.get_bind().engine.name == 'postgresql':
        new_type = sa.Enum(*new_enum_values, name=enum_name)
        temp_type_name = f"{enum_name}_migration_temp"
        op.execute(f"ALTER TYPE {enum_name} RENAME TO {temp_type_name}")

        new_type.create(op.get_bind())

        op.execute(
            f'ALTER TABLE {table} ALTER COLUMN {column} '
            f'TYPE {enum_name} USING {column}::text::{enum_name}'
        )
        op.execute(f'DROP TYPE {temp_type_name}')
    else:
        raise Exception("Invalid Database engine")

def constraint_exists(conn, table_name, constraint_name):
    dialect = conn.dialect.name

    if dialect == 'postgresql':
        result = conn.execute(sa.text("""
            SELECT 1
            FROM information_schema.table_constraints
            WHERE table_name = :table_name
              AND constraint_name = :constraint_name
        """), {'table_name': table_name, 'constraint_name': constraint_name})
        return result.scalar() is not None

    elif dialect == 'mysql':
        result = conn.execute(sa.text("""
            SELECT 1
            FROM information_schema.table_constraints
            WHERE table_name = :table_name
              AND constraint_name = :constraint_name
              AND constraint_schema = DATABASE()
        """), {'table_name': table_name, 'constraint_name': constraint_name})
        return result.scalar() is not None

    elif dialect == 'sqlite':
        # SQLite doesn’t support named foreign keys; you can’t drop them directly.
        # Best workaround: assume it doesn’t exist or refactor the table entirely.
        return False

    else:
        raise NotImplementedError(f"Unsupported DB dialect: {dialect}")