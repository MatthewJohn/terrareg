
from enum import Enum


class ApiKeyType(Enum):
    """Supported API key types."""

    ADMIN = 'admin'
    UPLOAD = 'upload'
    PUBLISH = 'publish'
    UPLOAD_AND_PUBLISH = 'upload_and_publish'
