
from enum import Enum


class ModuleExtractionStatusType(Enum):
    """Statuses for module extraction"""

    IN_PROGRESS = "IN_PROGRESS"
    FAILED = "FAILED"
    COMPLETE = "COMPLETE"
