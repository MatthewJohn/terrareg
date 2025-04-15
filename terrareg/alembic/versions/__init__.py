
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import ENUM

def Enum(*args, **kwargs):
    return sa.Enum(*args, **kwargs).with_variant(ENUM(*args, **kwargs), "postgresql")