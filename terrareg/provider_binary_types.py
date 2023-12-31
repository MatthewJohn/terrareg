

from enum import Enum


class ProviderBinaryOperatingSystemType(Enum):
    """Provider binary operating systems"""

    FREEBSD = "freebsd"
    DARWIN = "darwin"
    WINDOWS = "windows"
    LINUX = "linux"


class ProviderBinaryArchitectureType(Enum):
    """Provider binary architectures"""

    AMD64 = "amd64"
    ARM = "arm"
    ARM64 = "arm64"
    I386 = "386"
