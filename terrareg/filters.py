
from enum import Enum

class NamespaceTrustFilter(Enum):
    """Enum to be information about trusted namespaces."""

    UNSPECIFIED = 0
    TRUSTED_NAMESPACES = 1
    CONTRIBUTED = 2
    ALL = 3

