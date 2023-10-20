
from typing import Union, List

import re


class RepositoryReleaseMetadata:
    """Structure for storing repository release metadata"""

    _TAG_REGEX = re.compile(r'^v([0-9]+\.[0-9]+\.[0-9]+(:?-[a-z0-9]+)?)$')

    def __init__(self, name: str, tag: str, archive_url: str, commit_hash: str,
                 provider_id: Union[str, int], release_artifacts: List['ReleaseArtifactMetadata']):
        """Store member variables"""
        self.name = name
        self.tag = tag
        self.provider_id = provider_id
        self.release_artifacts = release_artifacts
        self.archive_url = archive_url
        self.commit_hash = commit_hash

    @classmethod
    def tag_to_version(cls, tag: str) -> Union[None, str]:
        """Convert tag to version"""
        # Since we currently support the original tagging strategy 'v{version}', simply strip off the v
        if (match := cls._TAG_REGEX.match(tag)):
            return match.group(1)
        return None

    @property
    def version(self) -> Union[None, str]:
        """Convert tag to version"""
        return self.tag_to_version(self.tag)


class ReleaseArtifactMetadata:
    """Structure for storing release artifact metadata"""

    def __init__(self, name: str, provider_id: Union[str, int]):
        """Store member variables"""
        self.name = name
        self.provider_id = provider_id
